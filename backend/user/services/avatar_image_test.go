package services

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/gen2brain/webp"

	"sen1or/letslive/user/domains"
)

func solidImage(width, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}
	return img
}

func encodeImage(t *testing.T, format string, img image.Image) []byte {
	t.Helper()

	var buf bytes.Buffer
	var err error
	switch format {
	case "png":
		err = png.Encode(&buf, img)
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "gif":
		err = gif.Encode(&buf, img, &gif.Options{NumColors: 2})
	case "webp":
		err = webp.Encode(&buf, img)
	default:
		t.Fatalf("unknown format %q", format)
	}
	if err != nil {
		t.Fatalf("encode %s: %v", format, err)
	}
	return buf.Bytes()
}

func TestCheckAvatarDimensionsAcceptsEveryFormat(t *testing.T) {
	for _, format := range []string{"png", "jpeg", "gif", "webp"} {
		t.Run(format, func(t *testing.T) {
			data := encodeImage(t, format, solidImage(120, 90))

			if err := checkAvatarDimensions(bytes.NewReader(data)); err != nil {
				t.Errorf("got %v, want nil", err)
			}
		})
	}
}

func TestCheckAvatarDimensionsBounds(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		wantErr       error
	}{
		{"smallest allowed", MinAvatarDimension, MinAvatarDimension, nil},
		{"largest allowed", MaxAvatarDimension, MinAvatarDimension, nil},
		{"too narrow", MinAvatarDimension - 1, 200, domains.ErrImageDimensionsOutOfRange},
		{"too short", 200, MinAvatarDimension - 1, domains.ErrImageDimensionsOutOfRange},
		{"too wide", MaxAvatarDimension + 1, 200, domains.ErrImageDimensionsOutOfRange},
		{"too tall", 200, MaxAvatarDimension + 1, domains.ErrImageDimensionsOutOfRange},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := encodeImage(t, "png", solidImage(tc.width, tc.height))

			err := checkAvatarDimensions(bytes.NewReader(data))
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestCheckAvatarDimensionsRejectsNonImages(t *testing.T) {
	inputs := map[string][]byte{
		"text":  []byte("definitely not an image"),
		"svg":   []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"></svg>`),
		"empty": {},
	}

	for name, data := range inputs {
		t.Run(name, func(t *testing.T) {
			err := checkAvatarDimensions(bytes.NewReader(data))
			if !errors.Is(err, domains.ErrInvalidImage) {
				t.Errorf("got %v, want %v", err, domains.ErrInvalidImage)
			}
		})
	}
}
