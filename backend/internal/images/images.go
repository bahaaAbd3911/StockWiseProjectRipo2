// Package images checks uploaded pictures and removes private metadata.
//
// Inspect accepts JPEG, PNG and WebP only. The type is decided from the bytes
// (never from the file name or the browser's Content-Type, both of which the
// client controls), and the header is parsed to make sure it really is a
// picture of a sensible size. Nothing is fully decoded, so a "decompression
// bomb" (tiny file, gigantic picture) cannot use up the server's memory.
//
// Phone photos carry EXIF data, often including the GPS position where they
// were taken. That is stripped before the file is stored:
//
//   - JPEG: APP1 (EXIF/XMP), APP13 (IPTC) and comments are removed. The
//     orientation is kept in a tiny new EXIF block, otherwise portrait phone
//     photos would show up sideways. Color profiles (APP2) stay.
//   - PNG: text, time and eXIf chunks are removed.
//   - WebP: EXIF and XMP chunks are removed.
package images

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // register decoders for DecodeConfig
	_ "image/png"
	"net/http"

	_ "golang.org/x/image/webp"
)

// Limits for accepted pictures.
const (
	MaxSide   = 10000      // pixels, either side
	MaxPixels = 40_000_000 // 40 megapixels
)

// ErrUnsupported is returned for anything that is not a JPEG, PNG or WebP.
var ErrUnsupported = errors.New("only JPEG, PNG and WebP images are accepted")

// Image is a checked picture, ready to store.
type Image struct {
	Data          []byte
	ContentType   string // image/jpeg, image/png, image/webp
	Ext           string // .jpg, .png, .webp
	Width, Height int
}

// Inspect validates data and returns it with metadata removed.
func Inspect(data []byte) (Image, error) {
	var img Image
	switch http.DetectContentType(data) {
	case "image/jpeg":
		img.ContentType, img.Ext = "image/jpeg", ".jpg"
	case "image/png":
		img.ContentType, img.Ext = "image/png", ".png"
	case "image/webp":
		img.ContentType, img.Ext = "image/webp", ".webp"
	default:
		return Image{}, ErrUnsupported
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || "image/"+format != img.ContentType {
		return Image{}, errors.New("the file is damaged or not a real image")
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > MaxSide || cfg.Height > MaxSide || cfg.Width*cfg.Height > MaxPixels {
		return Image{}, fmt.Errorf("images may be at most %dx%d pixels (%d megapixels)", MaxSide, MaxSide, MaxPixels/1_000_000)
	}
	img.Width, img.Height = cfg.Width, cfg.Height

	switch img.ContentType {
	case "image/jpeg":
		img.Data, err = stripJPEG(data)
	case "image/png":
		img.Data, err = stripPNG(data)
	case "image/webp":
		img.Data, err = stripWebP(data)
	}
	if err != nil {
		return Image{}, errors.New("the file is damaged or not a real image")
	}
	return img, nil
}

// ---------- JPEG ----------

// stripJPEG copies the segments before the image data, dropping metadata.
func stripJPEG(data []byte) ([]byte, error) {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return nil, errors.New("not a JPEG")
	}
	out := bytes.NewBuffer(make([]byte, 0, len(data)))
	out.Write(data[:2]) // SOI
	orientation := 0
	i := 2
	for i+4 <= len(data) {
		if data[i] != 0xFF {
			return nil, errors.New("bad JPEG marker")
		}
		marker := data[i+1]
		if marker == 0xFF { // fill byte
			i++
			continue
		}
		if marker == 0xDA { // start of scan: the rest is image data, copy as is
			if orientation > 1 {
				out.Write(orientationSegment(orientation))
			}
			out.Write(data[i:])
			return out.Bytes(), nil
		}
		size := int(binary.BigEndian.Uint16(data[i+2 : i+4]))
		if size < 2 || i+2+size > len(data) {
			return nil, errors.New("bad JPEG segment")
		}
		segment := data[i : i+2+size]
		switch {
		case marker == 0xE1: // APP1: EXIF or XMP
			if o := exifOrientation(segment[4:]); o > 0 {
				orientation = o
			}
		case marker == 0xED, marker == 0xFE: // APP13 (IPTC/Photoshop), COM
		default:
			out.Write(segment)
		}
		i += 2 + size
	}
	return nil, errors.New("JPEG without image data")
}

// exifOrientation reads tag 0x0112 from an EXIF block ("Exif\0\0" + TIFF).
func exifOrientation(b []byte) int {
	if len(b) < 14 || string(b[:6]) != "Exif\x00\x00" {
		return 0
	}
	tiff := b[6:]
	var order binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 0
	}
	ifd := int(order.Uint32(tiff[4:8]))
	if ifd+2 > len(tiff) {
		return 0
	}
	n := int(order.Uint16(tiff[ifd : ifd+2]))
	for k := 0; k < n; k++ {
		e := ifd + 2 + k*12
		if e+12 > len(tiff) {
			return 0
		}
		if order.Uint16(tiff[e:e+2]) == 0x0112 {
			if v := int(order.Uint16(tiff[e+8 : e+10])); v >= 1 && v <= 8 {
				return v
			}
			return 0
		}
	}
	return 0
}

// orientationSegment is an APP1 EXIF block with only the orientation tag.
func orientationSegment(o int) []byte {
	tiff := []byte{
		'M', 'M', 0x00, 0x2A, 0, 0, 0, 8, // big-endian TIFF, IFD0 at offset 8
		0x00, 0x01, // one entry
		0x01, 0x12, 0x00, 0x03, 0, 0, 0, 1, byte(o >> 8), byte(o), 0, 0, // Orientation, SHORT, 1 value
		0, 0, 0, 0, // no next IFD
	}
	payload := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(seg[2:], uint16(len(payload)+2))
	return append(seg, payload...)
}

// ---------- PNG ----------

var pngDrop = map[string]bool{"tEXt": true, "zTXt": true, "iTXt": true, "eXIf": true, "tIME": true}

func stripPNG(data []byte) ([]byte, error) {
	const sig = "\x89PNG\r\n\x1a\n"
	if len(data) < 8 || string(data[:8]) != sig {
		return nil, errors.New("not a PNG")
	}
	out := bytes.NewBuffer(make([]byte, 0, len(data)))
	out.WriteString(sig)
	for i := 8; i < len(data); {
		if i+8 > len(data) {
			return nil, errors.New("bad PNG chunk")
		}
		size := int(binary.BigEndian.Uint32(data[i : i+4]))
		end := i + 12 + size // length + type + data + CRC
		if size < 0 || end > len(data) {
			return nil, errors.New("bad PNG chunk")
		}
		typ := string(data[i+4 : i+8])
		if !pngDrop[typ] {
			out.Write(data[i:end])
		}
		i = end
		if typ == "IEND" {
			break
		}
	}
	return out.Bytes(), nil
}

// ---------- WebP ----------

func stripWebP(data []byte) ([]byte, error) {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return nil, errors.New("not a WebP")
	}
	body := bytes.NewBuffer(make([]byte, 0, len(data)))
	body.WriteString("WEBP")
	vp8x := -1 // offset of the VP8X flags byte inside body
	for i := 12; i+8 <= len(data); {
		typ := string(data[i : i+4])
		size := int(binary.LittleEndian.Uint32(data[i+4 : i+8]))
		end := i + 8 + size + size%2 // chunks are padded to an even length
		if size < 0 || end > len(data) {
			if i+8+size == len(data) { // tolerate a missing final pad byte
				end = len(data)
			} else {
				return nil, errors.New("bad WebP chunk")
			}
		}
		switch typ {
		case "EXIF", "XMP ":
		default:
			if typ == "VP8X" {
				vp8x = body.Len() + 8
			}
			body.Write(data[i:end])
		}
		i = end
	}
	out := body.Bytes()
	if vp8x >= 0 && vp8x < len(out) {
		out[vp8x] &^= 0x08 | 0x04 // clear the "has EXIF" and "has XMP" flags
	}
	riff := make([]byte, 8, 8+len(out))
	copy(riff, "RIFF")
	binary.LittleEndian.PutUint32(riff[4:], uint32(len(out)))
	return append(riff, out...), nil
}
