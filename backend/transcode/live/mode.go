package live

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"

	"sen1or/letslive/transcode/config"
)

// Mode holds the strategies in which a livestream and an uploaded file
// differ. LiveMode and FileMode build it; the Session and its parts call the
// strategies and never ask which mode they were given. One Mode value serves
// one Session.
type Mode struct {
	// noun starts every log line: "livestream" or "upload".
	noun string
	// open creates where viewers get the stream while it runs.
	open func(id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig) (Output, error)
	// bp is what a full encoder or recorder queue does.
	bp Backpressure
	// fail is told about everything that makes the recording incomplete.
	fail Failures
	// profile is the ffmpeg tuning of the encoded renditions.
	profile EncoderProfile
	// endEncoder ends ffmpeg when the session closes.
	endEncoder encoderEnd
	// thumbs picks the keyframes that become thumbnails; every is the
	// configured thumbnail interval.
	thumbs func(every time.Duration) ThumbSchedule
}

// LiveMode is the mode of an RTMP livestream: viewers watch it through reg
// while it runs.
func LiveMode(reg *Registry, cfg config.Live) (Mode, error) {
	if reg == nil {
		return Mode{}, errors.New("live mode needs a registry")
	}
	return Mode{
		noun: "livestream",
		open: func(id string, rs []Rendition, audio *mpeg4audio.AudioSpecificConfig) (Output, error) {
			stream, err := NewStream(id, rs, audio, cfg)
			if err != nil {
				return nil, err
			}
			if err := reg.Add(stream); err != nil {
				stream.Close()
				return nil, fmt.Errorf("register stream: %w", err)
			}
			return liveOutput{Stream: stream, reg: reg}, nil
		},
		bp:         dropWhenFull{},
		fail:       logOnly{},
		profile:    liveProfile,
		endEncoder: (*sessionEncoder).terminate,
		thumbs:     func(every time.Duration) ThumbSchedule { return newWallSchedule(every) },
	}, nil
}

// label is what log lines about session id start with.
func (m Mode) label(id string) string { return m.noun + " " + id }

// fileThumbnailAt is the media time of an upload's thumbnail; earlier
// keyframes are often a black or title frame.
const fileThumbnailAt = 5 * time.Second

// FileMode is the mode of an uploaded file: nobody watches it live, full
// queues wait instead of dropping, the first lost segment or quality fails the
// session (so the job is retried), ffmpeg uses normal tuning at the niceness
// in u and drains at Close, and one thumbnail is taken at fileThumbnailAt of
// media time. cancel releases the mode's context; call it once the session
// is closed.
func FileMode(ctx context.Context, u config.Upload) (m Mode, cancel context.CancelFunc) {
	fctx, cancel := context.WithCancel(ctx)
	return Mode{
		noun: "upload",
		open: func(_ string, rs []Rendition, _ *mpeg4audio.AudioSpecificConfig) (Output, error) {
			return newNoOutput(rs), nil
		},
		bp:         waitUntil{ctx: fctx},
		fail:       &firstError{cancel: cancel},
		profile:    fileProfile(u),
		endEncoder: (*sessionEncoder).drain,
		thumbs:     func(time.Duration) ThumbSchedule { return &mediaSchedule{min: fileThumbnailAt} },
	}, cancel
}
