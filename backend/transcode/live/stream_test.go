package live

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/gofrs/uuid/v5"
	"github.com/nareix/joy5/av"

	"sen1or/letslive/transcode/config"
)

const (
	notReadyDeadline = 100 * time.Millisecond
	blockedFor       = 100 * time.Millisecond
	unblockDeadline  = 3 * time.Second
	testHTTPTimeout  = 10 * time.Second
)

// streamFixture holds an FLV fixture's media packets, starting at the first
// keyframe, plus what a session would extract from its sequence headers.
type streamFixture struct {
	pkts     []av.Packet
	sps, pps []byte
	audio    *mpeg4audio.AudioSpecificConfig
	clock    *Clock
}

func loadStreamFixture(t *testing.T, name string) streamFixture {
	t.Helper()

	var f streamFixture
	started := false
	for _, pkt := range readFLV(t, name) {
		switch pkt.Type {
		case av.H264DecoderConfig:
			info, err := ProbeAVCConfig(pkt.Data)
			if err != nil {
				t.Fatalf("ProbeAVCConfig: %v", err)
			}
			f.sps, f.pps = info.SPS, info.PPS
		case av.AACDecoderConfig:
			asc, err := AudioConfig(pkt)
			if err != nil {
				t.Fatalf("AudioConfig: %v", err)
			}
			f.audio = asc
		case av.H264:
			if !started && pkt.IsKeyFrame {
				started = true
				f.clock = NewClock(time.Now(), pkt.Time)
			}
			if started {
				f.pkts = append(f.pkts, pkt)
			}
		case av.AAC:
			if started {
				f.pkts = append(f.pkts, pkt)
			}
		}
	}
	if !started || f.sps == nil {
		t.Fatalf("%s has no keyframe or no AVC config", name)
	}
	return f
}

func (f streamFixture) before(d time.Duration) []av.Packet {
	for i, pkt := range f.pkts {
		if pkt.Time >= d {
			return f.pkts[:i]
		}
	}
	return f.pkts
}

func (f streamFixture) from(d time.Duration) []av.Packet {
	return f.pkts[len(f.before(d)):]
}

// videoAU converts an RTMP video packet the way a session does: keyframes get
// the SPS and PPS from the sequence header so gohlslib can derive DTS.
func (f streamFixture) videoAU(pkt av.Packet) (pts90k int64, au [][]byte, err error) {
	nalus, err := VideoAU(pkt)
	if err != nil {
		return 0, nil, err
	}
	if pkt.IsKeyFrame && !hasNALU(nalus, h264.NALUTypeSPS) {
		nalus = append([][]byte{f.sps, f.pps}, nalus...)
	}
	return f.clock.Video90k(pkt.Time + pkt.CTime), nalus, nil
}

func hasNALU(au [][]byte, typ h264.NALUType) bool {
	for _, nalu := range au {
		if h264.NALUType(nalu[0]&0x1f) == typ {
			return true
		}
	}
	return false
}

func (f streamFixture) writeVideo(s *Stream, r int, pkt av.Packet) error {
	pts, au, err := f.videoAU(pkt)
	if err != nil {
		return err
	}
	if err := s.WriteVideo(r, f.clock.NTP(pts), pts, au); err != nil {
		return fmt.Errorf("WriteVideo(%d) at %v: %w", r, pkt.Time, err)
	}
	return nil
}

func (f streamFixture) writeAudio(s *Stream, pkt av.Packet) error {
	pts := f.clock.Audio(pkt.Time, f.audio.SampleRate)
	ntp := f.clock.NTP(f.clock.Video90k(pkt.Time))
	if err := s.WriteAudio(ntp, pts, [][]byte{pkt.Data}); err != nil {
		return fmt.Errorf("WriteAudio at %v: %w", pkt.Time, err)
	}
	return nil
}

// feed writes pkts in order; video goes to every rendition.
func (f streamFixture) feed(t *testing.T, s *Stream, renditions int, pkts []av.Packet) {
	t.Helper()
	for _, pkt := range pkts {
		switch pkt.Type {
		case av.H264:
			for r := range renditions {
				if err := f.writeVideo(s, r, pkt); err != nil {
					t.Fatal(err)
				}
			}
		case av.AAC:
			if f.audio != nil {
				if err := f.writeAudio(s, pkt); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func testLiveConfig() config.Live {
	return config.Live{}.WithDefaults()
}

func newTestStream(t *testing.T, names []string, audio *mpeg4audio.AudioSpecificConfig, cfg config.Live) *Stream {
	t.Helper()
	rs := make([]Rendition, len(names))
	for i, name := range names {
		rs[i] = Rendition{Name: name, Encoded: i > 0}
	}
	s, err := NewStream(uuid.Must(uuid.NewV4()).String(), rs, audio, cfg)
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func serveStream(t *testing.T, s *Stream) (*httptest.Server, string) {
	t.Helper()
	reg := NewRegistry()
	if err := reg.Add(s); err != nil {
		t.Fatalf("Add: %v", err)
	}
	srv := httptest.NewServer(reg.Handler())
	t.Cleanup(srv.Close)
	return srv, srv.URL + "/" + s.ID()
}

type response struct {
	status int
	header http.Header
	body   string
}

func get(ctx context.Context, url string) (response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return response{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return response{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response{}, err
	}
	return response{status: resp.StatusCode, header: resp.Header, body: string(body)}, nil
}

func mustGet(t *testing.T, url string, wantStatus int) response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testHTTPTimeout)
	defer cancel()
	resp, err := get(ctx, url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	if resp.status != wantStatus {
		t.Fatalf("GET %s: status %d, want %d; body:\n%s", url, resp.status, wantStatus, resp.body)
	}
	return resp
}

// requireNotReady asserts the master answers 404 quickly instead of blocking.
func requireNotReady(t *testing.T, base string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), notReadyDeadline)
	defer cancel()
	resp, err := get(ctx, base+"/index.m3u8")
	if err != nil {
		t.Fatalf("master did not answer within %v: %v", notReadyDeadline, err)
	}
	if resp.status != http.StatusNotFound {
		t.Fatalf("master status = %d, want 404; body:\n%s", resp.status, resp.body)
	}
	if cc := resp.header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

// variantURI returns the URI line after the first EXT-X-STREAM-INF.
func variantURI(t *testing.T, master string) string {
	t.Helper()
	sc := bufio.NewScanner(strings.NewReader(master))
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "#EXT-X-STREAM-INF:") && sc.Scan() {
			return sc.Text()
		}
	}
	t.Fatalf("master has no variant:\n%s", master)
	return ""
}

// quotedAttr returns the value of the first attr="..." on a line starting with tag.
func quotedAttr(t *testing.T, playlist, tag, attr string) string {
	t.Helper()
	for _, line := range strings.Split(playlist, "\n") {
		if !strings.HasPrefix(line, tag) {
			continue
		}
		_, rest, ok := strings.Cut(line, attr+`="`)
		if !ok {
			continue
		}
		value, _, _ := strings.Cut(rest, `"`)
		return value
	}
	t.Fatalf("no %s with %s in:\n%s", tag, attr, playlist)
	return ""
}

// nextMSN returns the Media Sequence Number of the segment still being
// written: EXT-X-MEDIA-SEQUENCE plus the number of complete segments.
func nextMSN(t *testing.T, playlist string) uint64 {
	t.Helper()
	var seq uint64
	found := false
	complete := uint64(0)
	for _, line := range strings.Split(playlist, "\n") {
		if v, ok := strings.CutPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"); ok {
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				t.Fatalf("bad media sequence %q: %v", v, err)
			}
			seq, found = n, true
		}
		if strings.HasPrefix(line, "#EXTINF:") {
			complete++
		}
	}
	if !found {
		t.Fatalf("no EXT-X-MEDIA-SEQUENCE in:\n%s", playlist)
	}
	return seq + complete
}

func TestStreamServesLLHLS(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	if f.audio == nil {
		t.Fatal("fixture has no AAC config")
	}
	s := newTestStream(t, []string{"source"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	// Keyframes are at 0s, 2s and 4s: feeding up to 3s closes exactly one
	// segment and leaves the next one half written.
	f.feed(t, s, 1, f.before(3*time.Second))

	master := mustGet(t, base+"/index.m3u8", http.StatusOK)
	if ct := master.header.Get("Content-Type"); ct != "application/vnd.apple.mpegurl" {
		t.Errorf("master Content-Type = %q", ct)
	}
	for _, want := range []string{"source/", `GROUP-ID="audio-source"`, `AUDIO="audio-source"`} {
		if !strings.Contains(master.body, want) {
			t.Errorf("master lacks %q:\n%s", want, master.body)
		}
	}

	mediaURL := base + "/" + variantURI(t, master.body)
	media := mustGet(t, mediaURL, http.StatusOK)
	for _, want := range []string{"#EXT-X-PART:", "#EXT-X-PRELOAD-HINT"} {
		if !strings.Contains(media.body, want) {
			t.Errorf("media playlist lacks %q:\n%s", want, media.body)
		}
	}

	audioURI := quotedAttr(t, master.body, "#EXT-X-MEDIA:", "URI")
	mustGet(t, base+"/"+audioURI, http.StatusOK)

	initFile := quotedAttr(t, media.body, "#EXT-X-MAP:", "URI")
	init := mustGet(t, base+"/source/"+initFile, http.StatusOK)
	if ct := init.header.Get("Content-Type"); ct != "video/mp4" {
		t.Errorf("init Content-Type = %q, want video/mp4", ct)
	}
	mustGet(t, base+"/source/no_such_file.mp4", http.StatusNotFound)

	// _HLS_msn names the segment after the one being written, so the reload
	// can only be answered once the next keyframe (4s) closes the current one.
	msn := nextMSN(t, media.body) + 1
	ctx, cancel := context.WithTimeout(context.Background(), testHTTPTimeout)
	defer cancel()
	type result struct {
		resp response
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := get(ctx, fmt.Sprintf("%s?_HLS_msn=%d", mediaURL, msn))
		done <- result{resp, err}
	}()

	select {
	case r := <-done:
		t.Fatalf("blocking reload returned before new media: status %d, err %v\n%s", r.resp.status, r.err, r.resp.body)
	case <-time.After(blockedFor):
	}

	f.feed(t, s, 1, f.from(3*time.Second))

	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("blocking reload: %v", r.err)
		}
		if r.resp.status != http.StatusOK {
			t.Fatalf("blocking reload status = %d; body:\n%s", r.resp.status, r.resp.body)
		}
		if got := nextMSN(t, r.resp.body); got < msn {
			t.Errorf("reload answered with next MSN %d, want >= %d:\n%s", got, msn, r.resp.body)
		}
	case <-time.After(unblockDeadline):
		t.Fatalf("blocking reload did not return within %v of new media", unblockDeadline)
	}
}

func TestStreamMasterNotReady(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	requireNotReady(t, base)

	// One keyframe only: the first segment is still open.
	f.feed(t, s, 1, f.before(1500*time.Millisecond))
	requireNotReady(t, base)
}

// A second keyframe earlier than SegmentMinDuration does not close a segment
// in gohlslib, so the master must stay 404 rather than block on index.m3u8.
func TestStreamMasterWaitsForSegmentMinDuration(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	cfg := testLiveConfig()
	cfg.SegmentMinDurationMs = 3000
	s := newTestStream(t, []string{"source"}, f.audio, cfg)
	_, base := serveStream(t, s)

	f.feed(t, s, 1, f.before(3*time.Second)) // keyframes at 0s and 2s
	requireNotReady(t, base)

	f.feed(t, s, 1, f.from(3*time.Second)) // keyframe at 4s closes [0s, 4s)
	mustGet(t, base+"/index.m3u8", http.StatusOK)
}

func TestStreamVideoOnly(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360_noaudio.flv")
	if f.audio != nil {
		t.Fatal("no-audio fixture has an AAC config")
	}
	s := newTestStream(t, []string{"source", "360p"}, nil, testLiveConfig())
	_, base := serveStream(t, s)

	if err := s.WriteAudio(time.Now(), 0, [][]byte{{1}}); err == nil {
		t.Error("WriteAudio on a video-only stream succeeded")
	}

	f.feed(t, s, 2, f.pkts)

	master := mustGet(t, base+"/index.m3u8", http.StatusOK)
	if strings.Contains(master.body, "EXT-X-MEDIA") || strings.Contains(master.body, "AUDIO=") {
		t.Errorf("video-only master has audio:\n%s", master.body)
	}
	for _, want := range []string{"\nsource/video1_stream.m3u8\n", "\n360p/video1_stream.m3u8\n"} {
		if !strings.Contains(master.body, want) {
			t.Errorf("master lacks %q:\n%s", want, master.body)
		}
	}
}

// Video for each rendition and the shared audio arrive from different
// goroutines while viewers poll; run with -race.
func TestStreamConcurrentWriters(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source", "360p"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	var writers sync.WaitGroup
	for r := range 2 {
		writers.Go(func() {
			for _, pkt := range f.pkts {
				if pkt.Type != av.H264 {
					continue
				}
				if err := f.writeVideo(s, r, pkt); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	writers.Go(func() {
		for _, pkt := range f.pkts {
			if pkt.Type != av.AAC {
				continue
			}
			if err := f.writeAudio(s, pkt); err != nil {
				t.Error(err)
				return
			}
		}
	})

	stop := make(chan struct{})
	var pollers sync.WaitGroup
	pollers.Go(func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			ctx, cancel := context.WithTimeout(context.Background(), testHTTPTimeout)
			resp, err := get(ctx, base+"/index.m3u8")
			cancel()
			if err != nil {
				t.Errorf("poll master: %v", err)
				return
			}
			if resp.status != http.StatusOK && resp.status != http.StatusNotFound {
				t.Errorf("poll master status = %d", resp.status)
				return
			}
		}
	})

	writers.Wait()
	close(stop)
	pollers.Wait()

	master := mustGet(t, base+"/index.m3u8", http.StatusOK)
	for _, want := range []string{`GROUP-ID="audio-source"`, `GROUP-ID="audio-360p"`, "\nsource/", "\n360p/"} {
		if !strings.Contains(master.body, want) {
			t.Errorf("master lacks %q:\n%s", want, master.body)
		}
	}
}

func TestStreamWriteValidation(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source"}, f.audio, testLiveConfig())

	pts, au, err := f.videoAU(f.pkts[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []int{-1, 1} {
		if err := s.WriteVideo(r, f.clock.NTP(pts), pts, au); err == nil {
			t.Errorf("WriteVideo(%d) succeeded", r)
		}
	}

	s.Close()
	if err := s.WriteVideo(0, f.clock.NTP(pts), pts, au); err == nil {
		t.Error("WriteVideo after Close succeeded")
	}
	if err := s.WriteAudio(f.clock.NTP(0), 0, [][]byte{{1}}); err == nil {
		t.Error("WriteAudio after Close succeeded")
	}
	s.Close() // idempotent
}

func TestNewStreamValidation(t *testing.T) {
	cfg := testLiveConfig()
	id := uuid.Must(uuid.NewV4()).String()
	source := Rendition{Name: "source"}

	cases := map[string]struct {
		rs    []Rendition
		audio *mpeg4audio.AudioSpecificConfig
		cfg   config.Live
	}{
		"no renditions":     {nil, nil, cfg},
		"bad name":          {[]Rendition{{Name: "Source"}}, nil, cfg},
		"duplicate name":    {[]Rendition{source, source}, nil, cfg},
		"zero sample rate":  {[]Rendition{source}, &mpeg4audio.AudioSpecificConfig{Type: 2, ChannelCount: 2}, cfg},
		"too few segments":  {[]Rendition{source}, nil, config.Live{SegmentCount: 6, SegmentMinDurationMs: 1000, PartMinDurationMs: 200}},
		"zero part length":  {[]Rendition{source}, nil, config.Live{SegmentCount: 7, SegmentMinDurationMs: 1000}},
		"zero segment time": {[]Rendition{source}, nil, config.Live{SegmentCount: 7, PartMinDurationMs: 200}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := NewStream(id, tc.rs, tc.audio, tc.cfg)
			if err == nil {
				s.Close()
				t.Fatal("NewStream succeeded")
			}
		})
	}
}

func TestStreamClosedServes404(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)
	f.feed(t, s, 1, f.pkts)
	mustGet(t, base+"/index.m3u8", http.StatusOK)

	s.Close()
	mustGet(t, base+"/index.m3u8", http.StatusNotFound)
	mustGet(t, base+"/source/video1_stream.m3u8", http.StatusNotFound)
}

func TestRegistry(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source"}, f.audio, testLiveConfig())
	f.feed(t, s, 1, f.pkts)

	reg := NewRegistry()
	if err := reg.Add(s); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := reg.Add(s); err == nil {
		t.Error("duplicate Add succeeded")
	}
	if err := reg.Add(nil); err == nil {
		t.Error("Add(nil) succeeded")
	}
	notUUID, err := NewStream("not-a-uuid", []Rendition{{Name: "source"}}, nil, testLiveConfig())
	if err != nil {
		t.Fatalf("NewStream: %v", err)
	}
	defer notUUID.Close()
	if err := reg.Add(notUUID); err == nil {
		t.Error("Add with a non-UUID id succeeded")
	}

	h := reg.Handler()
	serve := func(method, target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
		return rec
	}

	id := s.ID()
	cases := []struct {
		method, target string
		want           int
	}{
		{http.MethodGet, "/" + id + "/index.m3u8", http.StatusOK},
		{http.MethodGet, "/" + strings.ToUpper(id) + "/index.m3u8", http.StatusOK},
		{http.MethodGet, "/" + id + "/source/video1_stream.m3u8", http.StatusOK},
		{http.MethodPost, "/" + id + "/index.m3u8", http.StatusMethodNotAllowed},
		{http.MethodGet, "/" + id + "/720p/video1_stream.m3u8", http.StatusNotFound},
		{http.MethodGet, "/" + id + "/other.m3u8", http.StatusNotFound},
		{http.MethodGet, "/" + id + "/source/a/video1_stream.m3u8", http.StatusNotFound},
		{http.MethodGet, "/" + id, http.StatusNotFound},
		{http.MethodGet, "/", http.StatusNotFound},
		{http.MethodGet, "/" + uuid.Must(uuid.NewV4()).String() + "/index.m3u8", http.StatusNotFound},
		{http.MethodGet, "/../x/index.m3u8", http.StatusBadRequest},
		{http.MethodGet, "/..%2Fx/index.m3u8", http.StatusBadRequest},
		{http.MethodGet, "/not-a-uuid/source/video1_stream.m3u8", http.StatusBadRequest},
	}
	for _, tc := range cases {
		if rec := serve(tc.method, tc.target); rec.Code != tc.want {
			t.Errorf("%s %s = %d, want %d; body %q", tc.method, tc.target, rec.Code, tc.want, rec.Body.String())
		}
	}

	reg.Remove(id)
	if rec := serve(http.MethodGet, "/"+id+"/index.m3u8"); rec.Code != http.StatusNotFound {
		t.Errorf("after Remove: status %d, want 404", rec.Code)
	}
	if err := reg.Add(s); err != nil {
		t.Errorf("Add after Remove: %v", err)
	}
}

func requireMasterLists(t *testing.T, base string, want, absent []string) {
	t.Helper()
	master := mustGet(t, base+"/index.m3u8", http.StatusOK)
	for _, name := range want {
		if !strings.Contains(master.body, "\n"+name+"/video1_stream.m3u8\n") {
			t.Errorf("master lacks rendition %s:\n%s", name, master.body)
		}
	}
	for _, name := range absent {
		if strings.Contains(master.body, name+"/") {
			t.Errorf("master still lists rendition %s:\n%s", name, master.body)
		}
	}
}

// An encoded rendition whose ffmpeg dies before its first segment must not
// keep the master 404 forever once it is dropped.
func TestStreamDropRenditionUnblocksMaster(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	f.feed(t, s, 1, f.before(3*time.Second)) // source only
	requireNotReady(t, base)

	s.DropRendition(1)
	requireMasterLists(t, base, []string{"source"}, []string{"480p"})

	// Audio still reaches the source; the dropped rendition is skipped.
	f.feed(t, s, 1, f.from(3*time.Second))
	mustGet(t, base+"/source/audio2_stream.m3u8", http.StatusOK)
}

func TestStreamDropRenditionInvalidatesCache(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	f.feed(t, s, 2, f.pkts)
	requireMasterLists(t, base, []string{"source", "480p"}, nil)

	s.DropRendition(1)
	s.DropRendition(1) // idempotent
	requireMasterLists(t, base, []string{"source"}, []string{"480p"})
	mustGet(t, base+"/480p/index.m3u8", http.StatusNotFound)
	mustGet(t, base+"/480p/video1_stream.m3u8", http.StatusNotFound)
	mustGet(t, base+"/source/video1_stream.m3u8", http.StatusOK)

	pts, au, err := f.videoAU(f.pkts[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := s.WriteVideo(1, f.clock.NTP(pts), pts, au); !errors.Is(err, ErrRenditionDropped) {
		t.Errorf("WriteVideo to dropped rendition = %v, want ErrRenditionDropped", err)
	}
}

func TestStreamDropSourceIsNoop(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	s.DropRendition(0)
	s.DropRendition(-1)
	s.DropRendition(2)

	f.feed(t, s, 2, f.pkts)
	requireMasterLists(t, base, []string{"source", "480p"}, nil)
	mustGet(t, base+"/source/video1_stream.m3u8", http.StatusOK)
}

// Dropping while writers and viewers are active; run with -race.
func TestStreamDropRenditionConcurrent(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())
	_, base := serveStream(t, s)

	var wg sync.WaitGroup
	for r := range 2 {
		wg.Go(func() {
			for _, pkt := range f.pkts {
				if pkt.Type != av.H264 {
					continue
				}
				err := f.writeVideo(s, r, pkt)
				if r == 1 && errors.Is(err, ErrRenditionDropped) {
					return
				}
				if err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Go(func() {
		for _, pkt := range f.pkts {
			if pkt.Type != av.AAC {
				continue
			}
			if err := f.writeAudio(s, pkt); err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Go(func() {
		for range 20 {
			ctx, cancel := context.WithTimeout(context.Background(), testHTTPTimeout)
			resp, err := get(ctx, base+"/index.m3u8")
			cancel()
			if err != nil {
				t.Errorf("poll master: %v", err)
				return
			}
			if resp.status != http.StatusOK && resp.status != http.StatusNotFound {
				t.Errorf("poll master status = %d", resp.status)
				return
			}
		}
	})
	wg.Go(func() { s.DropRendition(1) })
	wg.Wait()

	requireMasterLists(t, base, []string{"source"}, []string{"480p"})
}

func TestStreamReady(t *testing.T) {
	f := loadStreamFixture(t, "src_640x360.flv")
	s := newTestStream(t, []string{"source", "480p"}, f.audio, testLiveConfig())

	for _, r := range []int{-1, 0, 1, 2} {
		if s.Ready(r) {
			t.Errorf("Ready(%d) = true before any write", r)
		}
	}

	f.feed(t, s, 1, f.pkts) // source only
	if !s.Ready(0) {
		t.Error("Ready(0) = false after the source closed its first segment")
	}
	if s.Ready(1) {
		t.Error("Ready(1) = true for a rendition that got no video")
	}

	s.DropRendition(1)
	if s.Ready(1) {
		t.Error("Ready(1) = true for a dropped rendition")
	}
	if s.Ready(2) || s.Ready(-1) {
		t.Error("Ready out of range = true")
	}
}
