package services

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"sen1or/letslive/user/domains"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
)

const (
	minAvatarDimension = 80
	maxAvatarDimension = 2048
	avatarSize         = 80

	avatarWebPQuality = 80
)

var avatarFormats = map[string]struct{ ext, contentType string }{
	"jpeg": {"jpg", "image/jpeg"},
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

	if cfg.Width < minAvatarDimension || cfg.Height < minAvatarDimension ||
		cfg.Width > maxAvatarDimension || cfg.Height > maxAvatarDimension {
		return "", fmt.Errorf("avatar is %dx%d: %w", cfg.Width, cfg.Height, domains.ErrImageDimensionsOutOfRange)
	}

	return format, nil
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

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode avatar: %w: %w", domains.ErrInvalidImage, err)
	}

	var served bytes.Buffer
	if err := webp.Encode(&served, scaleCenterSquare(img), webp.Options{Quality: avatarWebPQuality, Method: 6}); err != nil {
		return nil, fmt.Errorf("encode avatar: %w: %w", domains.ErrInternal, err)
	}

	return &processedAvatar{
		original:            data,
		originalExt:         meta.ext,
		originalContentType: meta.contentType,
		webp:                served.Bytes(),
	}, nil
}

func scaleCenterSquare(img image.Image) *image.RGBA {
	b := img.Bounds()
	side := min(b.Dx(), b.Dy())
	x0 := b.Min.X + (b.Dx()-side)/2
	y0 := b.Min.Y + (b.Dy()-side)/2

	dst := image.NewRGBA(image.Rect(0, 0, avatarSize, avatarSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, image.Rect(x0, y0, x0+side, y0+side), draw.Src, nil)
	return dst
}
