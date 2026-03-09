package transcoder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sen1or/letslive/transcode/config"
	"sen1or/letslive/transcode/pkg/logger"
	"strings"
)

// TranscodeFile transcodes a video file to HLS format (blocking).
// inputPath: path to the raw video file on disk.
// outputDir: directory where HLS segments and playlists will be written.
// Returns the path to the master playlist, thumbnail path, and any error.
func TranscodeFile(ctx context.Context, cfg config.Transcode, inputPath string, outputDir string) (string, string, error) {
	// Create output directory structure for each quality
	for i := range cfg.FFMpegSetting.Qualities {
		qualityDir := filepath.Join(outputDir, fmt.Sprintf("%d", i))
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create quality dir %s: %w", qualityDir, err)
		}
	}

	var videoMaps []string
	var audioMaps []string
	var streamMaps []string

	for index, quality := range cfg.FFMpegSetting.Qualities {
		keyint := quality.FPS * cfg.FFMpegSetting.HLSTime

		videoMaps = append(videoMaps, fmt.Sprintf("-map v:0 -s:%v %s -r:%v %v -maxrate:%v %s -bufsize:%v %s -g:%v %v -keyint_min:%v %v",
			index, quality.Resolution, index, quality.FPS, index, quality.MaxBitrate, index, quality.BufSize, index, keyint, index, keyint))
		audioMaps = append(audioMaps, "-map a:0")
		streamMaps = append(streamMaps, fmt.Sprintf("v:%v,a:%v", index, index))
	}

	args := []string{
		"-hide_banner",
		"-y",
		"-i", inputPath,
		"-sc_threshold", "0",
		"-preset", cfg.FFMpegSetting.Preset,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-crf", fmt.Sprintf("%v", cfg.FFMpegSetting.CRF),
	}
	args = append(args, strings.Fields(strings.Join(videoMaps, " "))...)
	args = append(args, strings.Fields(strings.Join(audioMaps, " "))...)
	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "1",
		"-ar", "44100",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%v", cfg.FFMpegSetting.HLSTime),
		"-hls_list_size", "0", // keep all segments for VOD
		"-hls_flags", "independent_segments",
		"-master_pl_name", cfg.FFMpegSetting.MasterFileName,
		"-var_stream_map", strings.Join(streamMaps, " "),
		filepath.Join(outputDir, "%v", "stream.m3u8"),
	)

	cmd := exec.CommandContext(ctx, cfg.FFMpegSetting.FFMpegPath, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	logger.Infof(ctx, "starting file transcode: %s", inputPath)

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Log stderr in background
	go func() {
		scanner := bufio.NewReader(stderr)
		for {
			line, readErr := scanner.ReadString('\n')
			if line != "" {
				line = strings.TrimSuffix(line, "\n")
				logger.Infof(ctx, "ffmpeg file transcode: %s", line)
			}
			if readErr != nil {
				if readErr != io.EOF {
					logger.Errorf(ctx, "ffmpeg stderr read error: %v", readErr)
				}
				break
			}
		}
	}()

	// Wait for FFmpeg to complete (blocking)
	if err := cmd.Wait(); err != nil {
		return "", "", fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	masterPlaylist := filepath.Join(outputDir, cfg.FFMpegSetting.MasterFileName)

	// Generate thumbnail from the file
	thumbnailPath := filepath.Join(outputDir, "thumbnail.jpg")
	generateThumbnailFromFile(ctx, cfg.FFMpegSetting.FFMpegPath, inputPath, thumbnailPath)

	logger.Infof(ctx, "file transcode completed successfully: %s", inputPath)
	return masterPlaylist, thumbnailPath, nil
}

func generateThumbnailFromFile(ctx context.Context, ffmpegPath, inputPath, outputPath string) {
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", inputPath,
		"-vf", "select='eq(pict_type\\,I)*gte(t\\,1)',scale=640:-1",
		"-frames:v", "1",
		"-q:v", "2",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		logger.Errorf(ctx, "thumbnail generation failed for %s: %v", inputPath, err)
	}
}
