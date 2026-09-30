package services

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sen1or/letslive/user/domains"

	_ "github.com/gen2brain/webp"
)

const (
	MinAvatarDimension = 80
	MaxAvatarDimension = 4096
)

func checkAvatarDimensions(r io.Reader) error {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return fmt.Errorf("read avatar header: %w: %w", domains.ErrInvalidImage, err)
	}

	if cfg.Width < MinAvatarDimension || cfg.Height < MinAvatarDimension ||
		cfg.Width > MaxAvatarDimension || cfg.Height > MaxAvatarDimension {
		return fmt.Errorf("avatar is %dx%d: %w", cfg.Width, cfg.Height, domains.ErrImageDimensionsOutOfRange)
	}

	return nil
}
