package transcoder

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sen1or/lets-live/core/config"
	"sen1or/lets-live/core/logger"
	"strings"
)

var configuration = config.GetConfig()

type Transcoder struct {
	stdin       *io.PipeReader
	commandExec *exec.Cmd
}

func NewTranscoder(pipeOut *io.PipeReader) *Transcoder {
	return &Transcoder{
		stdin: pipeOut,
	}
}

func (t *Transcoder) Start(publishName string) {
	var outputDir = filepath.Join(configuration.PrivateHLSPath, publishName)

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
		"-i pipe:0",
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
	ffmpegCommand := configuration.FFMpegSetting.FFMpegPath + " " + ffmpegFlagsString

	t.commandExec = exec.Command("sh", "-c", ffmpegCommand)
	t.commandExec.Stdin = t.stdin

	// TODO: logs out error
	// stderr, err := execCommand.StderrPipe()

	if err := t.commandExec.Start(); err != nil {
		logger.Errorf("error while starting ffmpeg command: %s", err)
	}

	err := t.commandExec.Wait()
	if err != nil {
		logger.Debugf("ffmpeg failed: %s", err)
	}
}

func (t *Transcoder) Stop() {
	err := t.commandExec.Process.Kill()
	if err != nil {
		logger.Debugf("transcoder is killed: %s", err)
	}
}
