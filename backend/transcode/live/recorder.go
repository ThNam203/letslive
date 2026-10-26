package live

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/storage"
)

const (
	// recorderQueueSize bounds the items buffered per rendition; beyond it
	// writes are dropped so the live path never waits on disk or MinIO.
	recorderQueueSize = 1024
	// recorderUploadQueue bounds the finished segments waiting for upload.
	recorderUploadQueue = 64

	recorderFileBuffer = 64 << 10

	// recorderDropLogInterval is the minimum time between two "input dropped"
	// log lines of one rendition.
	recorderDropLogInterval = 5 * time.Second

	// recorderMaxHeldAudio bounds the audio waiting for video to catch up
	// (about 6s of AAC); beyond it the oldest is written to the open segment.
	recorderMaxHeldAudio = 256
)

// Timeouts are variables only so tests can shorten them.
var (
	// recorderUploadTimeout bounds each AddSegment call. The call's context is
	// detached from the caller's, so uploads still finish when the caller's
	// context ends together with the stream.
	recorderUploadTimeout = 30 * time.Second
	// recorderFinishTimeout bounds how long Finish waits for the writers and
	// uploads; what has not finished by then is abandoned and logged.
	recorderFinishTimeout = 60 * time.Second
)

// dropLimiter turns a stream of drop events into at most one report per
// interval, each carrying the number of drops since the previous report.
type dropLimiter struct {
	interval time.Duration

	mu      sync.Mutex
	last    time.Time
	pending int64
}

// record counts one drop at now. When ok is true the caller should log n,
// the number of drops since the last report.
func (l *dropLimiter) record(now time.Time) (n int64, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pending++
	if !l.last.IsZero() && now.Sub(l.last) < l.interval {
		return 0, false
	}
	l.last = now
	n, l.pending = l.pending, 0
	return n, true
}

// flush returns, and forgets, the drops not yet reported.
func (l *dropLimiter) flush() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := l.pending
	l.pending = 0
	return n
}

// recItem is one queued write for a rendition's goroutine.
type recItem struct {
	isAudio bool
	pts     int64
	dts     int64
	au      [][]byte // video NAL units, or AAC access units
}

// recSegment is one MPEG-TS file. startDTS and path are fixed when it is
// opened; the rest is guarded by Recorder.resMu once the segment is listed.
type recSegment struct {
	index    int
	startDTS int64
	endDTS   int64 // DTS of the next segment's first AU (see Finish for the last)
	path     string

	settled bool // upload finished, failed or was discarded
	ok      bool // uploaded
	url     string
	size    int64
}

type recUpload struct {
	rendition int
	seg       *recSegment
}

type recRendition struct {
	index int
	name  string
	in    chan recItem
	done  chan struct{}
	drops dropLimiter

	// Guarded by Recorder.resMu.
	segments      []*recSegment
	width, height int
}

// Recorder records every rendition to MPEG-TS segment files, uploads them in
// the background and, at Finish, writes the VOD playlists. Nothing it does
// blocks WriteVideo or WriteAudio.
type Recorder struct {
	ctx      context.Context
	id       string
	audio    *mpeg4audio.AudioSpecificConfig
	st       storage.Storage
	dir      string
	segTicks int64

	rends   []*recRendition
	uploads chan recUpload
	upDone  chan struct{}
	abandon chan struct{} // closed when Finish gives up waiting

	dropped    atomic.Int64
	lateAudio  atomic.Int64
	badInput   dropLimiter
	badInputMu sync.Mutex

	// resMu guards rendition segment lists, sizes, and segment results.
	resMu sync.Mutex

	mu       sync.RWMutex // guards finished against sends on the rendition channels
	finished bool
}

// NewRecorder starts one writer goroutine per rendition and one uploader.
// Segment files go under {dir}/{id}/{rendition}/{n}.ts and are cut at the
// first random-access AU once at least segDur has passed. audio is nil for a
// video-only source. Finish must be called to stop the goroutines.
func NewRecorder(ctx context.Context, id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig, st storage.Storage, dir string, segDur time.Duration) *Recorder {
	return newRecorder(ctx, id, rs, audio, st, dir, segDur, recorderQueueSize, recorderUploadQueue)
}

func newRecorder(ctx context.Context, id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig, st storage.Storage, dir string, segDur time.Duration, itemQueue, uploadQueue int) *Recorder {
	r := &Recorder{
		ctx:      ctx,
		id:       id,
		audio:    audio,
		st:       st,
		dir:      dir,
		segTicks: ticks(segDur, videoClockRate),
		rends:    make([]*recRendition, len(rs)),
		uploads:  make(chan recUpload, uploadQueue),
		upDone:   make(chan struct{}),
		abandon:  make(chan struct{}),
		badInput: dropLimiter{interval: recorderDropLogInterval},
	}

	for i, rend := range rs {
		rr := &recRendition{
			index: i,
			name:  rend.Name,
			in:    make(chan recItem, itemQueue),
			done:  make(chan struct{}),
			drops: dropLimiter{interval: recorderDropLogInterval},
		}
		r.rends[i] = rr
		go func() {
			defer close(rr.done)
			(&segmentWriter{rec: r, rend: rr}).run()
		}()
	}
	go func() {
		defer close(r.upDone)
		r.uploadLoop()
	}()

	return r
}

// WriteVideo queues one H.264 access unit for rendition r. The NAL units must
// not be modified afterwards. Empty NAL units are discarded. It never blocks:
// a full queue drops the item.
func (r *Recorder) WriteVideo(rendition int, pts90k, dts90k int64, au [][]byte) {
	au = r.withoutEmpty(au)
	if len(au) == 0 {
		return
	}
	r.enqueue(rendition, recItem{pts: pts90k, dts: dts90k, au: au})
}

// WriteAudio queues AAC access units, stamped on the 90 kHz timeline, for
// every rendition so each VOD variant is self-contained. The slices must not
// be modified afterwards; empty ones are discarded. It never blocks.
func (r *Recorder) WriteAudio(pts90k int64, aus [][]byte) {
	if r.audio == nil {
		return
	}
	aus = r.withoutEmpty(aus)
	if len(aus) == 0 {
		return
	}
	for i := range r.rends {
		r.enqueue(i, recItem{isAudio: true, pts: pts90k, au: aus})
	}
}

// Dropped returns how many items were discarded because a queue was full.
func (r *Recorder) Dropped() int64 { return r.dropped.Load() }

// withoutEmpty returns units without its zero-length entries (the mpegts and
// h264 helpers index their first byte). It allocates only when it removes one.
func (r *Recorder) withoutEmpty(units [][]byte) [][]byte {
	empty := 0
	for _, u := range units {
		if len(u) == 0 {
			empty++
		}
	}
	if empty == 0 {
		return units
	}

	r.badInputMu.Lock()
	n, ok := r.badInput.record(time.Now())
	r.badInputMu.Unlock()
	if ok {
		logger.Errorf(r.ctx, "livestream %s: recorder ignored empty NAL/access units (%d report(s) since the last log)", r.id, n)
	}

	kept := make([][]byte, 0, len(units)-empty)
	for _, u := range units {
		if len(u) > 0 {
			kept = append(kept, u)
		}
	}
	return kept
}

func (r *Recorder) enqueue(i int, it recItem) {
	if i < 0 || i >= len(r.rends) {
		logger.Errorf(r.ctx, "livestream %s: recorder rendition index %d out of range [0, %d)", r.id, i, len(r.rends))
		return
	}
	rr := r.rends[i]

	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.finished {
		return
	}

	select {
	case rr.in <- it:
	default:
		r.dropped.Add(1)
		if n, ok := rr.drops.record(time.Now()); ok {
			logger.Errorf(r.ctx, "livestream %s rendition %s: recorder queue full, dropped %d item(s) since the last report", r.id, rr.name, n)
		}
	}
}

// settle records the outcome of seg's upload (or its discarding).
func (r *Recorder) settle(seg *recSegment, ok bool, url string, size int64) {
	r.resMu.Lock()
	defer r.resMu.Unlock()
	seg.settled, seg.ok, seg.url, seg.size = true, ok, url, size
}

func (r *Recorder) uploadLoop() {
	for {
		select {
		case <-r.abandon:
			return
		default:
		}

		select {
		case <-r.abandon:
			return
		case job, ok := <-r.uploads:
			if !ok {
				return
			}
			r.upload(job)
		}
	}
}

func (r *Recorder) upload(job recUpload) {
	seg := job.seg
	info, err := os.Stat(seg.path)
	if err != nil {
		logger.Errorf(r.ctx, "livestream %s rendition %d: stat segment %s: %v", r.id, job.rendition, seg.path, err)
		r.settle(seg, false, "", 0)
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.ctx), recorderUploadTimeout)
	url, err := r.st.AddSegment(ctx, seg.path, r.id, job.rendition)
	cancel()
	if err != nil {
		logger.Errorf(r.ctx, "livestream %s rendition %d: upload segment %d: %v", r.id, job.rendition, seg.index, err)
		r.settle(seg, false, "", 0)
		return
	}

	r.settle(seg, true, url, info.Size())
	if err := os.Remove(seg.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		logger.Errorf(r.ctx, "livestream %s rendition %d: remove uploaded segment %s: %v", r.id, job.rendition, seg.path, err)
	}
}

// segmentWriter is the state owned by one rendition's goroutine.
type segmentWriter struct {
	rec  *Recorder
	rend *recRendition

	cur  *recSegment
	file *os.File
	buf  *bufio.Writer
	ts   *mpegts.Writer
	vid  *mpegts.Track
	aud  *mpegts.Track

	nextIndex int
	prevDTS   int64
	lastDTS   int64
	havePrev  bool
	frameDur  int64

	// held is audio newer than the last video AU written. It waits for video
	// to reach its timestamp so a segment cut puts it on the right side.
	held      []recItem
	lastAudio int64
	haveAudio bool

	haveSize      bool
	spsFailLogged bool
}

func (w *segmentWriter) run() {
	for it := range w.rend.in {
		select {
		case <-w.rec.abandon:
			continue // Finish gave up; just drain
		default:
		}
		if it.isAudio {
			w.writeAudio(it)
		} else {
			w.writeVideo(it)
		}
	}
	w.flushHeld(0, true)
	w.closeSegment(w.lastDTS + w.frameDur)
}

func (w *segmentWriter) logf(format string, args ...any) {
	logger.Errorf(w.rec.ctx, "livestream %s rendition %s: "+format, append([]any{w.rec.id, w.rend.name}, args...)...)
}

func (w *segmentWriter) writeVideo(it recItem) {
	if !w.haveSize {
		w.readSPS(it.au)
	}

	// Audio older than this AU belongs to the segment that is still open,
	// even when this AU is about to start the next one.
	w.flushHeld(it.dts, false)

	randomAccess := h264.IsRandomAccess(it.au)
	if w.cur != nil && randomAccess && it.dts-w.cur.startDTS >= w.rec.segTicks {
		w.closeSegment(it.dts)
	}
	if w.cur == nil {
		// A segment must start on a random-access AU or it cannot be played alone.
		if !randomAccess || !w.openSegment(it.dts) {
			return
		}
	}

	if err := w.ts.WriteH264(w.vid, it.pts, it.dts, it.au); err != nil {
		w.logf("write video to segment %d: %v", w.cur.index, err)
		w.abandonSegment(it.dts)
		return
	}

	if w.havePrev && it.dts > w.prevDTS {
		w.frameDur = it.dts - w.prevDTS
	}
	w.prevDTS, w.lastDTS, w.havePrev = it.dts, it.dts, true
}

// writeAudio places audio by timestamp: audio at or before the last video AU
// goes straight to the open segment; newer audio is held until video catches
// up, because the keyframe that ends this segment (or starts the first one)
// has not been seen yet.
func (w *segmentWriter) writeAudio(it recItem) {
	if w.cur == nil && w.nextIndex > 0 {
		return // between segments after a write error: wait for the next keyframe
	}
	// Before the first keyframe (cur == nil, nextIndex == 0) audio is held
	// too: it may belong to the first segment once the keyframe arrives.
	if w.cur != nil && it.pts <= w.lastDTS {
		w.emitAudio(it)
		return
	}

	w.held = append(w.held, it)
	if len(w.held) > recorderMaxHeldAudio {
		w.emitAudio(w.held[0])
		w.held = w.held[1:]
	}
}

// flushHeld writes the held audio older than limit (all of it if all) to the
// open segment, in order.
func (w *segmentWriter) flushHeld(limit int64, all bool) {
	n := 0
	for n < len(w.held) && (all || w.held[n].pts < limit) {
		w.emitAudio(w.held[n])
		n++
	}
	if n == len(w.held) {
		w.held = nil
		return
	}
	w.held = w.held[n:]
}

// emitAudio writes it to the open segment, dropping audio that is older than
// the segment's first video AU or would make the audio DTS go backwards.
func (w *segmentWriter) emitAudio(it recItem) {
	if w.cur == nil || w.aud == nil {
		return
	}
	if it.pts < w.cur.startDTS || (w.haveAudio && it.pts <= w.lastAudio) {
		w.rec.lateAudio.Add(1)
		return
	}
	if err := w.ts.WriteMPEG4Audio(w.aud, it.pts, it.au); err != nil {
		w.logf("write audio to segment %d: %v", w.cur.index, err)
		w.abandonSegment(w.lastDTS)
		return
	}
	w.lastAudio, w.haveAudio = it.pts, true
}

// readSPS takes the picture size from the first parsable SPS. A bad SPS is
// logged once per rendition and a later one is still tried.
func (w *segmentWriter) readSPS(au [][]byte) {
	for _, nalu := range au {
		if len(nalu) == 0 || h264.NALUType(nalu[0]&0x1f) != h264.NALUTypeSPS {
			continue
		}
		var sps h264.SPS
		if err := sps.Unmarshal(nalu); err != nil {
			if !w.spsFailLogged {
				w.spsFailLogged = true
				w.logf("parse SPS (further failures are not logged): %v", err)
			}
			return
		}

		w.haveSize = true
		w.rec.resMu.Lock()
		w.rend.width, w.rend.height = sps.Width(), sps.Height()
		w.rec.resMu.Unlock()
		return
	}
}

// openSegment starts {dir}/{id}/{rendition}/{n}.ts with a fresh mpegts.Writer
// so the file begins with PAT/PMT and stands alone.
func (w *segmentWriter) openSegment(dts int64) bool {
	dir := filepath.Join(w.rec.dir, w.rec.id, fmt.Sprint(w.rend.index))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		w.logf("create segment dir %s: %v", dir, err)
		return false
	}
	path := filepath.Join(dir, fmt.Sprintf("%d.ts", w.nextIndex))
	file, err := os.Create(path)
	if err != nil {
		w.logf("create segment %s: %v", path, err)
		return false
	}

	w.vid = &mpegts.Track{Codec: &codecs.H264{}}
	tracks := []*mpegts.Track{w.vid}
	w.aud = nil
	if w.rec.audio != nil {
		w.aud = &mpegts.Track{Codec: &codecs.MPEG4Audio{Config: *w.rec.audio}}
		tracks = append(tracks, w.aud)
	}
	buf := bufio.NewWriterSize(file, recorderFileBuffer)
	ts := &mpegts.Writer{W: buf, Tracks: tracks}
	if err := ts.Initialize(); err != nil {
		w.logf("init mpegts writer for %s: %v", path, err)
		discard(file, path)
		return false
	}

	w.cur = &recSegment{index: w.nextIndex, startDTS: dts, path: path}
	w.nextIndex++
	w.file, w.buf, w.ts = file, buf, ts
	return true
}

// list makes seg part of the rendition's playlist plan.
func (w *segmentWriter) list(seg *recSegment, endDTS int64) {
	w.rec.resMu.Lock()
	defer w.rec.resMu.Unlock()
	seg.endDTS = endDTS
	w.rend.segments = append(w.rend.segments, seg)
}

// closeSegment finishes the open segment, ending it at endDTS, and queues it
// for upload. The send blocks only this goroutine when the upload queue is
// full, and gives up if Finish abandons the wait.
func (w *segmentWriter) closeSegment(endDTS int64) {
	seg := w.cur
	if seg == nil {
		return
	}
	file, buf := w.file, w.buf
	w.cur, w.file, w.buf, w.ts, w.vid, w.aud = nil, nil, nil, nil, nil, nil
	w.list(seg, endDTS)

	if err := buf.Flush(); err != nil {
		w.logf("flush segment %d: %v", seg.index, err)
		discard(file, seg.path)
		w.rec.settle(seg, false, "", 0)
		return
	}
	if err := file.Close(); err != nil {
		w.logf("close segment %d: %v", seg.index, err)
		_ = os.Remove(seg.path)
		w.rec.settle(seg, false, "", 0)
		return
	}

	select {
	case w.rec.uploads <- recUpload{rendition: w.rend.index, seg: seg}:
	case <-w.rec.abandon:
		_ = os.Remove(seg.path)
	}
}

// abandonSegment drops a segment that hit a write error. It stays in the
// segment list (not ok) so the neighbours' durations are unaffected.
func (w *segmentWriter) abandonSegment(endDTS int64) {
	seg, file := w.cur, w.file
	w.cur, w.file, w.buf, w.ts, w.vid, w.aud = nil, nil, nil, nil, nil, nil
	w.list(seg, endDTS)
	discard(file, seg.path)
	w.rec.settle(seg, false, "", 0)
}

func discard(file *os.File, path string) {
	_ = file.Close()
	_ = os.Remove(path)
}

// Finish closes the open segments, waits for the uploads and writes
//
//	{publicHLSPath}/vods/{id}/{masterFileName}
//	{publicHLSPath}/vods/{id}/{i}/stream.m3u8
//
// A segment's duration is the DTS of the next segment's first AU minus its
// own; the last one adds one frame duration (the last observed DTS step) to
// the DTS of its final AU. Segments that failed to upload are left out, and a
// rendition without any segment is left out of the master. If nothing was
// uploaded Finish writes nothing and returns an error. Later writes are
// ignored and a second Finish returns an error.
//
// Finish waits at most recorderFinishTimeout in total. Segments still being
// written or uploaded then are logged and left out of the playlists.
func (r *Recorder) Finish(publicHLSPath, masterFileName string) error {
	r.mu.Lock()
	if r.finished {
		r.mu.Unlock()
		return fmt.Errorf("livestream %s: recorder already finished", r.id)
	}
	r.finished = true
	for _, rr := range r.rends {
		close(rr.in)
	}
	r.mu.Unlock()

	r.waitForWork()

	defer func() {
		tree := filepath.Join(r.dir, r.id)
		if err := os.RemoveAll(tree); err != nil {
			logger.Errorf(r.ctx, "livestream %s: remove local segments %s: %v", r.id, tree, err)
		}
	}()
	r.logSummary()

	plans := r.plans()
	if len(plans) == 0 {
		return fmt.Errorf("livestream %s: no segment was recorded and uploaded, VOD not written", r.id)
	}

	outDir := filepath.Join(publicHLSPath, "vods", r.id)
	var master strings.Builder
	master.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n")
	for _, plan := range plans {
		dir := filepath.Join(outDir, fmt.Sprint(plan.rendition))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create VOD directory %s: %w", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "stream.m3u8"), []byte(plan.playlist()), 0o644); err != nil {
			return fmt.Errorf("write VOD playlist for rendition %d: %w", plan.rendition, err)
		}
		master.WriteString(plan.masterEntry())
	}

	if err := os.WriteFile(filepath.Join(outDir, masterFileName), []byte(master.String()), 0o644); err != nil {
		return fmt.Errorf("write VOD master %s: %w", masterFileName, err)
	}
	logger.Infof(r.ctx, "livestream %s: wrote VOD with %d rendition(s) to %s", r.id, len(plans), outDir)
	return nil
}

// waitForWork waits for the writers to close their last segments and for the
// uploader to drain, for at most recorderFinishTimeout in all. On timeout it
// signals everything still running to give up.
func (r *Recorder) waitForWork() {
	timer := time.NewTimer(recorderFinishTimeout)
	defer timer.Stop()

	timedOut := false
	for _, rr := range r.rends {
		select {
		case <-rr.done:
		case <-timer.C:
			timedOut = true
		}
		if timedOut {
			break
		}
	}
	if !timedOut {
		// Every writer has returned, so nobody sends on uploads any more.
		close(r.uploads)
		select {
		case <-r.upDone:
		case <-timer.C:
			timedOut = true
		}
	}
	if timedOut {
		close(r.abandon)
		logger.Errorf(r.ctx, "livestream %s: recorder gave up waiting after %v", r.id, recorderFinishTimeout)
	}
}

// logSummary reports drops and the segments that never finished.
func (r *Recorder) logSummary() {
	if n := r.dropped.Load(); n > 0 {
		logger.Errorf(r.ctx, "livestream %s: recorder dropped %d queued item(s) in total", r.id, n)
	}
	if n := r.lateAudio.Load(); n > 0 {
		logger.Errorf(r.ctx, "livestream %s: recorder discarded %d audio unit(s) that fell outside their segment", r.id, n)
	}

	r.resMu.Lock()
	defer r.resMu.Unlock()
	for _, rr := range r.rends {
		for _, seg := range rr.segments {
			if !seg.settled {
				logger.Errorf(r.ctx, "livestream %s rendition %s: segment %d was abandoned unfinished and is left out of the VOD", r.id, rr.name, seg.index)
			}
		}
	}
}

type variantEntry struct {
	uri   string
	ticks int64
}

type variantPlan struct {
	rendition     int
	entries       []variantEntry
	bytes         int64
	totalTicks    int64
	width, height int
}

// plans snapshots every rendition that has at least one uploaded segment.
func (r *Recorder) plans() []variantPlan {
	r.resMu.Lock()
	defer r.resMu.Unlock()

	plans := make([]variantPlan, 0, len(r.rends))
	for _, rr := range r.rends {
		plan := r.planVariant(rr)
		if len(plan.entries) == 0 {
			logger.Errorf(r.ctx, "livestream %s rendition %s: no segment was recorded, leaving it out of the VOD", r.id, rr.name)
			continue
		}
		plans = append(plans, plan)
	}
	return plans
}

// planVariant lists rr's uploaded segments with their durations. resMu is held.
func (r *Recorder) planVariant(rr *recRendition) variantPlan {
	plan := variantPlan{rendition: rr.index, width: rr.width, height: rr.height}
	for _, seg := range rr.segments {
		if !seg.ok {
			continue
		}
		d := seg.endDTS - seg.startDTS
		if d <= 0 {
			logger.Errorf(r.ctx, "livestream %s rendition %s: segment %d has no duration, leaving it out", r.id, rr.name, seg.index)
			continue
		}
		plan.entries = append(plan.entries, variantEntry{
			// fileName tells players the object is a .ts, as the previous VOD writer did.
			uri:   fmt.Sprintf("%s?fileName=%s", seg.url, filepath.Base(seg.path)),
			ticks: d,
		})
		plan.bytes += seg.size
		plan.totalTicks += d
	}
	return plan
}

func seconds(t int64) float64 { return float64(t) / videoClockRate }

func (p variantPlan) playlist() string {
	var maxTicks int64
	for _, e := range p.entries {
		maxTicks = max(maxTicks, e.ticks)
	}
	target := int(math.Ceil(math.Round(seconds(maxTicks)*1000) / 1000))

	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n")
	fmt.Fprintf(&b, "#EXT-X-TARGETDURATION:%d\n", target)
	b.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	for _, e := range p.entries {
		fmt.Fprintf(&b, "#EXTINF:%.3f,\n%s\n", seconds(e.ticks), e.uri)
	}
	b.WriteString("#EXT-X-ENDLIST\n")
	return b.String()
}

func (p variantPlan) masterEntry() string {
	bandwidth := int64(float64(p.bytes) * 8 / seconds(p.totalTicks))
	attrs := fmt.Sprintf("BANDWIDTH=%d", bandwidth)
	if p.width > 0 && p.height > 0 {
		attrs += fmt.Sprintf(",RESOLUTION=%dx%d", p.width, p.height)
	}
	return fmt.Sprintf("#EXT-X-STREAM-INF:%s\n%d/stream.m3u8\n", attrs, p.rendition)
}
