package images

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func picture() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for x := 0; x < 40; x++ {
		img.Set(x, x%20, color.RGBA{R: 200, A: 255})
	}
	return img
}

// exifSegment builds an APP1 EXIF block with an orientation tag and some
// fake GPS text that must not survive.
func exifSegment(orientation int) []byte {
	tiff := []byte{
		'I', 'I', 0x2A, 0x00, 8, 0, 0, 0,
		0x01, 0x00, // one entry
		0x12, 0x01, 0x03, 0x00, 1, 0, 0, 0, byte(orientation), 0, 0, 0,
		0, 0, 0, 0,
	}
	payload := append([]byte("Exif\x00\x00"), tiff...)
	payload = append(payload, []byte("GPSLatitude 31.9539 GPSLongitude 35.9106")...)
	seg := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(seg[2:], uint16(len(payload)+2))
	return append(seg, payload...)
}

func TestJPEGMetadataIsStrippedButOrientationKept(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, picture(), nil); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	withExif := append(append(append([]byte{}, raw[:2]...), exifSegment(6)...), raw[2:]...)

	img, err := Inspect(withExif)
	if err != nil {
		t.Fatal(err)
	}
	if img.ContentType != "image/jpeg" || img.Ext != ".jpg" || img.Width != 40 || img.Height != 20 {
		t.Fatalf("unexpected result: %+v", img)
	}
	if bytes.Contains(img.Data, []byte("GPSLatitude")) {
		t.Fatal("GPS data must be removed")
	}
	if _, err := jpeg.Decode(bytes.NewReader(img.Data)); err != nil {
		t.Fatalf("cleaned JPEG must still decode: %v", err)
	}
	// The orientation survived in the new minimal EXIF block.
	i := bytes.Index(img.Data, []byte("Exif\x00\x00"))
	if i < 0 || exifOrientation(img.Data[i:]) != 6 {
		t.Fatal("orientation should be kept so portrait photos are not shown sideways")
	}
}

func pngChunk(typ string, data []byte) []byte {
	out := make([]byte, 8, 12+len(data))
	binary.BigEndian.PutUint32(out, uint32(len(data)))
	copy(out[4:], typ)
	out = append(out, data...)
	crc := crc32.ChecksumIEEE(append([]byte(typ), data...))
	return binary.BigEndian.AppendUint32(out, crc)
}

func TestPNGTextChunksAreStripped(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, picture()); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	// Insert a tEXt chunk right after IHDR (8 signature + 25 IHDR bytes).
	withText := append(append(append([]byte{}, raw[:33]...), pngChunk("tEXt", []byte("Location\x00home address"))...), raw[33:]...)
	img, err := Inspect(withText)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(img.Data, []byte("home address")) {
		t.Fatal("PNG text must be removed")
	}
	if _, err := png.Decode(bytes.NewReader(img.Data)); err != nil {
		t.Fatalf("cleaned PNG must still decode: %v", err)
	}
}

func TestRejectsNonImagesAndHugeImages(t *testing.T) {
	if _, err := Inspect([]byte("<html><script>alert(1)</script></html>")); err != ErrUnsupported {
		t.Fatalf("HTML must be refused, got %v", err)
	}
	if _, err := Inspect([]byte("GIF89a\x01\x00\x01\x00")); err != ErrUnsupported {
		t.Fatalf("GIF is not accepted, got %v", err)
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, picture(), nil)
	if _, err := Inspect(buf.Bytes()[:200]); err == nil {
		t.Fatal("a truncated JPEG must be refused")
	}

	// A PNG header claiming 20000x20000 pixels (a "decompression bomb").
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:], 20000)
	binary.BigEndian.PutUint32(ihdr[4:], 20000)
	ihdr[8], ihdr[9] = 8, 2 // 8-bit RGB
	bomb := append([]byte("\x89PNG\r\n\x1a\n"), pngChunk("IHDR", ihdr)...)
	bomb = append(bomb, pngChunk("IEND", nil)...)
	if _, err := Inspect(bomb); err == nil {
		t.Fatal("gigantic dimensions must be refused")
	}
}

func TestWebPMetadataChunksAreStripped(t *testing.T) {
	chunk := func(typ string, data []byte) []byte {
		out := append([]byte(typ), 0, 0, 0, 0)
		binary.LittleEndian.PutUint32(out[4:], uint32(len(data)))
		out = append(out, data...)
		if len(data)%2 == 1 {
			out = append(out, 0)
		}
		return out
	}
	vp8x := make([]byte, 10)
	vp8x[0] = 0x08 | 0x04 // has EXIF + XMP
	body := append([]byte("WEBP"), chunk("VP8X", vp8x)...)
	body = append(body, chunk("EXIF", []byte("GPSLatitude 31.95"))...)
	body = append(body, chunk("XMP ", []byte("<x>secret</x>"))...)
	body = append(body, chunk("VP8L", []byte{0x2f, 1, 2, 3, 4})...)
	riff := append([]byte("RIFF"), 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(riff[4:], uint32(len(body)))
	data := append(riff, body...)

	out, err := stripWebP(data)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte("GPSLatitude")) || bytes.Contains(out, []byte("secret")) {
		t.Fatal("EXIF/XMP chunks must be removed")
	}
	if int(binary.LittleEndian.Uint32(out[4:8])) != len(out)-8 {
		t.Fatal("RIFF size must match the new length")
	}
	if out[20]&(0x08|0x04) != 0 { // VP8X flags byte: 12 (RIFF+WEBP) + 8 (chunk header)
		t.Fatal("VP8X metadata flags must be cleared")
	}
}
