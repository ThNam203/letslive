package upload

import (
	"errors"
	"math"
	"time"

	"sen1or/letslive/transcode/config"
)

const (
	// maxShortSide caps uploads (portrait included); larger files are
	// scaled down on intake.
	maxShortSide = 1080
	// maxFPS and defaultFPS bound the frame rate of converted video.
	maxFPS     = 60
	defaultFPS = 30
	// maxAudioChannels: more channels are mixed down to stereo.
	maxAudioChannels = 2
)

// ErrNoVideo is returned for a file without a video stream: the engine needs
// H.264 keyframes.
var ErrNoVideo = errors.New("the file has no video stream")

// Plan is how intake turns a file into what the live engine accepts: H.264 +
// AAC in FLV, at most 1080p short side, upright, progressive, SDR.
type Plan struct {
	CopyVideo   bool
	CopyAudio   bool
	HasAudio    bool
	Stereo      bool // mix the audio down to two channels
	ToneMap     bool // HDR to SDR BT.709
	Deinterlace bool
	ScaleShort  int    // short side to scale to; 0 keeps the size
	FPS         int    // frame rate of converted video
	MaxRate     string // bitrate cap of converted video, from the qualities
	BufSize     string // its rate-control buffer
}

// Decide picks what intake copies and what it converts. Video is copied only
// when it is H.264 8-bit 4:2:0 SDR, at most 1080p short side, not rotated,
// progressive, with no keyframe gap above maxGap, and forceConvert is false.
// Audio is copied when it is AAC with at most two channels and forceConvert
// is false. Converted video is capped at the bitrate of the configured
// quality for its size, like the qualities encoded below it; copied video
// keeps its own bitrate.
func Decide(p Probe, maxGap time.Duration, forceConvert bool, qualities []config.Quality) (Plan, error) {
	if !p.HasVideo {
		return Plan{}, ErrNoVideo
	}
	pl := Plan{
		HasAudio:    p.HasAudio,
		ToneMap:     isHDR(p.ColorTransfer),
		Deinterlace: isInterlaced(p.FieldOrder),
		FPS:         frameRate(p.FrameRate),
	}
	if min(p.Width, p.Height) > maxShortSide {
		pl.ScaleShort = maxShortSide
	}
	pl.CopyVideo = !forceConvert &&
		p.VideoCodec == "h264" &&
		(p.PixFmt == "yuv420p" || p.PixFmt == "yuvj420p") &&
		!pl.ToneMap && !pl.Deinterlace && pl.ScaleShort == 0 &&
		p.Rotation%360 == 0 &&
		p.MaxKeyframeGap <= maxGap
	if !pl.CopyVideo {
		short := min(p.Width, p.Height)
		if pl.ScaleShort > 0 {
			short = pl.ScaleShort
		}
		pl.MaxRate, pl.BufSize = rateCap(qualities, short)
	}
	if p.HasAudio {
		pl.Stereo = p.AudioChannels > maxAudioChannels
		pl.CopyAudio = !forceConvert && p.AudioCodec == "aac" && !pl.Stereo
	}
	return pl, nil
}

// rateCap is the bitrate cap of the smallest quality whose short side is at
// least short, or of the largest quality; none without qualities.
func rateCap(qualities []config.Quality, short int) (maxRate, bufSize string) {
	best, bestShort, largest, largestShort := -1, 0, -1, 0
	for i, q := range qualities {
		w, h, err := q.Size()
		if err != nil {
			continue // Validate rejects these at startup
		}
		s := min(w, h)
		if s > largestShort {
			largest, largestShort = i, s
		}
		if s >= short && (best < 0 || s < bestShort) {
			best, bestShort = i, s
		}
	}
	if best < 0 {
		best = largest
	}
	if best < 0 {
		return "", ""
	}
	return qualities[best].MaxBitrate, qualities[best].BufSize
}

func isHDR(transfer string) bool {
	return transfer == "smpte2084" || transfer == "arib-std-b67"
}

func isInterlaced(fieldOrder string) bool {
	return fieldOrder != "" && fieldOrder != "progressive" && fieldOrder != "unknown"
}

func frameRate(avg float64) int {
	if avg <= 0 {
		return defaultFPS
	}
	return min(int(math.Round(avg)), maxFPS)
}
