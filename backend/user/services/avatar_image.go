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

// AvatarCrop is a square in the image as browsers display it, i.e. after the
// EXIF orientation is applied.
type AvatarCrop struct {
	X, Y, Size int
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

func processAvatar(ctx context.Context, r io.Reader, crop *AvatarCrop) (*processedAvatar, error) {
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

	served, err := encodeAvatar(data, format, cfg, crop)
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

func encodeAvatar(data []byte, format string, cfg image.Config, crop *AvatarCrop) ([]byte, error) {
	img, err := decodeAvatar(data, format, cfg)
	if err != nil {
		return nil, fmt.Errorf("decode avatar: %w: %w", domains.ErrInvalidImage, err)
	}

	square, err := renderAvatar(img, avatarOrientation(data, format), crop)
	if err != nil {
		return nil, err
	}

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

// renderAvatar scales the crop (the center square when nil) to avatarSize.
// The crop is mapped onto the stored pixels so that only the small result
// needs rotating, never the full image.
func renderAvatar(img image.Image, orientation int, crop *AvatarCrop) (*image.RGBA, error) {
	b := img.Bounds()
	storedW, storedH := b.Dx(), b.Dy()
	shownW, shownH := storedW, storedH
	if orientation >= 5 {
		shownW, shownH = storedH, storedW
	}

	region, err := avatarRegion(shownW, shownH, crop)
	if err != nil {
		return nil, err
	}
	src := storedRect(region, orientation, storedW, storedH).Add(b.Min)

	dst := image.NewRGBA(image.Rect(0, 0, avatarSize, avatarSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, src, draw.Src, nil)
	return orientImage(dst, orientation), nil
}

func avatarRegion(width, height int, crop *AvatarCrop) (image.Rectangle, error) {
	if crop == nil {
		side := min(width, height)
		x0, y0 := (width-side)/2, (height-side)/2
		return image.Rect(x0, y0, x0+side, y0+side), nil
	}

	if crop.Size < minAvatarDimension || crop.X < 0 || crop.Y < 0 ||
		crop.Size > width || crop.Size > height ||
		crop.X > width-crop.Size || crop.Y > height-crop.Size {
		return image.Rectangle{}, fmt.Errorf("crop %+v outside %dx%d: %w", *crop, width, height, domains.ErrInvalidInput)
	}

	return image.Rect(crop.X, crop.Y, crop.X+crop.Size, crop.Y+crop.Size), nil
}
