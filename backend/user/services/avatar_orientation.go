package services

import (
	"bytes"
	"encoding/binary"
	"image"
)

const (
	exifOrientationTag = 0x0112
	exifTypeShort      = 3
)

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

// orientImage turns an image stored with the given EXIF orientation upright.
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
			var sx, sy int
			switch orientation {
			case 2:
				sx, sy = w-1-x, y
			case 3:
				sx, sy = w-1-x, h-1-y
			case 4:
				sx, sy = x, h-1-y
			case 5:
				sx, sy = y, x
			case 6:
				sx, sy = y, h-1-x
			case 7:
				sx, sy = w-1-y, h-1-x
			case 8:
				sx, sy = w-1-y, x
			}
			dst.SetRGBA(x, y, src.RGBAAt(b.Min.X+sx, b.Min.Y+sy))
		}
	}

	return dst
}
