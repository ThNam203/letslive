package services

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"testing"
)

var (
	bigEndian    binary.ByteOrder = binary.BigEndian
	littleEndian binary.ByteOrder = binary.LittleEndian
)

// withEXIFOrientation inserts an APP1 segment holding a single-entry IFD0
// with the orientation tag right after the JPEG SOI marker.
func withEXIFOrientation(jpegData []byte, order binary.ByteOrder, orientation uint16) []byte {
	tiff := make([]byte, 26)
	if order == littleEndian {
		copy(tiff, "II")
	} else {
		copy(tiff, "MM")
	}
	order.PutUint16(tiff[2:], 42)
	order.PutUint32(tiff[4:], 8)
	order.PutUint16(tiff[8:], 1)
	order.PutUint16(tiff[10:], exifOrientationTag)
	order.PutUint16(tiff[12:], exifTypeShort)
	order.PutUint32(tiff[14:], 1)
	order.PutUint16(tiff[18:], orientation)

	payload := append([]byte("Exif\x00\x00"), tiff...)
	segment := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(segment[2:], uint16(len(payload)+2))

	out := append([]byte{}, jpegData[:2]...)
	out = append(out, segment...)
	out = append(out, payload...)
	return append(out, jpegData[2:]...)
}

func TestJPEGOrientation(t *testing.T) {
	plain := encodeImage(t, "jpeg", solidImage(80, 80))

	cases := []struct {
		name string
		data []byte
		want int
	}{
		{"no exif", plain, 1},
		{"big endian", withEXIFOrientation(plain, bigEndian, 6), 6},
		{"little endian", withEXIFOrientation(plain, littleEndian, 8), 8},
		{"out of range value", withEXIFOrientation(plain, bigEndian, 9), 1},
		{"not a jpeg", []byte("hello world"), 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jpegOrientation(tc.data); got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestJPEGOrientationSurvivesTruncation(t *testing.T) {
	data := withEXIFOrientation(encodeImage(t, "jpeg", solidImage(80, 80)), littleEndian, 3)

	for i := range data {
		if got := jpegOrientation(data[:i]); got < 1 || got > 8 {
			t.Fatalf("prefix of %d bytes: got %d", i, got)
		}
	}
}

func TestOrientImage(t *testing.T) {
	// stored pixels, labelled by their red channel:
	//   a b c
	//   d e f
	const a, b, c, d, e, f = 1, 2, 3, 4, 5, 6
	src := image.NewRGBA(image.Rect(0, 0, 3, 2))
	for i, v := range []uint8{a, b, c, d, e, f} {
		src.SetRGBA(i%3, i/3, color.RGBA{R: v, A: 0xff})
	}

	cases := []struct {
		orientation int
		want        [][]uint8
	}{
		{1, [][]uint8{{a, b, c}, {d, e, f}}},
		{2, [][]uint8{{c, b, a}, {f, e, d}}},
		{3, [][]uint8{{f, e, d}, {c, b, a}}},
		{4, [][]uint8{{d, e, f}, {a, b, c}}},
		{5, [][]uint8{{a, d}, {b, e}, {c, f}}},
		{6, [][]uint8{{d, a}, {e, b}, {f, c}}},
		{7, [][]uint8{{f, c}, {e, b}, {d, a}}},
		{8, [][]uint8{{c, f}, {b, e}, {a, d}}},
	}

	for _, tc := range cases {
		got := orientImage(src, tc.orientation)

		size := got.Bounds().Size()
		if size != image.Pt(len(tc.want[0]), len(tc.want)) {
			t.Errorf("orientation %d: size %v, want %dx%d", tc.orientation, size, len(tc.want[0]), len(tc.want))
			continue
		}

		var rows [][]uint8
		for y := 0; y < size.Y; y++ {
			row := make([]uint8, size.X)
			for x := range row {
				row[x] = got.RGBAAt(x, y).R
			}
			rows = append(rows, row)
		}
		if !bytes.Equal(bytes.Join(rows, nil), bytes.Join(tc.want, nil)) {
			t.Errorf("orientation %d: got %v, want %v", tc.orientation, rows, tc.want)
		}
	}
}
