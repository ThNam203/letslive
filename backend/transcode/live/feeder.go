package live

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"

	"sen1or/letslive/shared/pkg/logger"
)

// FeedTS reads one ffmpeg MPEG-TS output from r and writes its H.264 access
// units to rendition of s, on clk's timeline. After each access unit,
// including one whose write failed, onAU (may be nil) receives the
// timestamps on that same timeline; au must not be modified or reused.
//
// FeedTS returns nil on EOF, when the rendition is dropped or the stream is
// closed (both are terminal for this rendition, not failures), and when ctx
// is done. A TS without an H.264 track or a read error other than EOF is
// returned. A write error for a single access unit is logged and skipped.
//
// On ctx cancellation FeedTS returns at once and closes r if it is an
// io.Closer so the pending read ends; for a reader that cannot be closed the
// read goroutine lives until that read returns, and calls no callback.
func FeedTS(ctx context.Context, r io.Reader, s *Stream, rendition int, clk *Clock, onAU func(pts90k, dts90k int64, au [][]byte)) error {
	if rendition < 0 || rendition >= len(s.renditions) {
		return fmt.Errorf("livestream %s: rendition index %d out of range [0, %d)", s.id, rendition, len(s.renditions))
	}

	if c, ok := r.(io.Closer); ok {
		defer context.AfterFunc(ctx, func() { closeLogged(ctx, "ts reader", c) })()
	}

	done := make(chan error, 1)
	go func() { done <- readTS(ctx, r, s, rendition, clk, onAU) }()

	select {
	case err := <-done:
		if ctx.Err() != nil {
			return nil
		}
		return err
	case <-ctx.Done():
		return nil
	}
}

func readTS(ctx context.Context, r io.Reader, s *Stream, rendition int, clk *Clock, onAU func(pts90k, dts90k int64, au [][]byte)) error {
	name := s.renditions[rendition].name

	reader := &mpegts.Reader{R: r}
	if err := reader.Initialize(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("livestream %s rendition %s: read ts: %w", s.id, name, err)
	}

	var track *mpegts.Track
	for _, t := range reader.Tracks() {
		if _, ok := t.Codec.(*codecs.H264); ok {
			track = t
			break
		}
	}
	if track == nil {
		return fmt.Errorf("livestream %s rendition %s: ts has no H.264 track", s.id, name)
	}

	unwrap := clk.NewTSUnwrapper()
	stopped := false
	reader.OnDataH264(track, func(rawPTS, rawDTS int64, au [][]byte) error {
		if err := ctx.Err(); err != nil {
			stopped = true
			return err
		}
		pts, dts := unwrap.Rel(rawPTS), unwrap.Rel(rawDTS)

		err := s.WriteVideo(rendition, clk.NTP(pts), pts, au)
		switch {
		case errors.Is(err, ErrRenditionDropped), errors.Is(err, ErrStreamClosed):
			stopped = true
			return err
		case err != nil:
			logger.Errorf(ctx, "livestream %s rendition %s: write video at pts %d: %v", s.id, name, pts, err)
		}

		if onAU != nil {
			onAU(pts, dts, au)
		}
		return nil
	})

	for {
		err := reader.Read()
		switch {
		case err == nil:
		case stopped, errors.Is(err, io.EOF):
			return nil
		default:
			return fmt.Errorf("livestream %s rendition %s: read ts: %w", s.id, name, err)
		}
	}
}
