package transcoder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sen1or/letslive/transcode/config"
	"sen1or/letslive/transcode/pkg/logger"
	"strings"
	"syscall"
)

type Transcoder struct {
	inputPipe   *io.PipeReader
	commandExec *exec.Cmd
	config      config.Transcode
	onStart     func()
}

func NewTranscoder(inputPipe *io.PipeReader, config config.Transcode, onStart func()) *Transcoder {
	return &Transcoder{
		inputPipe: inputPipe,
		config:    config,
		onStart:   onStart,
	}
}

func (t *Transcoder) Start(ctx context.Context, publishName string) {
	// if there is no remote (or external storage), just export files directly to public folder and serves
	var outputDir = filepath.Join(t.config.PrivateHLSPath, publishName)

	var videoMaps = make([]string, 0)
	var audioMaps = make([]string, 0)
	var streamMaps = make([]string, 0)

	for index, quality := range t.config.FFMpegSetting.Qualities {
		// GOP/keyframe should be keyint_min == g == fps * hls_time
		keyint := quality.FPS * t.config.FFMpegSetting.HLSTime

		videoMaps = append(videoMaps, fmt.Sprintf("-map v:0 -s:%v %s -r:%v %v -maxrate:%v %s -bufsize:%v %s -g:%v %v -keyint_min:%v %v", index, quality.Resolution, index, quality.FPS, index, quality.MaxBitrate, index, quality.BufSize, index, keyint, index, keyint))
		audioMaps = append(audioMaps, "-map a:0")
		streamMaps = append(streamMaps, fmt.Sprintf("v:%v,a:%v", index, index))
	}

	args := []string{
		"-hide_banner",
		"-i", "pipe:0",
		"-sc_threshold", "0",
		"-preset", t.config.FFMpegSetting.Preset,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-crf", fmt.Sprintf("%v", t.config.FFMpegSetting.CRF),
	}
	args = append(args, strings.Fields(strings.Join(videoMaps, " "))...)
	args = append(args, strings.Fields(strings.Join(audioMaps, " "))...)
	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "1",
		"-ar", "44100",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%v", t.config.FFMpegSetting.HLSTime),
		"-hls_delete_threshold", fmt.Sprintf("%v", t.config.FFMpegSetting.HlsMaxSize-t.config.FFMpegSetting.HlsListSize),
		"-hls_list_size", fmt.Sprintf("%v", t.config.FFMpegSetting.HlsListSize),
		"-hls_flags", "delete_segments",
		"-master_pl_name", t.config.FFMpegSetting.MasterFileName,
		"-var_stream_map", strings.Join(streamMaps, " "),
		filepath.Join(outputDir, "%v", "stream.m3u8"),
	)

	t.commandExec = exec.CommandContext(ctx, t.config.FFMpegSetting.FFMpegPath, args...)

	// clone into 2 pipes, one for stream, one for thumbnail
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	go func() {
		defer w1.Close()
		defer w2.Close()
		io.Copy(io.MultiWriter(w1, w2), t.inputPipe)
	}()

	t.commandExec.Stdin = r1
	go generateThumbnail(ctx, t.config.FFMpegSetting.FFMpegPath, outputDir, r2)

	stderr, err := t.commandExec.StderrPipe()
	if err != nil {
		logger.Errorf(ctx, "failed to get stderr pipe up: %v", err)
		return
	}

	if err := t.commandExec.Start(); err != nil {
		logger.Errorf(ctx, "error while starting ffmpeg command: %s", err)
		return
	}

	if t.onStart != nil {
		t.onStart()
	}

	go func() {
		scanner := bufio.NewReader(stderr)
		for {
			line, err := scanner.ReadString('\n')
			if line != "" {
				line = strings.TrimSuffix(line, "\n")
				logger.Infof(ctx, "ffmpeg stderr: %s", line)
			}
			if err != nil {
				if err != io.EOF {
					logger.Errorf(ctx, "ffmpeg stderr read error: %v", err)
				}
				break
			}
		}
	}()

	go func() {
		if err := t.commandExec.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				logger.Errorf(ctx, "ffmpeg process failed (exit code %d): %v", exitErr.ExitCode(), err)
			} else {
				logger.Errorf(ctx, "ffmpeg process failed: %v", err)
			}
		} else {
			logger.Infof(ctx, "ffmpeg process exited successfully")
		}
	}()
}

func (t *Transcoder) Stop(ctx context.Context) {
	if t.commandExec == nil || t.commandExec.Process == nil {
		return
	}

	err := t.commandExec.Process.Signal(syscall.SIGTERM)
	if err != nil {
		logger.Errorf(ctx, "transcoder error while terminating: %s", err)
	}
}

func generateThumbnail(ctx context.Context, ffmpegPath, outputDir string, inputStream io.Reader) {
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-i", "pipe:0",
		"-map", "v:0",
		"-vf", "select='eq(pict_type\\,I)*gte(t\\,5)',fps=1/60,scale=640:-1", // take only I frame, delay start 5 second, generate per 60 frames
		"-q:v", "2", // image decoder quality
		"-update", "1", // just one image instead of multiple
		filepath.Join(outputDir, "thumbnail.jpg"),
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.Stdin = inputStream

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Errorf(ctx, "failed to get thumbnail stderr pipe up: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		logger.Errorf(ctx, "error while starting thumnail command: %s", err)
		return
	}

	go func() {
		scanner := bufio.NewReader(stderr)
		for {
			_, err := scanner.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					logger.Errorf(ctx, "stderr read error: %v", err)
				}
				break
			}
		}
	}()

	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Errorf(ctx, "ffmpeg failed: %s", err)
		}
	}()
}
