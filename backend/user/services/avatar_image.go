package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sen1or/letslive/user/domains"

	"github.com/gen2brain/webp"
	"golang.org/x/image/draw"
	"golang.org/x/sync/semaphore"
)

const (
	minAvatarDimension = 80
	maxAvatarDimension = 10000
	avatarSize         = 80

	avatarWebPQuality = 80

	// a 10000x10000 upload decodes to 400-800 MB, so concurrent decodes share
	// one budget and the largest images run alone
	avatarDecodeBudgetBytes = 512 << 20
	avatarDecodeBytesPerPx  = 8
)

var avatarDecodeBudget = semaphore.NewWeighted(avatarDecodeBudgetBytes)

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

func checkAvatarDimensions(r io.Reader) (image.Config, string, error) {
	cfg, format, err := image.DecodeConfig(r)
	if err != nil {
		return image.Config{}, "", fmt.Errorf("read avatar header: %w: %w", domains.ErrInvalidImage, err)
	}

	if cfg.Width < minAvatarDimension || cfg.Height < minAvatarDimension ||
		cfg.Width > maxAvatarDimension || cfg.Height > maxAvatarDimension {
		return image.Config{}, "", fmt.Errorf("avatar is %dx%d: %w", cfg.Width, cfg.Height, domains.ErrImageDimensionsOutOfRange)
	}

	return cfg, format, nil
}

func avatarDecodeCost(cfg image.Config) int64 {
	return min(int64(cfg.Width)*int64(cfg.Height)*avatarDecodeBytesPerPx, avatarDecodeBudgetBytes)
}

func processAvatar(ctx context.Context, r io.Reader) (*processedAvatar, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read avatar upload: %w: %w", domains.ErrInternal, err)
	}

	cfg, format, err := checkAvatarDimensions(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	meta, ok := avatarFormats[format]
	if !ok {
		return nil, fmt.Errorf("avatar format %q: %w", format, domains.ErrInvalidImage)
	}

	cost := avatarDecodeCost(cfg)
	if err := avatarDecodeBudget.Acquire(ctx, cost); err != nil {
		return nil, fmt.Errorf("wait for avatar decode budget: %w: %w", domains.ErrInternal, err)
	}
	defer avatarDecodeBudget.Release(cost)

	served, err := encodeAvatar(data, format, cfg)
	if err != nil {
		return nil, err
	}

	return &processedAvatar{
		original:            data,
		originalExt:         meta.ext,
		originalContentType: meta.contentType,
		webp:                served,
	}, nil
}

func encodeAvatar(data []byte, format string, cfg image.Config) ([]byte, error) {
	img, err := decodeAvatar(data, format, cfg)
	if err != nil {
		return nil, fmt.Errorf("decode avatar: %w: %w", domains.ErrInvalidImage, err)
	}

	square := renderAvatar(img, avatarOrientation(data, format))

	var served bytes.Buffer
	if err := webp.Encode(&served, square, webp.Options{Quality: avatarWebPQuality, Method: 6}); err != nil {
		return nil, fmt.Errorf("encode avatar: %w: %w", domains.ErrInternal, err)
	}
	return served.Bytes(), nil
}

// decodeAvatar returns the stored pixels; the EXIF orientation is applied
// later, to the small result only.
func decodeAvatar(data []byte, format string, cfg image.Config) (image.Image, error) {
	if format != "gif" {
		img, _, err := image.Decode(bytes.NewReader(data))
		return img, err
	}

	frame, err := gif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// the first frame may cover only part of the logical screen
	canvas := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Src)
	return canvas, nil
}

// renderAvatar scales the center square to avatarSize and turns it upright.
// A center square rotates onto the center square, so orienting the small
// result is the same as orienting the full image first.
func renderAvatar(img image.Image, orientation int) *image.RGBA {
	b := img.Bounds()
	side := min(b.Dx(), b.Dy())
	x0 := b.Min.X + (b.Dx()-side)/2
	y0 := b.Min.Y + (b.Dy()-side)/2

	dst := image.NewRGBA(image.Rect(0, 0, avatarSize, avatarSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, image.Rect(x0, y0, x0+side, y0+side), draw.Src, nil)
	return orientImage(dst, orientation)
}
