package live

import (
	"bytes"
	"fmt"

	"github.com/nareix/joy5/av"

	"sen1or/letslive/shared/pkg/logger"
)

// configChange classifies an AVC config that arrives after start.
type configChange int

const (
	configBuffered      configChange = iota // before start: kept for the first keyframe
	configSame                              // identical bytes: ignored silently
	configInvalid                           // unparsable: logged and ignored
	configParamsChanged                     // same size, new parameter sets
	configResized                           // new size; the ladder stays as built
)

func (c configChange) String() string {
	return [...]string{"buffered", "same", "invalid", "params changed", "resized"}[c]
}

// onVideoConfig buffers the AVC config before start. After start (ruling
// R10) a config with different bytes replaces the parameter sets put ahead
// of source keyframes and given to the thumbnailer, and is forwarded to
// ffmpeg; the ladder and clock stay those of the first keyframe.
func (s *Session) onVideoConfig(pkt av.Packet) configChange {
	if !s.started {
		s.videoConfig = &pkt
		return configBuffered
	}
	if bytes.Equal(pkt.Data, s.videoConfig.Data) {
		return configSame
	}
	info, err := ProbeAVCConfig(pkt.Data)
	if err != nil {
		logger.Errorf(s.ctx, "%s: ignoring unparsable AVC config mid-stream: %v", s.label, err)
		s.d.Mode.fail.Report(fmt.Errorf("unparsable AVC config mid-stream: %w", err))
		return configInvalid
	}

	old := s.params
	s.videoConfig = &pkt
	s.params = info
	s.thumbs.SetSource(info)
	if s.enc != nil {
		s.enc.offerConfig(pkt)
	}

	if info.Width != old.Width || info.Height != old.Height {
		logger.Warnf(s.ctx, "%s: source resolution changed from %dx%d to %dx%d; keeping the original ladder",
			s.label, old.Width, old.Height, info.Width, info.Height)
		s.d.Mode.fail.Report(fmt.Errorf("source resolution changed from %dx%d to %dx%d",
			old.Width, old.Height, info.Width, info.Height))
		return configResized
	}
	logger.Infof(s.ctx, "%s: source sent new H.264 parameter sets (%dx%d)", s.label, info.Width, info.Height)
	return configParamsChanged
}
