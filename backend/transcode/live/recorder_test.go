package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/nareix/joy5/av"

	"sen1or/letslive/transcode/storage"
)

const (
	recorderTestTimeout = 30 * time.Second
	recorderTestSegDur  = 2 * time.Second
)

// recorderRig wires a Recorder to a recording storage and temp dirs.
type recorderRig struct {
	id      string
	f       streamFixture
	st      *recordingStorage
	rec     *Recorder
	cancel  context.CancelFunc
	workDir string
	pubDir  string
}

func newRecorderRig(t *testing.T, fixture string, names []string) *recorderRig {
	t.Helper()
	return newRecorderRigWith(t, fixture, names, nil)
}

// newRecorderRigWith is newRecorderRig with the Recorder talking to
// wrap(rig.st) instead of rig.st directly (nil: rig.st).
func newRecorderRigWith(t *testing.T, fixture string, names []string, wrap func(*recordingStorage) storage.Storage) *recorderRig {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), recorderTestTimeout)
	t.Cleanup(cancel)

	rig := &recorderRig{
		id:      "stream-1",
		f:       loadStreamFixture(t, fixture),
		st:      newRecordingStorage(t),
		cancel:  cancel,
		workDir: t.TempDir(),
		pubDir:  t.TempDir(),
	}
	var st storage.Storage = rig.st
	if wrap != nil {
		st = wrap(rig.st)
	}
	rs := make([]Rendition, len(names))
	for i, name := range names {
		rs[i] = Rendition{Name: name, Encoded: i > 0}
	}
	rig.rec = NewRecorder(ctx, rig.id, rs, rig.f.audio, st, rig.workDir, recorderTestSegDur)
	return rig
}

// feed hands pkts to the recorder the way a session does: video to every
// rendition (keyframes carry SPS/PPS), audio once on the 90 kHz timeline.
func (r *recorderRig) feed(t *testing.T, renditions int, pkts []av.Packet) {
	t.Helper()
	for _, pkt := range pkts {
		switch pkt.Type {
		case av.H264:
			_, au, err := r.f.videoAU(pkt)
			if err != nil {
				t.Fatal(err)
			}
			pts := r.f.clock.Video90k(pkt.Time + pkt.CTime)
			dts := r.f.clock.Video90k(pkt.Time)
			for i := range renditions {
				r.rec.WriteVideo(i, pts, dts, au)
			}
		case av.AAC:
			if r.f.audio != nil {
				r.rec.WriteAudio(r.f.clock.Video90k(pkt.Time), [][]byte{pkt.Data})
			}
		}
	}
}

func (r *recorderRig) finish(t *testing.T) error {
	t.Helper()
	return r.rec.Finish(r.pubDir, "index.m3u8")
}

func (r *recorderRig) read(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(r.pubDir, "vods", r.id, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

var extinfRE = regexp.MustCompile(`^#EXTINF:([0-9.]+),$`)

// parseVariant returns the EXTINF durations and segment URIs of a variant playlist.
func parseVariant(t *testing.T, playlist string) (durs []float64, uris []string) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(playlist), "\n")
	for i, line := range lines {
		m := extinfRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		d, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			t.Fatalf("bad EXTINF %q: %v", line, err)
		}
		if i+1 >= len(lines) {
			t.Fatalf("EXTINF without a URI line in:\n%s", playlist)
		}
		durs = append(durs, d)
		uris = append(uris, lines[i+1])
	}
	return durs, uris
}

func TestRecorderVOD(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	rig.feed(t, 1, rig.f.pkts)

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}

	variant := rig.read(t, "0/stream.m3u8")
	for _, want := range []string{"#EXTM3U\n", "#EXT-X-VERSION:3\n", "#EXT-X-PLAYLIST-TYPE:VOD\n", "#EXT-X-ENDLIST\n"} {
		if !strings.Contains(variant, want) {
			t.Errorf("variant playlist lacks %q:\n%s", want, variant)
		}
	}
	durs, uris := parseVariant(t, variant)
	if len(durs) != 3 {
		t.Fatalf("EXTINF lines = %d, want 3:\n%s", len(durs), variant)
	}
	var total, longest float64
	for i, uri := range uris {
		want := fmt.Sprintf("%s/%s/0/%d.ts?fileName=%d.ts", recordedURLPrefix, rig.id, i, i)
		if uri != want {
			t.Errorf("segment %d uri = %q, want %q", i, uri, want)
		}
		total += durs[i]
		longest = max(longest, durs[i])
	}
	wantTarget := fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", int(math.Ceil(longest)))
	if !strings.Contains(variant, wantTarget) {
		t.Errorf("variant playlist lacks %q:\n%s", wantTarget, variant)
	}
	if math.Abs(total-6) > 0.25 {
		t.Errorf("total duration = %.3f, want about 6s", total)
	}
	if calls := rig.st.Calls(); len(calls) != 3 {
		t.Errorf("AddSegment calls = %d, want 3", len(calls))
	}

	master := rig.read(t, "index.m3u8")
	if !strings.Contains(master, "RESOLUTION=640x360") {
		t.Errorf("master lacks RESOLUTION=640x360:\n%s", master)
	}
	m := regexp.MustCompile(`BANDWIDTH=(\d+)`).FindStringSubmatch(master)
	if m == nil {
		t.Fatalf("master lacks BANDWIDTH:\n%s", master)
	}
	if bw, _ := strconv.Atoi(m[1]); bw <= 0 {
		t.Errorf("BANDWIDTH = %s, want > 0", m[1])
	}
	if !strings.Contains(master, "\n0/stream.m3u8") {
		t.Errorf("master does not point at 0/stream.m3u8:\n%s", master)
	}

	if _, err := os.Stat(filepath.Join(rig.workDir, rig.id)); !os.IsNotExist(err) {
		t.Errorf("local segment tree still present after Finish: %v", err)
	}
}

func TestRecorderPartialSegmentOnFinish(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	rig.feed(t, 1, rig.f.before(3*time.Second))

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}

	durs, uris := parseVariant(t, rig.read(t, "0/stream.m3u8"))
	if len(durs) != 2 {
		t.Fatalf("segments = %d, want 2 (full + partial); uris %v", len(durs), uris)
	}
	if math.Abs(durs[1]-1) > 0.1 {
		t.Errorf("partial segment duration = %.3f, want about 1s", durs[1])
	}
	if calls := rig.st.Calls(); len(calls) != 2 {
		t.Errorf("AddSegment calls = %d, want 2", len(calls))
	}
}

func TestRecorderUploadFailure(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	rig.st.before = func(call storageCall) error {
		if call.Index == 1 {
			return errors.New("minio down")
		}
		return nil
	}
	rig.feed(t, 1, rig.f.pkts)

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}

	durs, uris := parseVariant(t, rig.read(t, "0/stream.m3u8"))
	if len(durs) != 2 {
		t.Fatalf("playlist entries = %d, want 2; uris %v", len(durs), uris)
	}
	for _, uri := range uris {
		if strings.Contains(uri, "/1.ts") {
			t.Errorf("failed segment 1 is listed: %s", uri)
		}
	}
	if _, err := os.Stat(filepath.Join(rig.workDir, rig.id)); !os.IsNotExist(err) {
		t.Errorf("failed segment's local file survived Finish: %v", err)
	}
}

// A segment whose upload failed is not in the VOD, so its local file is
// removed at once instead of filling the disk until Finish while MinIO is down.
func TestRecorderRemovesFailedUploadAtOnce(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	failed := make(chan struct{})
	rig.st.before = func(call storageCall) error {
		if call.Index == 1 {
			close(failed)
			return errors.New("minio down")
		}
		return nil
	}
	rig.feed(t, 1, rig.f.pkts)

	select {
	case <-failed:
	case <-time.After(recorderTestTimeout):
		t.Fatal("segment 1 was never uploaded")
	}
	seg1 := filepath.Join(rig.workDir, rig.id, "0", "1.ts")
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, err := os.Stat(seg1)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("failed segment still on disk 5s after its upload failed (stat: %v)", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if _, uris := parseVariant(t, rig.read(t, "0/stream.m3u8")); len(uris) != 2 {
		t.Errorf("playlist lists %d segments, want 2 (the failed one left out): %v", len(uris), uris)
	}
}

// A dropped encoded rendition is left out of the VOD; the source cannot be
// dropped.
func TestRecorderDropRendition(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source", "360p"})
	rig.feed(t, 2, rig.f.before(3*time.Second))
	rig.rec.DropRendition(1)
	rig.rec.DropRendition(0)
	rig.feed(t, 2, rig.f.from(3*time.Second))

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	master := rig.read(t, "index.m3u8")
	if !strings.Contains(master, "0/stream.m3u8") || strings.Contains(master, "1/stream.m3u8") {
		t.Errorf("master should list only the source:\n%s", master)
	}
	if durs, _ := parseVariant(t, rig.read(t, "0/stream.m3u8")); len(durs) != 3 {
		t.Errorf("source has %d segments, want all 3", len(durs))
	}
}

// Every PTS/DTS in a VOD segment is shifted by recorderTSOffset, so an AU
// slightly before the clock zero is still written non-negative; one too early
// even for the offset is dropped.
func TestRecorderTimestampOffset(t *testing.T) {
	requireFFmpeg(t)
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	rig.rec.WriteVideo(0, recorderMinTimestamp-1, recorderMinTimestamp-1, [][]byte{{0x65, 0x88}})
	rig.feed(t, 1, rig.f.pkts)
	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}

	var first string
	for _, c := range rig.st.Calls() {
		if c.Kind == "segment" && filepath.Base(c.SavedPath) == "0.ts" {
			first = c.SavedPath
		}
	}
	if first == "" {
		t.Fatal("segment 0 was not uploaded")
	}
	video, audio := probeFirstDTS(t, first)
	if want := float64(recorderTSOffset) / videoClockRate; math.Abs(video-want) > 1e-6 {
		t.Errorf("first video DTS = %.6fs, want the %.1fs offset (first keyframe at the clock zero)", video, want)
	}
	if audio < 0 || audio > maxSaneFirstDTS {
		t.Errorf("first audio DTS = %.6fs, want non-negative and unwrapped", audio)
	}
	if got := logsContaining("too early for the TS offset"); len(got) == 0 {
		t.Error("the too-early AU was not reported")
	}
}

func TestRecorderNoSegmentsWritesNothing(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	rig.st.before = func(storageCall) error { return errors.New("minio down") }
	rig.feed(t, 1, rig.f.pkts)

	if err := rig.finish(t); err == nil {
		t.Fatal("Finish = nil, want an error when no segment was uploaded")
	}
	if _, err := os.Stat(filepath.Join(rig.pubDir, "vods", rig.id)); !os.IsNotExist(err) {
		t.Errorf("VOD directory written despite no segments: %v", err)
	}

	empty := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	if err := empty.finish(t); err == nil {
		t.Fatal("Finish without any media = nil, want an error")
	}
}

func TestRecorderFinishIdempotentAndIgnoresLateWrites(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	rig.feed(t, 1, rig.f.before(3*time.Second))

	if err := rig.finish(t); err != nil {
		t.Fatalf("first Finish: %v", err)
	}
	callsAfter := len(rig.st.Calls())

	rig.feed(t, 1, rig.f.from(3*time.Second))
	rig.rec.WriteVideo(7, 0, 0, [][]byte{{0x65}})
	if err := rig.finish(t); err == nil {
		t.Error("second Finish = nil, want an error")
	}
	if got := len(rig.st.Calls()); got != callsAfter {
		t.Errorf("writes after Finish changed AddSegment calls: %d -> %d", callsAfter, got)
	}
	durs, _ := parseVariant(t, rig.read(t, "0/stream.m3u8"))
	if len(durs) != 2 {
		t.Errorf("playlist changed after late writes: %d entries", len(durs))
	}
}

func TestRecorderSlowStorageNeverBlocksWriters(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	release := make(chan struct{})
	rig.st.before = func(storageCall) error {
		<-release
		return nil
	}

	fed := make(chan struct{})
	go func() {
		rig.feed(t, 1, rig.f.pkts)
		close(fed)
	}()
	select {
	case <-fed:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("WriteVideo/WriteAudio blocked behind a stalled upload")
	}
	close(release)

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if durs, _ := parseVariant(t, rig.read(t, "0/stream.m3u8")); len(durs) != 3 {
		t.Errorf("segments = %d, want 3", len(durs))
	}
}

func TestRecorderDropsWhenQueueFull(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), recorderTestTimeout)
	defer cancel()

	f := loadStreamFixture(t, "src_640x360.flv")
	st := newRecordingStorage(t)
	release := make(chan struct{})
	st.before = func(storageCall) error {
		<-release
		return nil
	}
	// A tiny item queue and no upload buffer: the stalled first upload blocks
	// the rendition goroutine at the second cut, so the item queue overflows.
	rec := newRecorder(ctx, "stream-1", []Rendition{{Name: "source"}}, nil, st, t.TempDir(), recorderTestSegDur, 8, 0)

	rig := &recorderRig{f: f, rec: rec}
	fed := make(chan struct{})
	go func() {
		rig.feed(t, 1, f.pkts)
		close(fed)
	}()
	select {
	case <-fed:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("writers blocked on a full queue")
	}
	if rec.Dropped() == 0 {
		t.Error("Dropped() = 0, want > 0 after overflowing the queue")
	}
	close(release)

	if err := rec.Finish(t.TempDir(), "index.m3u8"); err != nil {
		t.Fatalf("Finish: %v", err)
	}
}

type ffprobeStreams struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		StartTime string `json:"start_time"`
	} `json:"streams"`
}

func probeStreamStarts(t *testing.T, path string) (video, audio float64, err error) {
	t.Helper()
	out, runErr := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=codec_type,start_time", "-of", "json", path).Output()
	if runErr != nil {
		return 0, 0, fmt.Errorf("ffprobe %s: %w", path, runErr)
	}
	var p ffprobeStreams
	if err := json.Unmarshal(out, &p); err != nil {
		return 0, 0, err
	}
	video, audio = math.NaN(), math.NaN()
	for _, s := range p.Streams {
		start, err := strconv.ParseFloat(s.StartTime, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("start_time %q: %w", s.StartTime, err)
		}
		switch s.CodecType {
		case "video":
			video = start
		case "audio":
			audio = start
		}
	}
	return video, audio, nil
}

// TestRecorderSegmentsArePlayableWithAudio pins ruling R3: every rendition's
// files are self-contained TS with video and audio on one timeline.
func TestRecorderSegmentsArePlayableWithAudio(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found in PATH")
	}

	rig := newRecorderRig(t, "src_640x360.flv", []string{"source", "360p"})
	rig.feed(t, 2, rig.f.pkts)
	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}

	calls := rig.st.Calls()
	if len(calls) != 6 {
		t.Fatalf("AddSegment calls = %d, want 6 (3 per rendition)", len(calls))
	}
	perQuality := map[int]int{}
	for _, c := range calls {
		perQuality[c.QualityIndex]++
		video, audio, err := probeStreamStarts(t, c.SavedPath)
		if err != nil {
			t.Fatal(err)
		}
		if math.IsNaN(video) || math.IsNaN(audio) {
			t.Errorf("%s: video start %v, audio start %v; want both streams", filepath.Base(c.SavedPath), video, audio)
			continue
		}
		if math.Abs(video-audio) > 0.25 {
			t.Errorf("%s (q%d): video starts %.3fs, audio %.3fs: A/V out of alignment", filepath.Base(c.SavedPath), c.QualityIndex, video, audio)
		}
	}
	if perQuality[0] != 3 || perQuality[1] != 3 {
		t.Errorf("segments per rendition = %v, want 3 each", perQuality)
	}

	master := rig.read(t, "index.m3u8")
	for _, want := range []string{"\n0/stream.m3u8", "\n1/stream.m3u8"} {
		if !strings.Contains(master, want) {
			t.Errorf("master lacks %q:\n%s", want, master)
		}
	}
}

func TestRecorderVideoOnlySource(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360_noaudio.flv", []string{"source"})
	rig.feed(t, 1, rig.f.pkts)
	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if durs, _ := parseVariant(t, rig.read(t, "0/stream.m3u8")); len(durs) != 3 {
		t.Errorf("segments = %d, want 3", len(durs))
	}
}

func TestRecorderConcurrentWritersRace(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source", "360p"})
	var started atomic.Int32
	done := make(chan struct{})
	for r := range 2 {
		go func() {
			started.Add(1)
			for _, pkt := range rig.f.pkts {
				if pkt.Type != av.H264 {
					continue
				}
				_, au, err := rig.f.videoAU(pkt)
				if err != nil {
					t.Error(err)
					break
				}
				rig.rec.WriteVideo(r, rig.f.clock.Video90k(pkt.Time+pkt.CTime), rig.f.clock.Video90k(pkt.Time), au)
			}
			done <- struct{}{}
		}()
	}
	for range 2 {
		<-done
	}
	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}
}

type ffprobePackets struct {
	Packets []struct {
		CodecType string `json:"codec_type"`
		DTSTime   string `json:"dts_time"`
	} `json:"packets"`
}

// probeFirstDTS returns the smallest DTS (seconds) of each stream's packets
// and fails the test if either stream's DTS is not strictly increasing in
// file order.
func probeFirstDTS(t *testing.T, path string) (video, audio float64) {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "packet=codec_type,dts_time", "-of", "json", path).Output()
	if err != nil {
		t.Fatalf("ffprobe %s: %v", path, err)
	}
	var p ffprobePackets
	if err := json.Unmarshal(out, &p); err != nil {
		t.Fatal(err)
	}
	video, audio = math.Inf(1), math.Inf(1)
	last := map[string]float64{}
	for _, pk := range p.Packets {
		d, err := strconv.ParseFloat(pk.DTSTime, 64)
		if err != nil {
			t.Fatalf("dts_time %q: %v", pk.DTSTime, err)
		}
		if prev, seen := last[pk.CodecType]; seen && d <= prev {
			t.Errorf("%s: %s DTS goes %.4f -> %.4f in file order", filepath.Base(path), pk.CodecType, prev, d)
		}
		last[pk.CodecType] = d
		switch pk.CodecType {
		case "video":
			video = math.Min(video, d)
		case "audio":
			audio = math.Min(audio, d)
		}
	}
	return video, audio
}

// requireCleanDecode fails when ffmpeg reports a timestamp problem for path.
// It stream-copies instead of decoding: a decode re-times frames onto the
// output frame rate, and for some audio/video phases the null muxer then
// reports "non monotonically increasing dts" for a perfectly ordered input.
func requireCleanDecode(t *testing.T, path string) {
	t.Helper()
	out, err := exec.Command("ffmpeg", "-v", "warning", "-i", path, "-c", "copy", "-f", "null", "-").CombinedOutput()
	if err != nil {
		t.Errorf("ffmpeg %s: %v\n%s", filepath.Base(path), err, out)
		return
	}
	if strings.Contains(strings.ToLower(string(out)), "monoton") {
		t.Errorf("%s has non-monotonic timestamps:\n%s", filepath.Base(path), out)
	}
}

// feedSkewed hands pkts to the recorder in arrival order after delaying every
// audio packet by skew (negative: early), the way RTMP interleaving jitter
// does around a keyframe.
func (r *recorderRig) feedSkewed(t *testing.T, skew time.Duration) {
	t.Helper()
	pkts := append([]av.Packet(nil), r.f.pkts...)
	arrival := func(p av.Packet) time.Duration {
		if p.Type == av.AAC {
			return p.Time + skew
		}
		return p.Time
	}
	sort.SliceStable(pkts, func(i, j int) bool { return arrival(pkts[i]) < arrival(pkts[j]) })
	r.feed(t, 1, pkts)
}

// requireSegmentsAligned checks each uploaded segment: no timestamp warnings
// and the first audio DTS within one AAC frame (plus ms rounding) after the
// first video DTS, i.e. audio was assigned to the segment by timestamp.
func requireSegmentsAligned(t *testing.T, rig *recorderRig) {
	t.Helper()
	// An AAC frame is 1024 samples; RTMP timestamps are whole milliseconds,
	// which adds up to 1ms of rounding on each side.
	frame := 1024 / float64(rig.f.audio.SampleRate)
	const rounding = 0.002

	calls := rig.st.Calls()
	if len(calls) != 3 {
		t.Fatalf("segments = %d, want 3", len(calls))
	}
	for _, c := range calls {
		requireCleanDecode(t, c.SavedPath)
		video, audio := probeFirstDTS(t, c.SavedPath)
		name := filepath.Base(c.SavedPath)
		if math.IsInf(audio, 1) {
			t.Errorf("%s has no audio", name)
			continue
		}
		skew := audio - video
		if skew < -rounding {
			t.Errorf("%s: audio starts %.3fs before the first video DTS", name, -skew)
		}
		// The first segment starts where the fixture's audio does, which is
		// up to 3 frames after the first keyframe; later ones start at a cut.
		limit := frame + rounding
		if c.SavedPath == calls[0].SavedPath {
			limit = 3*frame + rounding
		}
		if skew > limit {
			t.Errorf("%s: audio starts %.3fs after the first video DTS, want <= %.3fs", name, skew, limit)
		}
	}
}

func TestRecorderAudioAssignedByTimestamp(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found in PATH")
	}
	requireFFmpeg(t)

	for _, tc := range []struct {
		name string
		skew time.Duration
	}{
		{"in order", 0},
		{"audio late", 150 * time.Millisecond},
		{"audio early", -150 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
			rig.feedSkewed(t, tc.skew)
			if err := rig.finish(t); err != nil {
				t.Fatalf("Finish: %v", err)
			}
			requireSegmentsAligned(t, rig)
		})
	}
}

func TestDropLimiter(t *testing.T) {
	l := dropLimiter{interval: 5 * time.Second}
	t0 := time.Unix(1000, 0)

	if n, ok := l.record(t0); !ok || n != 1 {
		t.Fatalf("first drop: (%d, %v), want (1, true)", n, ok)
	}
	for _, d := range []time.Duration{time.Second, 2 * time.Second, 4900 * time.Millisecond} {
		if _, ok := l.record(t0.Add(d)); ok {
			t.Errorf("drop at +%v reported inside the interval", d)
		}
	}
	if n, ok := l.record(t0.Add(5 * time.Second)); !ok || n != 4 {
		t.Errorf("drop at +5s: (%d, %v), want (4, true): count since the last report", n, ok)
	}
	if _, ok := l.record(t0.Add(6 * time.Second)); ok {
		t.Error("drop at +6s reported 1s after the previous report")
	}
	if got := l.flush(); got != 1 {
		t.Errorf("flush = %d, want 1 unreported drop", got)
	}
	if got := l.flush(); got != 0 {
		t.Errorf("second flush = %d, want 0", got)
	}
}

func TestSegmentWriterReadSPS(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	newWriter := func() *segmentWriter {
		return &segmentWriter{rec: &Recorder{ctx: context.Background(), id: "x"}, rend: &recRendition{name: "source"}}
	}

	w := newWriter()
	w.readSPS([][]byte{{}, nil}) // empty NAL units must not panic
	if w.rend.width != 0 {
		t.Fatalf("width = %d after empty NAL units", w.rend.width)
	}

	w.readSPS([][]byte{{byte(h264.NALUTypeSPS), 0x00}}) // truncated SPS
	w.readSPS([][]byte{{byte(h264.NALUTypeSPS), 0x00}})
	if w.rend.width != 0 || !w.spsFailLogged {
		t.Fatalf("after bad SPS: width %d, spsFailLogged %v", w.rend.width, w.spsFailLogged)
	}

	// A later valid SPS is still picked up after earlier failures.
	w.readSPS([][]byte{f.sps})
	if w.rend.width != 640 || w.rend.height != 360 {
		t.Errorf("size = %dx%d, want 640x360", w.rend.width, w.rend.height)
	}
}

func TestRecorderIgnoresEmptyNALUnits(t *testing.T) {
	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})

	for _, pkt := range rig.f.pkts {
		switch pkt.Type {
		case av.H264:
			_, au, err := rig.f.videoAU(pkt)
			if err != nil {
				t.Fatal(err)
			}
			// An RTMP publisher can send zero-length NAL units anywhere in an AU.
			au = append([][]byte{{}, nil}, au...)
			au = append(au, []byte{})
			rig.rec.WriteVideo(0, rig.f.clock.Video90k(pkt.Time+pkt.CTime), rig.f.clock.Video90k(pkt.Time), au)
		case av.AAC:
			rig.rec.WriteAudio(rig.f.clock.Video90k(pkt.Time), [][]byte{{}, pkt.Data})
		}
	}
	rig.rec.WriteVideo(0, 0, 0, [][]byte{{}}) // an AU of nothing but empty NAL units

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if durs, _ := parseVariant(t, rig.read(t, "0/stream.m3u8")); len(durs) != 3 {
		t.Errorf("segments = %d, want 3", len(durs))
	}
	if !strings.Contains(rig.read(t, "index.m3u8"), "RESOLUTION=640x360") {
		t.Error("master lacks RESOLUTION=640x360")
	}
}

// shortRecorderTimeouts shrinks the package timeouts for one test.
func shortRecorderTimeouts(t *testing.T, finish, upload time.Duration) {
	t.Helper()
	oldFinish, oldUpload := recorderFinishTimeout, recorderUploadTimeout
	recorderFinishTimeout, recorderUploadTimeout = finish, upload
	t.Cleanup(func() { recorderFinishTimeout, recorderUploadTimeout = oldFinish, oldUpload })
}

func TestRecorderFinishBoundedWhenStorageHangs(t *testing.T) {
	const bound = 400 * time.Millisecond
	shortRecorderTimeouts(t, bound, 30*time.Second)

	rig := newRecorderRig(t, "src_640x360.flv", []string{"source"})
	hang := make(chan struct{})
	rig.st.before = func(call storageCall) error {
		if call.Index >= 1 {
			<-hang // a storage call that never returns on its own
		}
		return nil
	}
	rig.feed(t, 1, rig.f.pkts)

	start := time.Now()
	result := make(chan error, 1)
	go func() { result <- rig.finish(t) }()
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Finish: %v", err)
		}
	case <-time.After(10 * time.Second):
		close(hang)
		t.Fatal("Finish did not return although storage hung")
	}
	if took := time.Since(start); took > bound+5*time.Second {
		t.Errorf("Finish took %v, want about %v", took, bound)
	}

	// The segment that finished before the hang is still listed.
	if durs, _ := parseVariant(t, rig.read(t, "0/stream.m3u8")); len(durs) != 1 {
		t.Errorf("playlist entries = %d, want 1 (only the upload that finished)", len(durs))
	}

	close(hang)
	select {
	case <-rig.rec.upDone:
	case <-time.After(10 * time.Second):
		t.Error("uploader goroutine did not stop after the abandoned call returned")
	}
}

// ctxStorage fails calls whose context is done and records their deadlines.
type ctxStorage struct {
	inner *recordingStorage
	gate  chan struct{}

	mu        sync.Mutex
	deadlines []time.Time
	hadNone   int
}

func (s *ctxStorage) AddSegment(ctx context.Context, filePath, streamId string, q int) (string, error) {
	<-s.gate
	s.mu.Lock()
	if dl, ok := ctx.Deadline(); ok {
		s.deadlines = append(s.deadlines, dl)
	} else {
		s.hadNone++
	}
	s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return s.inner.AddSegment(ctx, filePath, streamId, q)
}

func (s *ctxStorage) AddThumbnail(ctx context.Context, filePath, streamId, contentType string) (string, error) {
	return s.inner.AddThumbnail(ctx, filePath, streamId, contentType)
}

func TestRecorderUploadsSurviveCancelledContext(t *testing.T) {
	cs := &ctxStorage{gate: make(chan struct{})}
	rig := newRecorderRigWith(t, "src_640x360.flv", []string{"source"}, func(rs *recordingStorage) storage.Storage {
		cs.inner = rs
		return cs
	})
	rig.feed(t, 1, rig.f.pkts)

	// The caller's context ends with the stream, before the uploads run.
	rig.cancel()
	close(cs.gate)

	if err := rig.finish(t); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if durs, _ := parseVariant(t, rig.read(t, "0/stream.m3u8")); len(durs) != 3 {
		t.Errorf("segments = %d after cancelling the parent context, want 3", len(durs))
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.hadNone != 0 || len(cs.deadlines) != 3 {
		t.Fatalf("AddSegment contexts: %d with a deadline, %d without; want 3 and 0", len(cs.deadlines), cs.hadNone)
	}
	for _, dl := range cs.deadlines {
		if left := time.Until(dl); left > recorderUploadTimeout || left < recorderUploadTimeout-10*time.Second {
			t.Errorf("upload deadline in %v, want about %v", left, recorderUploadTimeout)
		}
	}
}
