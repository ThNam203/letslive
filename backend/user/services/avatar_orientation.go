package services

import (
	"bytes"
	"encoding/binary"
	"image"

	"github.com/gen2brain/webp"
)

const (
	exifOrientationTag = 0x0112
	exifTypeShort      = 3
)

// avatarOrientation reads the EXIF orientation (1-8) of an upload, 1 when the
// format carries none or the metadata is malformed.
func avatarOrientation(data []byte, format string) int {
	switch format {
	case "jpeg":
		return jpegOrientation(data)
	case "webp":
		exif, err := webp.DecodeExif(bytes.NewReader(data))
		if err != nil || exif.Orientation < 1 || exif.Orientation > 8 {
			return 1
		}
		return exif.Orientation
	default:
		return 1
	}
}

// jpegOrientation reads the EXIF orientation (1-8) from a JPEG, returning 1
// when there is none or the metadata is malformed.
func jpegOrientation(data []byte) int {
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}

	for i := 2; i+4 <= len(data); {
		if data[i] != 0xFF {
			return 1
		}
		marker := data[i+1]
		if marker == 0xFF {
			i++
			continue
		}
		// start of scan or end of image: no metadata segments follow
		if marker == 0xDA || marker == 0xD9 {
			return 1
		}

		size := int(binary.BigEndian.Uint16(data[i+2:]))
		if size < 2 || i+2+size > len(data) {
			return 1
		}
		segment := data[i+4 : i+2+size]
		if marker == 0xE1 && bytes.HasPrefix(segment, []byte("Exif\x00\x00")) {
			return tiffOrientation(segment[6:])
		}
		i += 2 + size
	}

	return 1
}

func tiffOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 1
	}

	var order binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}

	ifd := int64(order.Uint32(tiff[4:]))
	if ifd < 8 || ifd > int64(len(tiff))-2 {
		return 1
	}

	count := int(order.Uint16(tiff[ifd:]))
	for n := 0; n < count; n++ {
		entry := int(ifd) + 2 + n*12
		if entry+12 > len(tiff) {
			return 1
		}
		if order.Uint16(tiff[entry:]) != exifOrientationTag {
			continue
		}
		if order.Uint16(tiff[entry+2:]) != exifTypeShort {
			return 1
		}
		if o := int(order.Uint16(tiff[entry+8:])); o >= 1 && o <= 8 {
			return o
		}
		return 1
	}

	return 1
}

// storedPoint maps pixel (x, y) of the upright image to the pixel that holds
// it in a w x h image stored with the given EXIF orientation.
func storedPoint(orientation, w, h, x, y int) image.Point {
	switch orientation {
	case 2:
		return image.Pt(w-1-x, y)
	case 3:
		return image.Pt(w-1-x, h-1-y)
	case 4:
		return image.Pt(x, h-1-y)
	case 5:
		return image.Pt(y, x)
	case 6:
		return image.Pt(y, h-1-x)
	case 7:
		return image.Pt(w-1-y, h-1-x)
	case 8:
		return image.Pt(w-1-y, x)
	default:
		return image.Pt(x, y)
	}
}

func orientImage(src *image.RGBA, orientation int) *image.RGBA {
	if orientation < 2 || orientation > 8 {
		return src
	}

	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	if orientation >= 5 {
		dw, dh = h, w
	}

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			p := storedPoint(orientation, w, h, x, y)
			dst.SetRGBA(x, y, src.RGBAAt(b.Min.X+p.X, b.Min.Y+p.Y))
		}
	}

	return dst
}
