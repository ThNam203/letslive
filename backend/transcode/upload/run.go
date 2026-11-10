package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/nareix/joy5/format/flv"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/config"
	"sen1or/letslive/transcode/live"
	"sen1or/letslive/transcode/storage"
)

// ErrInput marks a failure after the file was accepted for processing, which
// converting every stream may fix (a copied stream the engine could not use).
// The worker retries such a job with forceConvert. A storage failure is
// marked too: converting on its retry costs CPU but is still correct.
var ErrInput = errors.New("upload input")

// Result is what an upload produced.
type Result struct {
	Duration     time.Duration // from the first keyframe to the last video packet
	ThumbnailURL string        // "" when no thumbnail could be made
	Plan         Plan          // what intake copied and converted
}

// Run turns the file at input into VOD vodID, exactly as a recorded
// livestream: segments in st, playlists under tc.PublicHLSPath/vods/{vodID}.
// With forceConvert every stream is converted, none copied. It returns an
// error, never a partial VOD, when anything is missing.
func Run(ctx context.Context, tc config.Transcode, st storage.Storage, vodID, input string, forceConvert bool) (Result, error) {
	ffmpeg := tc.FFMpegSetting.FFMpegPath
	probe, err := RunProbe(ctx, ffmpeg, input)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInput, err)
	}
	maxGap := 2 * time.Duration(tc.Live.RecordSegmentDurationMs) * time.Millisecond
	plan, err := Decide(probe, maxGap, forceConvert, tc.FFMpegSetting.Qualities)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInput, err)
	}
	logger.Infof(ctx, "upload %s: %dx%d %s/%s, plan %+v", vodID, probe.Width, probe.Height, probe.VideoCodec, probe.AudioCodec, plan)

	mode, cancelMode := live.FileMode(ctx, tc.Upload)
	defer cancelMode()
	sess := live.NewSession(ctx, vodID, live.SessionDeps{Mode: mode, Storage: st, Transcode: tc})

	feedErr := feed(ctx, tc, input, plan, sess)

	// Close is not given a deadline: the encoder drain, each recorder upload
	// and the thumbnail are bounded on their own, and returning while the
	// session still uploads would let the retry overlap this attempt.
	closeErr := sess.Close(context.WithoutCancel(ctx))

	if err := errors.Join(feedErr, closeErr); err != nil {
		return Result{Plan: plan}, fmt.Errorf("%w: %w", ErrInput, err)
	}
	return Result{Duration: sess.Duration(), ThumbnailURL: sess.ThumbnailURL(), Plan: plan}, nil
}

// feed runs the intake ffmpeg and hands every FLV packet it writes to sess,
// until EOF or the first error.
func feed(ctx context.Context, tc config.Transcode, input string, plan Plan, sess *live.Session) error {
	ff := tc.FFMpegSetting
	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, ff.FFMpegPath, IntakeArgs(input, plan, ff.Preset, ff.CRF, tc.Upload.Threads)...)
	stderr := &tailBuffer{}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("intake stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start intake ffmpeg: %w", err)
	}
	if err := live.SetNiceness(cmd.Process.Pid, tc.Upload.Niceness); err != nil {
		logger.Warnf(ctx, "intake runs at normal priority: %v", err)
	}

	readErr := demux(stdout, sess)
	if readErr != nil {
		cancel() // ffmpeg may be blocked writing to us
	}
	waitErr := cmd.Wait()
	switch {
	case readErr != nil:
		return readErr
	case waitErr != nil:
		return fmt.Errorf("intake ffmpeg: %w (stderr: %q)", waitErr, stderr.String())
	}
	return nil
}

// demux reads FLV packets from r into sess until EOF.
func demux(r io.Reader, sess *live.Session) error {
	d := flv.NewDemuxer(r)
	for {
		pkt, err := d.ReadPacket()
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil // ffmpeg's exit status says whether the end was clean
		}
		if err != nil {
			return fmt.Errorf("read intake flv: %w", err)
		}
		if err := sess.HandlePacket(pkt); err != nil {
			return err
		}
	}
}

// tailBuffer keeps the end of ffmpeg's stderr for error messages.
type tailBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Write(p)
	if s := b.buf.String(); len(s) > 2*stderrKeep {
		b.buf.Reset()
		b.buf.WriteString(s[len(s)-stderrKeep:])
	}
	return len(p), nil
}

func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return tail(b.buf.String())
}

const stderrKeep = 4096
