package live

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/storage"
)

const (
	thumbnailFileName = "thumbnail.jpg"
	thumbnailWidth    = 640

	// thumbnailStderrLimit caps the ffmpeg stderr kept for a log line.
	thumbnailStderrLimit = 2048

	// thumbnailWaitDelay bounds how long Wait lingers on ffmpeg's pipes after the
	// process was killed.
	thumbnailWaitDelay = 2 * time.Second
)

// Timeouts are variables only so tests can shorten them.
var (
	// thumbnailFFmpegTimeout bounds the one-shot ffmpeg decode.
	thumbnailFFmpegTimeout = 15 * time.Second
	// thumbnailUploadTimeout bounds the AddThumbnail call. Its context is
	// detached from the caller's, so the upload still finishes when the caller's
	// context ends together with the stream.
	thumbnailUploadTimeout = 30 * time.Second
	// thumbnailStartDelay skips the stream's first keyframes, which are often
	// a black or title frame (the old ffmpeg filter used gte(t,5)).
	thumbnailStartDelay = 5 * time.Second
)

var annexBStartCode = []byte{0, 0, 0, 1}

// Thumbnailer turns the keyframes its ThumbSchedule picks into JPEGs. It
// decodes only the offered keyframe with a one-shot ffmpeg and uploads the
// result, one job at a time, never blocking the caller of Offer.
type Thumbnailer struct {
	ctx        context.Context
	id         string
	label      string // starts every log line
	ffmpegPath string
	st         storage.Storage

	mu      sync.Mutex // guards the fields below; held only for O(1) work
	header  []byte     // Annex-B SPS and PPS; empty disables the thumbnailer
	sched   ThumbSchedule
	closed  bool
	running bool
	first   []byte // Annex-B stream of the first usable keyframe, for Fallback
	url     string // of the last uploaded thumbnail
	jobs    sync.WaitGroup
}

// NewThumbnailer returns a livestream's Thumbnailer: the first thumbnail after
// thumbnailStartDelay, then one every `every` of wall-clock time.
func NewThumbnailer(ctx context.Context, label, id, ffmpegPath string, src SourceInfo, st storage.Storage, every time.Duration) *Thumbnailer {
	return newThumbnailer(ctx, label, id, ffmpegPath, src, st, newWallSchedule(every))
}

// newThumbnailer returns a Thumbnailer for session id that logs under label
// and takes the keyframes sched picks. ctx is only used for logging; uploads
// and ffmpeg runs get their own bounded contexts.
func newThumbnailer(ctx context.Context, label, id, ffmpegPath string, src SourceInfo, st storage.Storage, sched ThumbSchedule) *Thumbnailer {
	t := &Thumbnailer{
		ctx:        ctx,
		id:         id,
		label:      label,
		ffmpegPath: ffmpegPath,
		st:         st,
		sched:      sched,
	}
	t.header = t.paramHeader(src)
	return t
}

// SetSource replaces the SPS and PPS put ahead of later offers, after the
// publisher sent a new AVC config. Without both, thumbnails are disabled
// rather than decoded with stale parameter sets.
func (t *Thumbnailer) SetSource(src SourceInfo) {
	header := t.paramHeader(src)
	t.mu.Lock()
	t.header = header
	t.mu.Unlock()
}

// paramHeader is src's SPS and PPS in Annex-B form, or nil (logged) when
// either is missing.
func (t *Thumbnailer) paramHeader(src SourceInfo) []byte {
	if len(src.SPS) == 0 || len(src.PPS) == 0 {
		logger.Errorf(t.ctx, "%s: thumbnails disabled, source has no SPS/PPS", t.label)
		return nil
	}
	return appendAnnexB(nil, [][]byte{src.SPS, src.PPS})
}

// Offer submits a keyframe access unit (NAL units without start codes) at
// media time `media` from the first keyframe. It returns at once. The access
// unit is ignored when the schedule does not want it, when a thumbnail job is
// running, after Close, or when it holds no data. The first usable one is
// kept for the schedule's Fallback. The access unit is copied; the caller may
// reuse it.
func (t *Thumbnailer) Offer(au [][]byte, media time.Duration) {
	t.mu.Lock()
	header := t.header
	t.mu.Unlock()
	if len(header) == 0 {
		return
	}
	if !hasNALData(au) {
		logger.Errorf(t.ctx, "%s: thumbnail skipped, access unit is empty", t.label)
		return
	}

	stream := appendAnnexB(append([]byte(nil), header...), au)
	t.mu.Lock()
	if t.first == nil {
		t.first = stream
	}
	if t.closed || t.running || !t.sched.Want(media) {
		t.mu.Unlock()
		return
	}
	t.running = true
	t.jobs.Add(1)
	t.mu.Unlock()

	go func() {
		defer t.jobs.Done()
		t.run(stream)
		t.mu.Lock()
		t.running = false
		t.mu.Unlock()
	}()
}

// Close stops accepting offers and waits for a job in flight, which is bounded
// by the ffmpeg and upload timeouts. When the schedule asks for its Fallback,
// the first keyframe is turned into the thumbnail first. It is safe to call
// more than once.
func (t *Thumbnailer) Close() {
	t.mu.Lock()
	wasClosed := t.closed
	t.closed = true
	t.mu.Unlock()
	t.jobs.Wait()
	if wasClosed {
		return
	}

	t.mu.Lock()
	first, fallback := t.first, t.sched.Fallback()
	t.mu.Unlock()
	if fallback && first != nil {
		t.run(first)
	}
}

// URL is the URL of the last uploaded thumbnail, or "" when none was.
func (t *Thumbnailer) URL() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.url
}

func hasNALData(au [][]byte) bool {
	for _, nalu := range au {
		if len(nalu) > 0 {
			return true
		}
	}
	return false
}

// appendAnnexB appends each non-empty NAL unit to dst behind a start code.
func appendAnnexB(dst []byte, nalus [][]byte) []byte {
	for _, nalu := range nalus {
		if len(nalu) == 0 {
			continue
		}
		dst = append(dst, annexBStartCode...)
		dst = append(dst, nalu...)
	}
	return dst
}

// run decodes stream into a JPEG in a private temp dir, uploads it and removes
// the dir. Every failure is logged; none is returned because nobody waits.
func (t *Thumbnailer) run(stream []byte) {
	dir, err := os.MkdirTemp("", "thumbnail-"+t.id+"-")
	if err != nil {
		logger.Errorf(t.ctx, "%s: create thumbnail temp dir: %v", t.label, err)
		return
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			logger.Errorf(t.ctx, "%s: remove thumbnail temp dir %s: %v", t.label, dir, err)
		}
	}()

	out := filepath.Join(dir, thumbnailFileName)
	if err := t.decode(stream, out); err != nil {
		logger.Errorf(t.ctx, "%s: thumbnail: %v", t.label, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.ctx), thumbnailUploadTimeout)
	defer cancel()
	url, err := t.st.AddThumbnail(ctx, out, t.id, "image/jpeg")
	if err != nil {
		logger.Errorf(t.ctx, "%s: upload thumbnail: %v", t.label, err)
		return
	}
	t.mu.Lock()
	t.url = url
	t.mu.Unlock()
}

// decode runs ffmpeg once over stream, writing the first picture to out.
func (t *Thumbnailer) decode(stream []byte, out string) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.ctx), thumbnailFFmpegTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, t.ffmpegPath,
		"-hide_banner", "-loglevel", "error",
		"-f", "h264", "-i", "pipe:0",
		"-frames:v", "1",
		"-vf", fmt.Sprintf("scale=%d:-2", thumbnailWidth),
		"-q:v", "2",
		"-f", "image2",
		out,
	)
	cmd.Stdin = bytes.NewReader(stream)
	var stderr bytes.Buffer
	cmd.Stderr = &limitedWriter{buf: &stderr, limit: thumbnailStderrLimit}
	cmd.WaitDelay = thumbnailWaitDelay

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w (stderr: %q)", err, stderr.String())
	}
	info, err := os.Stat(out)
	if err != nil {
		return fmt.Errorf("ffmpeg produced no output file: %w (stderr: %q)", err, stderr.String())
	}
	if info.Size() == 0 {
		return fmt.Errorf("ffmpeg produced an empty output file (stderr: %q)", stderr.String())
	}
	return nil
}

// limitedWriter keeps the first `limit` bytes written and discards the rest,
// reporting every write as complete so the writer never blocks the process.
type limitedWriter struct {
	buf   *bytes.Buffer
	limit int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if room := w.limit - w.buf.Len(); room > 0 {
		w.buf.Write(p[:min(len(p), room)])
	}
	return len(p), nil
}
