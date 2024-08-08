package rtmpserver

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sen1or/lets-live/config"
	"strings"
	"time"
)

func startFfmpeg(pipePath string) {
	var configuration = config.GetConfig()

	var outputDir = configuration.PublicHLSPath

	var videoMaps = make([]string, 0)
	var audioMaps = make([]string, 0)
	var streamMaps = make([]string, 0)

	for index, quality := range configuration.FFMpegSetting.Qualities {
		var g = quality.FPS * configuration.FFMpegSetting.HlsTime
		var keyint_min = configuration.FFMpegSetting.HlsTime

		videoMaps = append(videoMaps, fmt.Sprintf("-map v:0 -s:%v %s -r:%v %v -maxrate:%v %s -bufsize:%v %s -g:%v %v -keyint_min:%v %v", index, quality.Resolution, index, quality.FPS, index, quality.MaxBitrate, index, quality.BufSize, index, g, index, keyint_min))
		audioMaps = append(audioMaps, "-map a:0")
		streamMaps = append(streamMaps, fmt.Sprintf("v:%v,a:%v", index, index))
	}

	var ffmpegFlags = []string{
		"-hide_banner",
		"-re",
		"-i pipe:",
		fmt.Sprintf("-preset %s", configuration.FFMpegSetting.Preset),
		"-sc_threshold 0",
		"-c:v libx264",
		"-pix_fmt yuv420p",
		fmt.Sprintf("-crf %v", configuration.FFMpegSetting.Crf),
		strings.Join(videoMaps, " "),
		strings.Join(audioMaps, " "),
		"-c:a aac",
		"-b:a 128k",
		"-ac 1",
		"-ar 44100",
		"-f hls",
		fmt.Sprintf("-hls_time %v", configuration.FFMpegSetting.HlsTime),
		fmt.Sprintf("-hls_delete_threshold %v", configuration.FFMpegSetting.HlsMaxSize-configuration.FFMpegSetting.HlsListSize),
		fmt.Sprintf("-hls_list_size %v", configuration.FFMpegSetting.HlsListSize),
		"-hls_flags delete_segments",
		fmt.Sprintf("-master_pl_name %s", configuration.FFMpegSetting.MasterFileName),
		fmt.Sprintf(`-var_stream_map "%s"`, strings.Join(streamMaps, " ")),
		filepath.Join(outputDir, "/%v/stream.m3u8"),
	}

	ffmpegFlagsString := strings.Join(ffmpegFlags, " ")
	ffmpegCommand := "cat " + pipePath + " | " + configuration.FFMpegSetting.FFMpegPath + " " + ffmpegFlagsString

	// TODO: implements a function to check if the file has been created or not
	time.Sleep(4 * time.Second)
	_, err := exec.Command("sh", "-c", ffmpegCommand).Output()
	if err != nil {
		log.Panic(err)
	}
}
