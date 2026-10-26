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
)

// recItem is one queued write for a rendition's goroutine.
type recItem struct {
	isAudio bool
	pts     int64
	dts     int64
	au      [][]byte // video NAL units, or AAC access units
}

// recSegment is one MPEG-TS file. The rendition goroutine owns it until it
// hands it to the uploader (which then sets url, size and ok); Finish reads it
// only after both have stopped.
type recSegment struct {
	index    int
	startDTS int64
	endDTS   int64 // DTS of the next segment's first AU (see Finish for the last)
	path     string

	ok   bool
	url  string
	size int64
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

	// dropping is true from the first dropped item of a run until the next
	// accepted one, so each run of drops is logged once.
	dropping atomic.Bool

	// Written by the rendition goroutine, read by Finish after done closes.
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
	dropped atomic.Int64

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
	}

	for i, rend := range rs {
		rr := &recRendition{index: i, name: rend.Name, in: make(chan recItem, itemQueue), done: make(chan struct{})}
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
// not be modified afterwards. It never blocks: a full queue drops the item.
func (r *Recorder) WriteVideo(rendition int, pts90k, dts90k int64, au [][]byte) {
	if len(au) == 0 {
		return
	}
	r.enqueue(rendition, recItem{pts: pts90k, dts: dts90k, au: au})
}

// WriteAudio queues AAC access units, stamped on the 90 kHz timeline, for
// every rendition so each VOD variant is self-contained. The slices must not
// be modified afterwards. It never blocks.
func (r *Recorder) WriteAudio(pts90k int64, aus [][]byte) {
	if r.audio == nil || len(aus) == 0 {
		return
	}
	for i := range r.rends {
		r.enqueue(i, recItem{isAudio: true, pts: pts90k, au: aus})
	}
}

// Dropped returns how many items were discarded because a queue was full.
func (r *Recorder) Dropped() int64 { return r.dropped.Load() }

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
		rr.dropping.Store(false)
	default:
		r.dropped.Add(1)
		if !rr.dropping.Swap(true) {
			logger.Errorf(r.ctx, "livestream %s rendition %s: recorder queue full, dropping input", r.id, rr.name)
		}
	}
}

func (r *Recorder) uploadLoop() {
	for job := range r.uploads {
		seg := job.seg
		info, err := os.Stat(seg.path)
		if err != nil {
			logger.Errorf(r.ctx, "livestream %s rendition %d: stat segment %s: %v", r.id, job.rendition, seg.path, err)
			continue
		}
		url, err := r.st.AddSegment(r.ctx, seg.path, r.id, job.rendition)
		if err != nil {
			logger.Errorf(r.ctx, "livestream %s rendition %d: upload segment %d: %v", r.id, job.rendition, seg.index, err)
			continue
		}
		seg.url, seg.size, seg.ok = url, info.Size(), true
		if err := os.Remove(seg.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			logger.Errorf(r.ctx, "livestream %s rendition %d: remove uploaded segment %s: %v", r.id, job.rendition, seg.path, err)
		}
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
}

func (w *segmentWriter) run() {
	for it := range w.rend.in {
		if it.isAudio {
			w.writeAudio(it)
		} else {
			w.writeVideo(it)
		}
	}
	w.closeSegment(w.lastDTS + w.frameDur)
}

func (w *segmentWriter) logf(format string, args ...any) {
	logger.Errorf(w.rec.ctx, "livestream %s rendition %s: "+format, append([]any{w.rec.id, w.rend.name}, args...)...)
}

func (w *segmentWriter) writeVideo(it recItem) {
	if w.rend.width == 0 {
		w.readSPS(it.au)
	}

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

func (w *segmentWriter) writeAudio(it recItem) {
	if w.cur == nil || w.aud == nil {
		return
	}
	if err := w.ts.WriteMPEG4Audio(w.aud, it.pts, it.au); err != nil {
		w.logf("write audio to segment %d: %v", w.cur.index, err)
		w.abandonSegment(w.lastDTS)
	}
}

func (w *segmentWriter) readSPS(au [][]byte) {
	for _, nalu := range au {
		if h264.NALUType(nalu[0]&0x1f) != h264.NALUTypeSPS {
			continue
		}
		var sps h264.SPS
		if err := sps.Unmarshal(nalu); err != nil {
			w.logf("parse SPS: %v", err)
			return
		}
		w.rend.width, w.rend.height = sps.Width(), sps.Height()
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

// closeSegment finishes the open segment, ending it at endDTS, and queues it
// for upload. The send blocks only this goroutine when the upload queue is full.
func (w *segmentWriter) closeSegment(endDTS int64) {
	seg := w.cur
	if seg == nil {
		return
	}
	file, buf := w.file, w.buf
	w.cur, w.file, w.buf, w.ts, w.vid, w.aud = nil, nil, nil, nil, nil, nil

	seg.endDTS = endDTS
	w.rend.segments = append(w.rend.segments, seg)

	if err := buf.Flush(); err != nil {
		w.logf("flush segment %d: %v", seg.index, err)
		discard(file, seg.path)
		return
	}
	if err := file.Close(); err != nil {
		w.logf("close segment %d: %v", seg.index, err)
		return
	}
	w.rec.uploads <- recUpload{rendition: w.rend.index, seg: seg}
}

// abandonSegment drops a segment that hit a write error. It stays in the
// segment list (not ok) so the neighbours' durations are unaffected.
func (w *segmentWriter) abandonSegment(endDTS int64) {
	seg, file := w.cur, w.file
	w.cur, w.file, w.buf, w.ts, w.vid, w.aud = nil, nil, nil, nil, nil, nil
	seg.endDTS = endDTS
	w.rend.segments = append(w.rend.segments, seg)
	discard(file, seg.path)
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

	for _, rr := range r.rends {
		<-rr.done
	}
	close(r.uploads)
	<-r.upDone

	defer func() {
		tree := filepath.Join(r.dir, r.id)
		if err := os.RemoveAll(tree); err != nil {
			logger.Errorf(r.ctx, "livestream %s: remove local segments %s: %v", r.id, tree, err)
		}
	}()

	outDir := filepath.Join(publicHLSPath, "vods", r.id)
	var master strings.Builder
	master.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n")
	written := 0

	plans := make([]variantPlan, 0, len(r.rends))
	for _, rr := range r.rends {
		plan := r.planVariant(rr)
		if len(plan.entries) == 0 {
			logger.Errorf(r.ctx, "livestream %s rendition %s: no segment was recorded, leaving it out of the VOD", r.id, rr.name)
			continue
		}
		plans = append(plans, plan)
	}
	if len(plans) == 0 {
		return fmt.Errorf("livestream %s: no segment was recorded and uploaded, VOD not written", r.id)
	}

	for _, plan := range plans {
		dir := filepath.Join(outDir, fmt.Sprint(plan.rendition))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create VOD directory %s: %w", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "stream.m3u8"), []byte(plan.playlist()), 0o644); err != nil {
			return fmt.Errorf("write VOD playlist for rendition %d: %w", plan.rendition, err)
		}
		master.WriteString(plan.masterEntry())
		written++
	}

	if err := os.WriteFile(filepath.Join(outDir, masterFileName), []byte(master.String()), 0o644); err != nil {
		return fmt.Errorf("write VOD master %s: %w", masterFileName, err)
	}
	logger.Infof(r.ctx, "livestream %s: wrote VOD with %d rendition(s) to %s", r.id, written, outDir)
	return nil
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

// planVariant lists rr's uploaded segments with their durations.
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
