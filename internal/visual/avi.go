package visual

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

// EncodeMJPEGAVI writes a motion-JPEG AVI file playable on Windows without extra software.
func EncodeMJPEGAVI(path string, frames []*Canvas, fps int) error {
	if len(frames) == 0 || frames[0] == nil {
		return os.ErrInvalid
	}
	if fps <= 0 {
		fps = 30
	}
	w, h := frames[0].W, frames[0].H

	jpegs := make([][]byte, 0, len(frames))
	for _, f := range frames {
		if f == nil || f.W != w || f.H != h {
			return os.ErrInvalid
		}
		buf := &bytes.Buffer{}
		if err := jpeg.Encode(buf, canvasToRGBA(f), &jpeg.Options{Quality: 85}); err != nil {
			return err
		}
		jpegs = append(jpegs, buf.Bytes())
	}

	var moviData bytes.Buffer
	index := make([]idxEntry, 0, len(jpegs))
	moviDataOffset := 0
	for _, jpegData := range jpegs {
		index = append(index, idxEntry{offset: moviDataOffset, size: len(jpegData)})
		writeChunk(&moviData, "00dc", jpegData)
		moviDataOffset = moviData.Len()
	}

	var hdrl bytes.Buffer
	writeChunk(&hdrl, "avih", aviMainHeader(w, h, fps, len(jpegs), moviData.Len()))
	var strl bytes.Buffer
	writeChunk(&strl, "strh", streamHeader(w, h, fps, len(jpegs)))
	writeChunk(&strl, "strf", bitmapInfoHeader(w, h))
	writeList(&hdrl, "strl", strl.Bytes())

	var hdrlList bytes.Buffer
	writeList(&hdrlList, "hdrl", hdrl.Bytes())

	var moviList bytes.Buffer
	writeList(&moviList, "movi", moviData.Bytes())

	var idx1 bytes.Buffer
	writeChunk(&idx1, "idx1", buildIndex(index))

	var body bytes.Buffer
	body.Write(hdrlList.Bytes())
	body.Write(moviList.Bytes())
	body.Write(idx1.Bytes())

	var riff bytes.Buffer
	riff.WriteString("RIFF")
	sizePos := riff.Len()
	riff.Write([]byte{0, 0, 0, 0})
	riff.WriteString("AVI ")
	riff.Write(body.Bytes())
	binary.LittleEndian.PutUint32(riff.Bytes()[sizePos:], uint32(riff.Len()-8))

	return os.WriteFile(path, riff.Bytes(), 0644)
}

func canvasToRGBA(c *Canvas) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.W, c.H))
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			p := c.Pix[y*c.W+x]
			img.SetRGBA(x, y, color.RGBA{
				R: byte(p >> 16),
				G: byte(p >> 8),
				B: byte(p),
				A: 255,
			})
		}
	}
	return img
}

func writeList(buf *bytes.Buffer, listType string, content []byte) {
	buf.WriteString("LIST")
	writeU32(buf, uint32(4+len(content)))
	buf.WriteString(listType)
	buf.Write(content)
}

func writeChunk(buf *bytes.Buffer, fourCC string, data []byte) {
	buf.WriteString(fourCC)
	writeU32(buf, uint32(len(data)))
	buf.Write(data)
	if len(data)%2 == 1 {
		buf.WriteByte(0)
	}
}

func writeU32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

func aviMainHeader(w, h, fps, frameCount, moviSize int) []byte {
	b := make([]byte, 56)
	usPerFrame := uint32(1_000_000 / fps)
	binary.LittleEndian.PutUint32(b[0:], usPerFrame)
	binary.LittleEndian.PutUint32(b[4:], uint32(moviSize*fps/iMax(1, frameCount)))
	binary.LittleEndian.PutUint32(b[12:], 0x10) // has index
	binary.LittleEndian.PutUint32(b[16:], uint32(frameCount))
	binary.LittleEndian.PutUint32(b[24:], 1)
	binary.LittleEndian.PutUint32(b[28:], uint32(w*h*3))
	binary.LittleEndian.PutUint32(b[32:], uint32(w))
	binary.LittleEndian.PutUint32(b[36:], uint32(h))
	return b
}

func streamHeader(w, h, fps, frameCount int) []byte {
	b := make([]byte, 56)
	copy(b[0:], []byte("vids"))
	copy(b[4:], []byte("MJPG"))
	binary.LittleEndian.PutUint32(b[20:], 1)
	binary.LittleEndian.PutUint32(b[24:], uint32(fps))
	binary.LittleEndian.PutUint32(b[32:], uint32(frameCount))
	binary.LittleEndian.PutUint32(b[40:], uint32(frameCount))
	binary.LittleEndian.PutUint32(b[44:], uint32(1_000_000 / fps))
	binary.LittleEndian.PutUint32(b[48:], uint32(w))
	binary.LittleEndian.PutUint32(b[52:], uint32(h))
	return b
}

func bitmapInfoHeader(w, h int) []byte {
	b := make([]byte, 40)
	binary.LittleEndian.PutUint32(b[0:], 40)
	binary.LittleEndian.PutUint32(b[4:], uint32(w))
	binary.LittleEndian.PutUint32(b[8:], uint32(h))
	binary.LittleEndian.PutUint16(b[12:], 1)
	binary.LittleEndian.PutUint16(b[14:], 24)
	copy(b[16:], []byte("MJPG"))
	binary.LittleEndian.PutUint32(b[20:], uint32(w*h*3))
	return b
}

func buildIndex(entries []idxEntry) []byte {
	b := make([]byte, 16*len(entries))
	for i, e := range entries {
		off := i * 16
		binary.LittleEndian.PutUint32(b[off:], 0x63643030) // '00dc'
		binary.LittleEndian.PutUint32(b[off+4:], 0x10)
		binary.LittleEndian.PutUint32(b[off+8:], uint32(4+e.offset))
		binary.LittleEndian.PutUint32(b[off+12:], uint32(e.size))
	}
	return b
}

type idxEntry struct {
	offset int
	size   int
}
