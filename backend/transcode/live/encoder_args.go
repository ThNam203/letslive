package live

import (
	"fmt"
	"strconv"
	"strings"

	"sen1or/letslive/transcode/config"
)

const (
	// First extra file descriptor handed to ffmpeg; ExtraFiles[0] becomes fd 3.
	firstOutputFD = 3

	// Upper bound on GOP length; real keyframes follow the source via
	// -force_key_frames source, so this only caps pathological sources.
	maxGOPFrames = 300
)

// EncoderProfile holds the ffmpeg settings in which a livestream and an
// upload differ; the ladder rules are the same for both.
type EncoderProfile struct {
	Tune   []string // x264 tuning, per output
	Input  []string // input probing flags
	Output []string // muxer flags, per output
	// Threads caps each output's encoder threads; 0 lets ffmpeg decide.
	Threads int
	// Niceness is ffmpeg's nice value; 0 leaves it unchanged.
	Niceness int
}

// liveProfile keeps latency low: no lookahead, no probing, every packet
// flushed at once.
var liveProfile = EncoderProfile{
	Tune:   []string{"-tune", "zerolatency"},
	Input:  []string{"-probesize", "32", "-analyzeduration", "0"},
	Output: []string{"-muxdelay", "0", "-muxpreload", "0", "-flush_packets", "1"},
}

// fileProfile uses normal x264 tuning, so an upload's files are smaller at the
// same quality, and runs niced beside live encoders.
func fileProfile(u config.Upload) EncoderProfile {
	return EncoderProfile{
		Output:   []string{"-muxdelay", "0", "-muxpreload", "0"},
		Threads:  u.Threads,
		Niceness: u.Niceness,
	}
}

// EncoderArgs builds the ffmpeg command line that reads FLV from stdin and
// writes one video-only MPEG-TS stream per rendition to pipe:3, pipe:4, ...
// encoded must hold only the renditions with Encoded set, in ladder order.
func EncoderArgs(src SourceInfo, encoded []Rendition, preset string, crf int, p EncoderProfile) []string {
	args := []string{"-hide_banner", "-loglevel", "error"}
	args = append(args, p.Input...)
	args = append(args,
		"-f", "flv", "-i", "pipe:0",
		"-an", "-copyts",
		"-filter_complex", filterGraph(src, encoded),
	)

	for i, q := range encoded {
		args = append(args,
			"-map", fmt.Sprintf("[v%d]", i),
			"-c:v", "libx264",
			"-preset", preset,
		)
		args = append(args, p.Tune...)
		args = append(args,
			"-pix_fmt", "yuv420p",
			"-crf", strconv.Itoa(crf),
			"-maxrate", q.MaxBitrate,
			"-bufsize", q.BufSize,
			"-force_key_frames", "source",
			"-sc_threshold", "0",
			"-g", strconv.Itoa(maxGOPFrames),
		)
		// q.FPS is already capped at the rounded source rate. It is passed even
		// when equal: the FLV from the session carries no frame rate, and
		// ffmpeg's guess (60 for 30 fps) would end up in x264's VUI timing and
		// in the advertised FRAME-RATE.
		if q.FPS > 0 && src.FPS > 0 {
			args = append(args, "-r", strconv.Itoa(q.FPS))
		}
		if p.Threads > 0 {
			args = append(args, "-threads", strconv.Itoa(p.Threads))
		}
		args = append(args, p.Output...)
		args = append(args, "-f", "mpegts", fmt.Sprintf("pipe:%d", firstOutputFD+i))
	}

	return args
}

// filterGraph scales the source once per rendition. The short side is pinned
// and the long side follows (-2 keeps it even), so portrait stays portrait.
func filterGraph(src SourceInfo, encoded []Rendition) string {
	scale := func(short int) string {
		if src.Portrait() {
			return fmt.Sprintf("scale=%d:-2", short)
		}
		return fmt.Sprintf("scale=-2:%d", short)
	}

	if len(encoded) == 1 {
		return fmt.Sprintf("[0:v]%s[v0]", scale(encoded[0].ShortSide))
	}

	var graph strings.Builder
	fmt.Fprintf(&graph, "[0:v]split=%d", len(encoded))
	for i := range encoded {
		fmt.Fprintf(&graph, "[s%d]", i)
	}
	for i, q := range encoded {
		fmt.Fprintf(&graph, ";[s%d]%s[v%d]", i, scale(q.ShortSide), i)
	}
	return graph.String()
}
