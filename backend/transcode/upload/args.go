package upload

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// convertedAudioBitrate is the AAC bitrate of converted audio.
	convertedAudioBitrate = "128k"
	// convertedKeyframeExpr puts a keyframe at least every 2 s of converted
	// video, so recorded segments stay close to their target length.
	convertedKeyframeExpr = "expr:gte(t,n_forced*2)"
)

// toneMapFilter maps HDR (PQ, HLG) to SDR BT.709 so it does not come out
// washed out.
const toneMapFilter = "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709," +
	"tonemap=tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv"

// scaleShortFilter pins the short side to short, after ffmpeg's auto-rotation,
// keeping the aspect ratio and an even long side.
func scaleShortFilter(short int) string {
	return fmt.Sprintf("scale='if(lt(iw,ih),%[1]d,-2)':'if(lt(iw,ih),-2,%[1]d)'", short)
}

// IntakeArgs is the ffmpeg command line that turns input into FLV on stdout
// as pl says. Rotated video is turned upright by ffmpeg's default
// auto-rotation whenever video is converted.
func IntakeArgs(input string, pl Plan, preset string, crf, threads int) []string {
	args := []string{"-hide_banner", "-loglevel", "error", "-i", input, "-map", "0:v:0"}
	if pl.HasAudio {
		args = append(args, "-map", "0:a:0")
	}

	if pl.CopyVideo {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args,
			"-vf", videoFilters(pl),
			"-c:v", "libx264", "-preset", preset, "-crf", strconv.Itoa(crf),
			"-fps_mode", "cfr", "-r", strconv.Itoa(pl.FPS),
			"-force_key_frames", convertedKeyframeExpr,
		)
	}

	switch {
	case !pl.HasAudio:
	case pl.CopyAudio:
		args = append(args, "-c:a", "copy")
	default:
		args = append(args, "-c:a", "aac", "-b:a", convertedAudioBitrate)
		if pl.Stereo {
			args = append(args, "-ac", "2")
		}
	}

	if threads > 0 {
		args = append(args, "-threads", strconv.Itoa(threads))
	}
	return append(args,
		"-avoid_negative_ts", "make_zero", "-flvflags", "no_duration_filesize", "-f", "flv", "pipe:1")
}

func videoFilters(pl Plan) string {
	var f []string
	if pl.Deinterlace {
		f = append(f, "yadif")
	}
	if pl.ToneMap {
		f = append(f, toneMapFilter)
	}
	if pl.ScaleShort > 0 {
		f = append(f, scaleShortFilter(pl.ScaleShort))
	}
	return strings.Join(append(f, "format=yuv420p"), ",")
}
