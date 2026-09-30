package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"sen1or/letslive/user/domains"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
)

const (
	MinAvatarDimension = 80
	MaxAvatarDimension = 4096
	AvatarSize         = 80

	avatarWebPQuality = 80

	// a 4096x4096 upload decodes to ~64 MB, so bound how many run at once
	maxConcurrentAvatarProcessing = 2
)

var avatarProcessingSlots = make(chan struct{}, maxConcurrentAvatarProcessing)

var avatarFormats = map[string]struct{ ext, contentType string }{
	"jpeg": {"jpg", "image/jpeg"},
	"png":  {"png", "image/png"},
	"gif":  {"gif", "image/gif"},
	"webp": {"webp", "image/webp"},
}

type processedAvatar struct {
	original            []byte
	originalExt         string
	originalContentType string
	webp                []byte
}

func checkAvatarDimensions(r io.Reader) (string, error) {
	cfg, format, err := image.DecodeConfig(r)
	if err != nil {
		return "", fmt.Errorf("read avatar header: %w: %w", domains.ErrInvalidImage, err)
	}

	if cfg.Width < MinAvatarDimension || cfg.Height < MinAvatarDimension ||
		cfg.Width > MaxAvatarDimension || cfg.Height > MaxAvatarDimension {
		return "", fmt.Errorf("avatar is %dx%d: %w", cfg.Width, cfg.Height, domains.ErrImageDimensionsOutOfRange)
	}

	return format, nil
}

func processAvatarLimited(ctx context.Context, r io.Reader) (*processedAvatar, error) {
	select {
	case avatarProcessingSlots <- struct{}{}:
	case <-ctx.Done():
		return nil, fmt.Errorf("wait for avatar processing slot: %w: %w", domains.ErrInternal, ctx.Err())
	}
	defer func() { <-avatarProcessingSlots }()

	return processAvatar(r)
}

func processAvatar(r io.Reader) (*processedAvatar, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read avatar upload: %w: %w", domains.ErrInternal, err)
	}

	format, err := checkAvatarDimensions(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	meta, ok := avatarFormats[format]
	if !ok {
		return nil, fmt.Errorf("avatar format %q: %w", format, domains.ErrInvalidImage)
	}

	img, err := decodeAvatar(data, format)
	if err != nil {
		return nil, fmt.Errorf("decode avatar: %w: %w", domains.ErrInvalidImage, err)
	}

	var served bytes.Buffer
	if err := webp.Encode(&served, img, webp.Options{Quality: avatarWebPQuality, Method: 6}); err != nil {
		return nil, fmt.Errorf("encode avatar: %w: %w", domains.ErrInternal, err)
	}

	return &processedAvatar{
		original:            data,
		originalExt:         meta.ext,
		originalContentType: meta.contentType,
		webp:                served.Bytes(),
	}, nil
}

// decodeAvatar returns the upright, center-cropped AvatarSize square.
func decodeAvatar(data []byte, format string) (*image.RGBA, error) {
	switch format {
	case "jpeg":
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		// a center square rotates onto the center square, so orienting the
		// small result is the same as orienting the full image first
		return orientImage(cropToAvatar(img), jpegOrientation(data)), nil
	case "webp":
		img, err := webp.Decode(bytes.NewReader(data), webp.Options{AutoRotate: true})
		if err != nil {
			return nil, err
		}
		return cropToAvatar(img), nil
	case "gif":
		cfg, err := gif.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		frame, err := gif.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		// the first frame may cover only part of the logical screen
		canvas := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Src)
		return cropToAvatar(canvas), nil
	default:
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return cropToAvatar(img), nil
	}
}

func cropToAvatar(src image.Image) *image.RGBA {
	b := src.Bounds()
	side := min(b.Dx(), b.Dy())
	x0 := b.Min.X + (b.Dx()-side)/2
	y0 := b.Min.Y + (b.Dy()-side)/2

	dst := image.NewRGBA(image.Rect(0, 0, AvatarSize, AvatarSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, image.Rect(x0, y0, x0+side, y0+side), draw.Src, nil)
	return dst
}
