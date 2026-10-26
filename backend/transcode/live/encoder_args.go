package live

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// First extra file descriptor handed to ffmpeg; ExtraFiles[0] becomes fd 3.
	firstOutputFD = 3

	// Upper bound on GOP length; real keyframes follow the source via
	// -force_key_frames source, so this only caps pathological sources.
	maxGOPFrames = 300
)

// EncoderArgs builds the ffmpeg command line that reads FLV from stdin and
// writes one video-only MPEG-TS stream per rendition to pipe:3, pipe:4, ...
// encoded must hold only the renditions with Encoded set, in ladder order.
func EncoderArgs(src SourceInfo, encoded []Rendition, preset string, crf int) []string {
	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-probesize", "32", "-analyzeduration", "0",
		"-f", "flv", "-i", "pipe:0",
		"-an", "-copyts",
		"-filter_complex", filterGraph(src, encoded),
	}

	for i, q := range encoded {
		args = append(args,
			"-map", fmt.Sprintf("[v%d]", i),
			"-c:v", "libx264",
			"-preset", preset,
			"-tune", "zerolatency",
			"-pix_fmt", "yuv420p",
			"-crf", strconv.Itoa(crf),
			"-maxrate", q.MaxBitrate,
			"-bufsize", q.BufSize,
			"-force_key_frames", "source",
			"-sc_threshold", "0",
			"-g", strconv.Itoa(maxGOPFrames),
		)
		if q.FPS > 0 && src.FPS > 0 && float64(q.FPS) < src.FPS {
			args = append(args, "-r", strconv.Itoa(q.FPS))
		}
		args = append(args,
			"-muxdelay", "0",
			"-muxpreload", "0",
			"-flush_packets", "1",
			"-f", "mpegts", fmt.Sprintf("pipe:%d", firstOutputFD+i),
		)
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
