package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultPartMinDurationMs       = 200
	defaultSegmentMinDurationMs    = 1000
	defaultSegmentCount            = 7
	defaultRecordSegmentDurationMs = 4000
	defaultEncoderQueuePackets     = 512
	defaultThumbnailIntervalMs     = 60000

	// LL-HLS requires the playlist to hold at least this many segments.
	minSegmentCount = 7
)

type Quality struct {
	Resolution string `yaml:"resolution"`
	MaxBitrate string `yaml:"maxBitrate"`
	FPS        int    `yaml:"fps"`
	BufSize    string `yaml:"bufSize"`
}

// Size parses Resolution, which must be formatted as WxH.
func (q Quality) Size() (w, h int, err error) {
	ws, hs, ok := strings.Cut(q.Resolution, "x")
	if !ok {
		return 0, 0, fmt.Errorf("resolution %q must be formatted as WxH", q.Resolution)
	}

	w, err = strconv.Atoi(ws)
	if err != nil || w <= 0 {
		return 0, 0, fmt.Errorf("resolution %q has an invalid width", q.Resolution)
	}
	h, err = strconv.Atoi(hs)
	if err != nil || h <= 0 {
		return 0, 0, fmt.Errorf("resolution %q has an invalid height", q.Resolution)
	}

	return w, h, nil
}

type Live struct {
	PartMinDurationMs       int `yaml:"partMinDurationMs"`
	SegmentMinDurationMs    int `yaml:"segmentMinDurationMs"`
	SegmentCount            int `yaml:"segmentCount"`
	RecordSegmentDurationMs int `yaml:"recordSegmentDurationMs"`
	EncoderQueuePackets     int `yaml:"encoderQueuePackets"`
	ThumbnailIntervalMs     int `yaml:"thumbnailIntervalMs"`
}

// WithDefaults returns a copy of l with every zero field replaced by its default.
func (l Live) WithDefaults() Live {
	if l.PartMinDurationMs == 0 {
		l.PartMinDurationMs = defaultPartMinDurationMs
	}
	if l.SegmentMinDurationMs == 0 {
		l.SegmentMinDurationMs = defaultSegmentMinDurationMs
	}
	if l.SegmentCount == 0 {
		l.SegmentCount = defaultSegmentCount
	}
	if l.RecordSegmentDurationMs == 0 {
		l.RecordSegmentDurationMs = defaultRecordSegmentDurationMs
	}
	if l.EncoderQueuePackets == 0 {
		l.EncoderQueuePackets = defaultEncoderQueuePackets
	}
	if l.ThumbnailIntervalMs == 0 {
		l.ThumbnailIntervalMs = defaultThumbnailIntervalMs
	}
	return l
}

// Validate reports the first setting the live pipeline cannot run with.
func Validate(t Transcode) error {
	if t.Live.SegmentCount < minSegmentCount {
		return fmt.Errorf("transcode.live.segmentCount must be >= %d, got %d", minSegmentCount, t.Live.SegmentCount)
	}

	for i, q := range t.FFMpegSetting.Qualities {
		if _, _, err := q.Size(); err != nil {
			return fmt.Errorf("transcode.ffmpegSetting.qualities[%d].resolution: %w", i, err)
		}
	}

	return nil
}
