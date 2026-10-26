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
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nareix/joy5/av"
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
	workDir string
	pubDir  string
}

func newRecorderRig(t *testing.T, fixture string, names []string) *recorderRig {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), recorderTestTimeout)
	t.Cleanup(cancel)

	rig := &recorderRig{
		id:      "stream-1",
		f:       loadStreamFixture(t, fixture),
		st:      newRecordingStorage(t),
		workDir: t.TempDir(),
		pubDir:  t.TempDir(),
	}
	rs := make([]Rendition, len(names))
	for i, name := range names {
		rs[i] = Rendition{Name: name, Encoded: i > 0}
	}
	rig.rec = NewRecorder(ctx, rig.id, rs, rig.f.audio, rig.st, rig.workDir, recorderTestSegDur)
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
