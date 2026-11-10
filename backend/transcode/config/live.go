package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultPartMinDurationMs       = 190
	defaultSegmentMinDurationMs    = 1000
	defaultSegmentCount            = 7
	defaultRecordSegmentDurationMs = 4000
	defaultEncoderQueuePackets     = 512
	defaultThumbnailIntervalMs     = 60000
	defaultUploadNiceness          = 10

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

// Upload tunes how uploaded videos are encoded beside live streams.
type Upload struct {
	// Niceness is the nice value of an upload's ffmpeg processes, so live
	// encoders win the CPU. 0 means the default (10).
	Niceness int `yaml:"niceness"`
	// Threads caps each ffmpeg output's encoder threads; 0 lets ffmpeg decide.
	Threads int `yaml:"threads"`
}

// WithDefaults returns a copy of u with every zero field replaced by its default.
func (u Upload) WithDefaults() Upload {
	if u.Niceness == 0 {
		u.Niceness = defaultUploadNiceness
	}
	return u
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
