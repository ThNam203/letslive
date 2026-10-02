package services

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/gen2brain/webp"

	"sen1or/letslive/user/domains"
)

var (
	red   = color.RGBA{R: 0xff, A: 0xff}
	green = color.RGBA{G: 0xff, A: 0xff}
	blue  = color.RGBA{B: 0xff, A: 0xff}
)

func solidImage(width, height int) image.Image {
	return stripedImage(width, height, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
}

func stripedImage(width, height int, stripes ...color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		c := stripes[x*len(stripes)/width]
		for y := 0; y < height; y++ {
			img.SetRGBA(x, y, c)
		}
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
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
	case "webp":
		err = webp.Encode(&buf, img, webp.Options{Lossless: true})
	default:
		t.Fatalf("unknown format %q", format)
	}
	if err != nil {
		t.Fatalf("encode %s: %v", format, err)
	}
	return buf.Bytes()
}

func decodeServedAvatar(t *testing.T, avatar *processedAvatar) image.Image {
	t.Helper()

	img, err := webp.Decode(bytes.NewReader(avatar.webp))
	if err != nil {
		t.Fatalf("served avatar is not a webp: %v", err)
	}
	if got := img.Bounds().Size(); got != image.Pt(avatarSize, avatarSize) {
		t.Fatalf("served avatar is %v, want %dx%d", got, avatarSize, avatarSize)
	}
	return img
}

func assertColorNear(t *testing.T, img image.Image, x, y int, want color.RGBA) {
	t.Helper()

	const tolerance = 48
	r, g, b, _ := img.At(x, y).RGBA()
	got := [3]int{int(r >> 8), int(g >> 8), int(b >> 8)}
	exp := [3]int{int(want.R), int(want.G), int(want.B)}
	for i := range got {
		if diff := got[i] - exp[i]; diff > tolerance || diff < -tolerance {
			t.Errorf("pixel (%d,%d) = %v, want about %v", x, y, got, exp)
			return
		}
	}
}

func TestCheckAvatarDimensionsBounds(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		wantErr       error
	}{
		{"smallest allowed", 80, 80, nil},
		{"largest allowed", 2048, 80, nil},
		{"largest allowed tall", 80, 2048, nil},
		{"too narrow", 79, 200, domains.ErrImageDimensionsOutOfRange},
		{"too short", 200, 79, domains.ErrImageDimensionsOutOfRange},
		{"too wide", 2049, 200, domains.ErrImageDimensionsOutOfRange},
		{"too tall", 200, 2049, domains.ErrImageDimensionsOutOfRange},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := encodeImage(t, "jpeg", solidImage(tc.width, tc.height))

			_, err := checkAvatarDimensions(bytes.NewReader(data))
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestProcessAvatarServesSquareWebPAndKeepsOriginal(t *testing.T) {
	cases := []struct {
		format, ext, contentType string
	}{
		{"jpeg", "jpg", "image/jpeg"},
		{"webp", "webp", "image/webp"},
	}

	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			data := encodeImage(t, tc.format, solidImage(300, 200))

			avatar, err := processAvatar(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("got %v, want nil", err)
			}

			decodeServedAvatar(t, avatar)
			if !bytes.Equal(avatar.original, data) {
				t.Error("original bytes were altered")
			}
			if avatar.originalExt != tc.ext {
				t.Errorf("ext = %q, want %q", avatar.originalExt, tc.ext)
			}
			if avatar.originalContentType != tc.contentType {
				t.Errorf("content type = %q, want %q", avatar.originalContentType, tc.contentType)
			}
		})
	}
}

func TestProcessAvatarCropsTheCenter(t *testing.T) {
	data := encodeImage(t, "webp", stripedImage(480, 160, red, green, blue))

	avatar, err := processAvatar(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	img := decodeServedAvatar(t, avatar)
	for _, x := range []int{4, avatarSize / 2, avatarSize - 5} {
		assertColorNear(t, img, x, avatarSize/2, green)
	}
}

func TestProcessAvatarRejectsBadInput(t *testing.T) {
	valid := encodeImage(t, "jpeg", solidImage(200, 200))

	cases := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{"not an image", []byte("hello"), domains.ErrInvalidImage},
		{"empty", []byte{}, domains.ErrInvalidImage},
		{"svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"></svg>`), domains.ErrInvalidImage},
		{"png", encodeImage(t, "png", solidImage(200, 200)), domains.ErrInvalidImage},
		{"truncated body", valid[:len(valid)/3], domains.ErrInvalidImage},
		{"too small", encodeImage(t, "jpeg", solidImage(40, 40)), domains.ErrImageDimensionsOutOfRange},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := processAvatar(bytes.NewReader(tc.data))
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
