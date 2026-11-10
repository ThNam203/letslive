package live

import (
	"errors"
	"fmt"

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
		bp:   dropWhenFull{},
		fail: logOnly{},
	}, nil
}

// label is what log lines about session id start with.
func (m Mode) label(id string) string { return m.noun + " " + id }
