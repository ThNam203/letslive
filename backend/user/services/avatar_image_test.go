package services

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/gif"
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

// stripedImage fills equal-width vertical stripes, left to right.
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
	case "gif":
		err = gif.Encode(&buf, img, nil)
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
	if got := img.Bounds().Size(); got != image.Pt(AvatarSize, AvatarSize) {
		t.Fatalf("served avatar is %v, want %dx%d", got, AvatarSize, AvatarSize)
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

func TestCheckAvatarDimensionsAcceptsEveryFormat(t *testing.T) {
	for _, format := range []string{"png", "jpeg", "gif", "webp"} {
		t.Run(format, func(t *testing.T) {
			data := encodeImage(t, format, solidImage(120, 90))

			got, err := checkAvatarDimensions(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("got %v, want nil", err)
			}
			if got != format {
				t.Errorf("format = %q, want %q", got, format)
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

			_, err := checkAvatarDimensions(bytes.NewReader(data))
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
			_, err := checkAvatarDimensions(bytes.NewReader(data))
			if !errors.Is(err, domains.ErrInvalidImage) {
				t.Errorf("got %v, want %v", err, domains.ErrInvalidImage)
			}
		})
	}
}

func TestProcessAvatarServesSquareWebPAndKeepsOriginal(t *testing.T) {
	cases := []struct {
		format, ext, contentType string
	}{
		{"png", "png", "image/png"},
		{"jpeg", "jpg", "image/jpeg"},
		{"gif", "gif", "image/gif"},
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
	data := encodeImage(t, "png", stripedImage(480, 160, red, green, blue))

	avatar, err := processAvatar(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	img := decodeServedAvatar(t, avatar)
	for _, x := range []int{4, AvatarSize / 2, AvatarSize - 5} {
		assertColorNear(t, img, x, AvatarSize/2, green)
	}
}

func TestProcessAvatarAppliesJPEGOrientation(t *testing.T) {
	// left half red, right half blue as stored; the tag says how to display it
	stored := encodeImage(t, "jpeg", stripedImage(160, 80, red, blue))

	cases := []struct {
		name        string
		orientation uint16
		top, bottom color.RGBA
	}{
		{"rotate 90 clockwise", 6, red, blue},
		{"rotate 90 counter-clockwise", 8, blue, red},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := withEXIFOrientation(stored, bigEndian, tc.orientation)

			avatar, err := processAvatar(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("got %v, want nil", err)
			}

			img := decodeServedAvatar(t, avatar)
			assertColorNear(t, img, AvatarSize/2, 8, tc.top)
			assertColorNear(t, img, AvatarSize/2, AvatarSize-9, tc.bottom)
		})
	}
}

func TestProcessAvatarPlacesPartialGIFFrameOnItsCanvas(t *testing.T) {
	frame := image.NewPaletted(image.Rect(50, 50, 150, 150), color.Palette{color.Transparent, green})
	for i := range frame.Pix {
		frame.Pix[i] = 1
	}
	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, &gif.GIF{
		Image:  []*image.Paletted{frame},
		Delay:  []int{0},
		Config: image.Config{Width: 200, Height: 200},
	})
	if err != nil {
		t.Fatalf("encode gif: %v", err)
	}

	avatar, err := processAvatar(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}

	img := decodeServedAvatar(t, avatar)
	assertColorNear(t, img, AvatarSize/2, AvatarSize/2, green)
	if _, _, _, a := img.At(2, 2).RGBA(); a>>8 > 48 {
		t.Errorf("corner alpha = %d, want the transparent canvas around the frame", a>>8)
	}
}

func TestProcessAvatarLimitedGivesUpWhenTheRequestEnds(t *testing.T) {
	for range cap(avatarProcessingSlots) {
		avatarProcessingSlots <- struct{}{}
	}
	defer func() {
		for range cap(avatarProcessingSlots) {
			<-avatarProcessingSlots
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := processAvatarLimited(ctx, bytes.NewReader(encodeImage(t, "png", solidImage(100, 100))))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want %v", err, context.Canceled)
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
		{"truncated body", valid[:len(valid)/3], domains.ErrInvalidImage},
		{"too small", encodeImage(t, "png", solidImage(40, 40)), domains.ErrImageDimensionsOutOfRange},
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
