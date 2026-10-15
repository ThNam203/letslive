package services

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"io"
	"sen1or/letslive/user/domains"

	"github.com/gen2brain/webp"
)

const (
	minAvatarDimension              = 80
	maxAvatarDimension              = 10000
	maxBackgroundDimension          = 10000
	minLivestreamThumbnailDimension = 180
	maxLivestreamThumbnailDimension = 10000
	maxGeneralImageDimension        = 10000

	storedWebPQuality = 80
)

type imageFormat struct{ ext, contentType string }

var (
	jpegFormat = imageFormat{"jpg", "image/jpeg"}
	webpFormat = imageFormat{"webp", "image/webp"}
	gifFormat  = imageFormat{"gif", "image/gif"}
)

// formats is keyed by the format name that image.DecodeConfig reports
type imageRule struct {
	formats      map[string]imageFormat
	minDimension int
	maxDimension int
}

var (
	preparedFormats = map[string]imageFormat{"jpeg": jpegFormat, "webp": webpFormat}

	avatarRule = imageRule{
		formats:      preparedFormats,
		minDimension: minAvatarDimension,
		maxDimension: maxAvatarDimension,
	}
	backgroundRule = imageRule{
		formats:      preparedFormats,
		maxDimension: maxBackgroundDimension,
	}
	livestreamThumbnailRule = imageRule{
		formats:      preparedFormats,
		minDimension: minLivestreamThumbnailDimension,
		maxDimension: maxLivestreamThumbnailDimension,
	}
	generalImageRule = imageRule{
		formats:      map[string]imageFormat{"jpeg": jpegFormat, "webp": webpFormat, "gif": gifFormat},
		maxDimension: maxGeneralImageDimension,
	}
)

type uploadedImage struct {
	data   []byte
	format imageFormat
}

// readImage checks only the header; toStoredImage decodes the pixels.
func readImage(r io.Reader, rule imageRule) (*uploadedImage, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read image upload: %w: %w", domains.ErrInternal, err)
	}

	cfg, name, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("read image header: %w: %w", domains.ErrInvalidImage, err)
	}
	format, ok := rule.formats[name]
	if !ok {
		return nil, fmt.Errorf("image format %q: %w", name, domains.ErrInvalidImage)
	}
	if min(cfg.Width, cfg.Height) < rule.minDimension || max(cfg.Width, cfg.Height) > rule.maxDimension {
		return nil, fmt.Errorf("image is %dx%d: %w", cfg.Width, cfg.Height, domains.ErrImageDimensionsOutOfRange)
	}

	return &uploadedImage{data: data, format: format}, nil
}

// GIFs are kept as uploaded: the WebP encoder writes a single frame.
func toStoredImage(upload *uploadedImage) (*uploadedImage, error) {
	if upload.format == gifFormat {
		return upload, nil
	}

	img, _, err := image.Decode(bytes.NewReader(upload.data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w: %w", domains.ErrInvalidImage, err)
	}

	var out bytes.Buffer
	if err := webp.Encode(&out, img, webp.Options{Quality: storedWebPQuality, Method: 6}); err != nil {
		return nil, fmt.Errorf("encode webp: %w: %w", domains.ErrInternal, err)
	}

	return &uploadedImage{data: out.Bytes(), format: webpFormat}, nil
}
