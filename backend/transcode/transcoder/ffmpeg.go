package transcoder

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sen1or/letslive/transcode/config"
	"sen1or/letslive/transcode/pkg/logger"
	"strings"
)

type Transcoder struct {
	stdin       *io.PipeReader
	commandExec *exec.Cmd
	config      config.Transcode
	onStart     func()
}

func NewTranscoder(pipeOut *io.PipeReader, config config.Transcode, onStart func()) *Transcoder {
	return &Transcoder{
		stdin:   pipeOut,
		config:  config,
		onStart: onStart,
	}
}

func (t *Transcoder) Start(publishName string) {
	// if there is no remote (or external storage), just export files directly to public folder and serves
	var outputDir = filepath.Join(t.config.PrivateHLSPath, publishName)

	var videoMaps = make([]string, 0)
	var audioMaps = make([]string, 0)
	var streamMaps = make([]string, 0)

	for index, quality := range t.config.FFMpegSetting.Qualities {
		var g = quality.FPS * t.config.FFMpegSetting.HLSTime
		var keyint_min = t.config.FFMpegSetting.HLSTime

		videoMaps = append(videoMaps, fmt.Sprintf("-map v:0 -s:%v %s -r:%v %v -maxrate:%v %s -bufsize:%v %s -g:%v %v -keyint_min:%v %v", index, quality.Resolution, index, quality.FPS, index, quality.MaxBitrate, index, quality.BufSize, index, g, index, keyint_min))
		audioMaps = append(audioMaps, "-map a:0")
		streamMaps = append(streamMaps, fmt.Sprintf("v:%v,a:%v", index, index))
	}

	var ffmpegFlags = []string{
		"-hide_banner",
		"-re",
		"-i pipe:0",
		fmt.Sprintf("-preset %s", t.config.FFMpegSetting.Preset),
		"-sc_threshold 0",
		"-c:v libx264",
		"-pix_fmt yuv420p",
		fmt.Sprintf("-crf %v", t.config.FFMpegSetting.CRF),
		strings.Join(videoMaps, " "),
		strings.Join(audioMaps, " "),
		"-c:a aac",
		"-b:a 128k",
		"-ac 1",
		"-ar 44100",
		"-f hls",
		fmt.Sprintf("-hls_time %v", t.config.FFMpegSetting.HLSTime),
		fmt.Sprintf("-hls_delete_threshold %v", t.config.FFMpegSetting.HlsMaxSize-t.config.FFMpegSetting.HlsListSize),
		fmt.Sprintf("-hls_list_size %v", t.config.FFMpegSetting.HlsListSize),
		"-hls_flags delete_segments",
		fmt.Sprintf("-master_pl_name %s", t.config.FFMpegSetting.MasterFileName),
		fmt.Sprintf(`-var_stream_map "%s"`, strings.Join(streamMaps, " ")),
		filepath.Join(outputDir, "/%v/stream.m3u8"),
		"-vf fps=1/60 -update 1",
		filepath.Join(outputDir, "thumbnail.jpeg"),
	}

	ffmpegFlagsString := strings.Join(ffmpegFlags, " ")
	ffmpegCommand := t.config.FFMpegSetting.FFMpegPath + " " + ffmpegFlagsString

	t.commandExec = exec.Command("sh", "-c", ffmpegCommand)
	t.commandExec.Stdin = t.stdin

	stderr, err := t.commandExec.StderrPipe()
	if err != nil {
		logger.Errorf("failed to get stderr pipe up: %v", err)
	}

	if err := t.commandExec.Start(); err != nil {
		logger.Errorf("error while starting ffmpeg command: %s", err)
	}

	t.onStart()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			//logger.Warnf("ffmpeg error pipe: %v", scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			logger.Errorf("error in reading stderr pipe: %v", err)
		}
	}()

	err = t.commandExec.Wait()
	if err != nil {
		logger.Errorf("ffmpeg failed: %s", err)
		t.Stop()
	}
}

func (t *Transcoder) Stop() {
	err := t.commandExec.Process.Kill()
	if err != nil {
		logger.Errorf("transcoder error while being killed: %s", err)
	}
}
