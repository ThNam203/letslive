package transcoder

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/config"
	"strings"
)

type Transcoder struct {
	stdin       *io.PipeReader
	commandExec *exec.Cmd
	config      config.Config
}

func NewTranscoder(pipeOut *io.PipeReader, config config.Config) *Transcoder {
	return &Transcoder{
		stdin:  pipeOut,
		config: config,
	}
}

func (t *Transcoder) Start(publishName string) {
	// if there is no remote (or external storage), just export files directly to public folder and serves
	var outputDir = filepath.Join(t.config.Transcode.PublicHLSPath, publishName)
	if t.config.IPFS.Enabled {
		outputDir = filepath.Join(t.config.Transcode.PrivateHLSPath, publishName)
	}

	var videoMaps = make([]string, 0)
	var audioMaps = make([]string, 0)
	var streamMaps = make([]string, 0)

	for index, quality := range t.config.Transcode.FFMpegSetting.Qualities {
		var g = quality.FPS * t.config.Transcode.FFMpegSetting.HLSTime
		var keyint_min = t.config.Transcode.FFMpegSetting.HLSTime

		videoMaps = append(videoMaps, fmt.Sprintf("-map v:0 -s:%v %s -r:%v %v -maxrate:%v %s -bufsize:%v %s -g:%v %v -keyint_min:%v %v", index, quality.Resolution, index, quality.FPS, index, quality.MaxBitrate, index, quality.BufSize, index, g, index, keyint_min))
		audioMaps = append(audioMaps, "-map a:0")
		streamMaps = append(streamMaps, fmt.Sprintf("v:%v,a:%v", index, index))
	}

	var ffmpegFlags = []string{
		"-hide_banner",
		"-re",
		"-i pipe:0",
		fmt.Sprintf("-preset %s", t.config.Transcode.FFMpegSetting.Preset),
		"-sc_threshold 0",
		"-c:v libx264",
		"-pix_fmt yuv420p",
		fmt.Sprintf("-crf %v", t.config.Transcode.FFMpegSetting.CRF),
		strings.Join(videoMaps, " "),
		strings.Join(audioMaps, " "),
		"-c:a aac",
		"-b:a 128k",
		"-ac 1",
		"-ar 44100",
		"-f hls",
		fmt.Sprintf("-hls_time %v", t.config.Transcode.FFMpegSetting.HLSTime),
		fmt.Sprintf("-hls_delete_threshold %v", t.config.Transcode.FFMpegSetting.HlsMaxSize-t.config.Transcode.FFMpegSetting.HlsListSize),
		fmt.Sprintf("-hls_list_size %v", t.config.Transcode.FFMpegSetting.HlsListSize),
		"-hls_flags delete_segments",
		fmt.Sprintf("-master_pl_name %s", t.config.Transcode.FFMpegSetting.MasterFileName),
		fmt.Sprintf(`-var_stream_map "%s"`, strings.Join(streamMaps, " ")),
		filepath.Join(outputDir, "/%v/stream.m3u8"),
	}

	ffmpegFlagsString := strings.Join(ffmpegFlags, " ")
	ffmpegCommand := t.config.Transcode.FFMpegSetting.FFMpegPath + " " + ffmpegFlagsString

	t.commandExec = exec.Command("sh", "-c", ffmpegCommand)
	t.commandExec.Stdin = t.stdin

	// TODO: logs out error
	// stderr, err := execCommand.StderrPipe()

	if err := t.commandExec.Start(); err != nil {
		logger.Errorf("error while starting ffmpeg command: %s", err)
	}

	err := t.commandExec.Wait()
	if err != nil {
		logger.Errorf("ffmpeg failed: %s", err)
	}
}

func (t *Transcoder) Stop() {
	err := t.commandExec.Process.Kill()
	if err != nil {
		logger.Errorf("transcoder error while being killed: %s", err)
	}
}
