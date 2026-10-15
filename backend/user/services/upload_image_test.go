package services

import (
	"bytes"
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
	case "gif":
		err = gif.Encode(&buf, img, nil)
	default:
		t.Fatalf("unknown format %q", format)
	}
	if err != nil {
		t.Fatalf("encode %s: %v", format, err)
	}
	return buf.Bytes()
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

func TestReadImageBounds(t *testing.T) {
	cases := []struct {
		name          string
		rule          imageRule
		width, height int
		wantErr       error
	}{
		{"avatar smallest", avatarRule, 80, 80, nil},
		{"avatar too narrow", avatarRule, 79, 200, domains.ErrImageDimensionsOutOfRange},
		{"avatar too short", avatarRule, 200, 79, domains.ErrImageDimensionsOutOfRange},
		{"background tiny", backgroundRule, 1, 1, nil},
		{"thumbnail smallest", livestreamThumbnailRule, 320, 180, nil},
		{"thumbnail too short", livestreamThumbnailRule, 320, 179, domains.ErrImageDimensionsOutOfRange},
		{"general narrow", generalImageRule, 1, 300, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := encodeImage(t, "jpeg", solidImage(tc.width, tc.height))

			_, err := readImage(bytes.NewReader(data), tc.rule)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestReadImageRejectsSidesOverTheLimit(t *testing.T) {
	rules := map[string]imageRule{
		"avatar":     avatarRule,
		"background": backgroundRule,
		"thumbnail":  livestreamThumbnailRule,
		"general":    generalImageRule,
	}
	for name, rule := range rules {
		for _, size := range [][2]int{{rule.maxDimension, 200}, {rule.maxDimension + 1, 200}, {200, rule.maxDimension + 1}} {
			data := encodeImage(t, "jpeg", solidImage(size[0], size[1]))
			_, err := readImage(bytes.NewReader(data), rule)
			want := error(nil)
			if size[0] > rule.maxDimension || size[1] > rule.maxDimension {
				want = domains.ErrImageDimensionsOutOfRange
			}
			if !errors.Is(err, want) {
				t.Errorf("%s %dx%d: got %v, want %v", name, size[0], size[1], err, want)
			}
		}
	}
}

func TestReadImageFormats(t *testing.T) {
	rules := map[string]imageRule{
		"avatar":     avatarRule,
		"background": backgroundRule,
		"thumbnail":  livestreamThumbnailRule,
		"general":    generalImageRule,
	}
	cases := []struct {
		format     string
		want       imageFormat
		acceptedBy map[string]bool
	}{
		{"jpeg", jpegFormat, map[string]bool{"avatar": true, "background": true, "thumbnail": true, "general": true}},
		{"webp", webpFormat, map[string]bool{"avatar": true, "background": true, "thumbnail": true, "general": true}},
		{"gif", gifFormat, map[string]bool{"general": true}},
		{"png", imageFormat{}, map[string]bool{}},
	}

	for _, tc := range cases {
		data := encodeImage(t, tc.format, solidImage(400, 225))
		for ruleName, rule := range rules {
			t.Run(tc.format+" "+ruleName, func(t *testing.T) {
				img, err := readImage(bytes.NewReader(data), rule)

				if !tc.acceptedBy[ruleName] {
					if !errors.Is(err, domains.ErrInvalidImage) {
						t.Errorf("got %v, want %v", err, domains.ErrInvalidImage)
					}
					return
				}
				if err != nil {
					t.Fatalf("got %v, want nil", err)
				}
				if img.format != tc.want {
					t.Errorf("format = %+v, want %+v", img.format, tc.want)
				}
				if !bytes.Equal(img.data, data) {
					t.Error("image bytes were altered")
				}
			})
		}
	}
}

func TestReadImageRejectsNonImages(t *testing.T) {
	cases := map[string][]byte{
		"text":  []byte("hello"),
		"empty": {},
		"svg":   []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"></svg>`),
		"html":  []byte("<!doctype html><script>alert(1)</script>"),
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := readImage(bytes.NewReader(data), generalImageRule)
			if !errors.Is(err, domains.ErrInvalidImage) {
				t.Errorf("got %v, want %v", err, domains.ErrInvalidImage)
			}
		})
	}
}

func decodeWebP(t *testing.T, data []byte, width, height int) image.Image {
	t.Helper()

	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("not a webp: %v", err)
	}
	if got := img.Bounds().Size(); got != image.Pt(width, height) {
		t.Fatalf("webp is %v, want %dx%d", got, width, height)
	}
	return img
}

func TestToStoredImageReencodesToWebPAtTheSameSize(t *testing.T) {
	for _, format := range []string{"jpeg", "webp"} {
		t.Run(format, func(t *testing.T) {
			data := encodeImage(t, format, stripedImage(480, 160, red, green, blue))
			upload, err := readImage(bytes.NewReader(data), generalImageRule)
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			stored, err := toStoredImage(upload)
			if err != nil {
				t.Fatalf("got %v, want nil", err)
			}
			if stored.format != webpFormat {
				t.Errorf("format = %+v, want %+v", stored.format, webpFormat)
			}

			img := decodeWebP(t, stored.data, 480, 160)
			assertColorNear(t, img, 40, 80, red)
			assertColorNear(t, img, 240, 80, green)
			assertColorNear(t, img, 440, 80, blue)
		})
	}
}

func TestToStoredImageKeepsGIFs(t *testing.T) {
	data := encodeImage(t, "gif", solidImage(300, 200))
	upload, err := readImage(bytes.NewReader(data), generalImageRule)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	stored, err := toStoredImage(upload)
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}
	if stored.format != gifFormat || !bytes.Equal(stored.data, data) {
		t.Errorf("GIF was altered: format %+v, %d bytes, want %d", stored.format, len(stored.data), len(data))
	}
}

func TestToStoredImageRejectsTruncatedImages(t *testing.T) {
	data := encodeImage(t, "jpeg", solidImage(200, 200))

	_, err := toStoredImage(&uploadedImage{data: data[:len(data)/3], format: jpegFormat})
	if !errors.Is(err, domains.ErrInvalidImage) {
		t.Errorf("got %v, want %v", err, domains.ErrInvalidImage)
	}
}
