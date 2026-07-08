package audio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

// DecodeFileAIFF reads an AIFF/AIFC file and returns mono PCM samples.
func DecodeFileAIFF(path string) ([]float64, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	return DecodeAIFF(f)
}

// DecodeAIFF decodes AIFF PCM to mono samples normalized to [-1, 1].
func DecodeAIFF(r io.Reader) ([]float64, int, error) {
	if r == nil {
		return nil, 0, errors.New("nil reader")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, err
	}
	if len(data) < 12 || string(data[0:4]) != "FORM" {
		return nil, 0, errors.New("not an AIFF file")
	}
	formType := string(data[8:12])
	if formType != "AIFF" && formType != "AIFC" {
		return nil, 0, fmt.Errorf("unsupported FORM type %q", formType)
	}

	var (
		channels      uint16
		sampleRate    float64
		bitsPerSample uint16
		encoding      string
		soundData     []byte
	)

	for offset := 12; offset+8 <= len(data); {
		id := string(data[offset : offset+4])
		size := int32(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
		offset += 8
		if size < 0 || offset+int(size) > len(data) {
			return nil, 0, fmt.Errorf("invalid %q chunk", id)
		}
		chunk := data[offset : offset+int(size)]
		switch id {
		case "COMM":
			if len(chunk) < 18 {
				return nil, 0, errors.New("COMM chunk too short")
			}
			channels = binary.BigEndian.Uint16(chunk[0:2])
			_ = binary.BigEndian.Uint32(chunk[2:6]) // numSampleFrames
			bitsPerSample = binary.BigEndian.Uint16(chunk[6:8])
			sampleRate = readIEEEExtended(chunk[8:18])
			if formType == "AIFC" && len(chunk) >= 22 {
				encoding = string(chunk[18:22])
			}
		case "SSND":
			if len(chunk) < 8 {
				return nil, 0, errors.New("SSND chunk too short")
			}
			dataOffset := int(binary.BigEndian.Uint32(chunk[0:4]))
			if dataOffset+8 <= len(chunk) {
				soundData = chunk[8+dataOffset:]
			}
		}
		offset += int(size)
		if size%2 == 1 {
			offset++
		}
	}

	if channels == 0 || bitsPerSample == 0 || len(soundData) == 0 {
		return nil, 0, errors.New("AIFF missing audio data")
	}
	if encoding != "" && encoding != "NONE" && encoding != "twos" {
		return nil, 0, fmt.Errorf("unsupported AIFF encoding %q", encoding)
	}

	bytesPerSample := int(bitsPerSample / 8)
	if bytesPerSample <= 0 || bitsPerSample%8 != 0 {
		return nil, 0, fmt.Errorf("unsupported bit depth %d", bitsPerSample)
	}
	frameSize := int(channels) * bytesPerSample
	if frameSize == 0 || len(soundData)%frameSize != 0 {
		return nil, 0, errors.New("AIFF sample data misaligned")
	}

	mono := make([]float64, 0, len(soundData)/frameSize)
	for off := 0; off+frameSize <= len(soundData); off += frameSize {
		var sum float64
		for ch := 0; ch < int(channels); ch++ {
			sum += decodeAIFFSample(soundData[off+ch*bytesPerSample:off+(ch+1)*bytesPerSample], bitsPerSample)
		}
		mono = append(mono, sum/float64(channels))
	}
	sr := int(sampleRate)
	if sr <= 0 {
		sr = 44100
	}
	return mono, sr, nil
}

func decodeAIFFSample(sample []byte, bits uint16) float64 {
	switch bits {
	case 8:
		return (float64(sample[0]) - 128) / 128
	case 16:
		v := int16(binary.BigEndian.Uint16(sample))
		return float64(v) / float64(math.MaxInt16)
	case 24:
		v := int32(sample[0])<<16 | int32(sample[1])<<8 | int32(sample[2])
		if v&0x800000 != 0 {
			v |= ^0xffffff
		}
		return float64(v) / 8388607
	case 32:
		v := int32(binary.BigEndian.Uint32(sample))
		return float64(v) / float64(math.MaxInt32)
	default:
		return 0
	}
}

func readIEEEExtended(b []byte) float64 {
	if len(b) < 10 {
		return 44100
	}
	exp := int16(binary.BigEndian.Uint16(b[0:2])) - 16383
	hi := binary.BigEndian.Uint32(b[2:6])
	lo := binary.BigEndian.Uint32(b[6:10])
	mantissa := float64(hi)*math.Pow(2, 32) + float64(lo)
	return math.Ldexp(mantissa/float64(1<<31), int(exp))
}
