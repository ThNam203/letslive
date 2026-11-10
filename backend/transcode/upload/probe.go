// Package upload turns an uploaded video file into a VOD through the live
// engine (live.FileMode): the file is probed, made into an H.264/AAC FLV
// stream by one ffmpeg, and fed to a session like an RTMP connection.
package upload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Probe is what Decide needs to know about an input file: its first video and
// first audio stream.
type Probe struct {
	HasVideo       bool
	VideoCodec     string
	PixFmt         string
	FieldOrder     string
	ColorTransfer  string
	Width, Height  int
	Rotation       int     // degrees from the display matrix or rotate tag
	FrameRate      float64 // average; 0 when unknown
	MaxKeyframeGap time.Duration
	Duration       time.Duration

	HasAudio      bool
	AudioCodec    string
	AudioChannels int
}

// ffprobePath is the ffprobe next to ffmpegPath, or "ffprobe" on PATH when
// ffmpegPath has no directory.
func ffprobePath(ffmpegPath string) string {
	dir := filepath.Dir(ffmpegPath)
	if dir == "." && !strings.ContainsRune(ffmpegPath, filepath.Separator) {
		return "ffprobe"
	}
	return filepath.Join(dir, "ffprobe")
}

// RunProbe probes input with the ffprobe next to ffmpegPath: its streams, and
// then the keyframe times of a video stream (a fast scan of the whole file).
func RunProbe(ctx context.Context, ffmpegPath, input string) (Probe, error) {
	probe := ffprobePath(ffmpegPath)
	out, err := runFFprobe(ctx, probe, "-show_streams", "-show_format", input)
	if err != nil {
		return Probe{}, err
	}
	p, err := parseStreams(out)
	if err != nil {
		return Probe{}, err
	}
	if !p.HasVideo {
		return p, nil
	}
	out, err = runFFprobe(ctx, probe, "-select_streams", "v:0", "-skip_frame", "nokey",
		"-show_entries", "frame=pts_time", input)
	if err != nil {
		return Probe{}, err
	}
	if p.MaxKeyframeGap, err = maxKeyframeGap(out, p.Duration); err != nil {
		return Probe{}, err
	}
	return p, nil
}

func runFFprobe(ctx context.Context, ffprobe string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, ffprobe, append([]string{"-v", "error", "-of", "json"}, args...)...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w (stderr: %q)", err, tail(stderr.String()))
	}
	return out, nil
}

type probeJSON struct {
	Streams []struct {
		CodecType     string            `json:"codec_type"`
		CodecName     string            `json:"codec_name"`
		PixFmt        string            `json:"pix_fmt"`
		FieldOrder    string            `json:"field_order"`
		ColorTransfer string            `json:"color_transfer"`
		Width         int               `json:"width"`
		Height        int               `json:"height"`
		AvgFrameRate  string            `json:"avg_frame_rate"`
		Channels      int               `json:"channels"`
		Tags          map[string]string `json:"tags"`
		SideData      []struct {
			Rotation *float64 `json:"rotation"`
		} `json:"side_data_list"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// parseStreams reads ffprobe's -show_streams -show_format JSON.
func parseStreams(data []byte) (Probe, error) {
	var j probeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return Probe{}, fmt.Errorf("parse ffprobe streams: %w", err)
	}
	var p Probe
	if d, err := strconv.ParseFloat(j.Format.Duration, 64); err == nil {
		p.Duration = time.Duration(d * float64(time.Second))
	}
	for _, s := range j.Streams {
		switch {
		case s.CodecType == "video" && !p.HasVideo:
			p.HasVideo = true
			p.VideoCodec, p.PixFmt, p.FieldOrder, p.ColorTransfer = s.CodecName, s.PixFmt, s.FieldOrder, s.ColorTransfer
			p.Width, p.Height = s.Width, s.Height
			p.FrameRate = parseRate(s.AvgFrameRate)
			if r, err := strconv.Atoi(s.Tags["rotate"]); err == nil {
				p.Rotation = r
			}
			for _, sd := range s.SideData {
				if sd.Rotation != nil {
					p.Rotation = int(math.Round(*sd.Rotation))
				}
			}
		case s.CodecType == "audio" && !p.HasAudio:
			p.HasAudio = true
			p.AudioCodec, p.AudioChannels = s.CodecName, s.Channels
		}
	}
	return p, nil
}

// parseRate reads "num/den"; 0 when unknown.
func parseRate(r string) float64 {
	num, den, ok := strings.Cut(r, "/")
	if !ok {
		return 0
	}
	n, err1 := strconv.ParseFloat(num, 64)
	d, err2 := strconv.ParseFloat(den, 64)
	if err1 != nil || err2 != nil || d == 0 {
		return 0
	}
	return n / d
}

// maxKeyframeGap reads ffprobe's keyframe pts_time list and returns the
// longest gap between keyframes, counting the tail up to duration.
func maxKeyframeGap(data []byte, duration time.Duration) (time.Duration, error) {
	var j struct {
		Frames []struct {
			PTSTime string `json:"pts_time"`
		} `json:"frames"`
	}
	if err := json.Unmarshal(data, &j); err != nil {
		return 0, fmt.Errorf("parse ffprobe keyframes: %w", err)
	}
	var times []time.Duration
	for _, f := range j.Frames {
		s, err := strconv.ParseFloat(f.PTSTime, 64)
		if err != nil {
			continue // a keyframe without a timestamp tells nothing
		}
		times = append(times, time.Duration(s*float64(time.Second)))
	}
	if len(times) == 0 {
		return 0, errors.New("the video stream has no keyframe")
	}
	var gap time.Duration
	for i := 1; i < len(times); i++ {
		gap = max(gap, times[i]-times[i-1])
	}
	return max(gap, duration-times[len(times)-1]), nil
}

// tail keeps the end of a long ffmpeg/ffprobe stderr for an error message.
func tail(s string) string {
	const keep = 4096
	s = strings.TrimSpace(s)
	if len(s) > keep {
		return "…" + s[len(s)-keep:]
	}
	return s
}
