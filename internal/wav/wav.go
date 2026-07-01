package wav

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

// Audio contains decoded PCM samples normalized to [-1, 1].
type Audio struct {
	Channels      int
	SampleRate    int
	BitsPerSample int
	Mono          []float64
}

// DurationSeconds returns the decoded audio length in seconds.
func (a *Audio) DurationSeconds() float64 {
	if a == nil || a.SampleRate == 0 {
		return 0
	}
	return float64(len(a.Mono)) / float64(a.SampleRate)
}

// DecodeFile reads a PCM WAV file from disk.
func DecodeFile(path string) (*Audio, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Decode(f)
}

// Decode reads PCM WAV data and downmixes it to mono.
func Decode(r io.Reader) (*Audio, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, errors.New("not a RIFF/WAVE file")
	}

	var (
		format        uint16
		channels      uint16
		sampleRate    uint32
		bitsPerSample uint16
		dataChunk     []byte
	)

	for offset := 12; offset+8 <= len(data); {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		offset += 8
		if size < 0 || offset+size > len(data) {
			return nil, fmt.Errorf("invalid %q chunk size", id)
		}

		chunk := data[offset : offset+size]
		switch id {
		case "fmt ":
			if len(chunk) < 16 {
				return nil, errors.New("fmt chunk is too short")
			}
			format = binary.LittleEndian.Uint16(chunk[0:2])
			channels = binary.LittleEndian.Uint16(chunk[2:4])
			sampleRate = binary.LittleEndian.Uint32(chunk[4:8])
			bitsPerSample = binary.LittleEndian.Uint16(chunk[14:16])
		case "data":
			dataChunk = chunk
		}

		offset += size
		if size%2 == 1 {
			offset++
		}
	}

	if format != 1 {
		return nil, fmt.Errorf("unsupported WAV format %d: only PCM is supported", format)
	}
	if channels == 0 || sampleRate == 0 {
		return nil, errors.New("WAV file is missing channel or sample rate data")
	}
	if dataChunk == nil {
		return nil, errors.New("WAV file has no data chunk")
	}

	bytesPerSample := int(bitsPerSample / 8)
	if bytesPerSample == 0 || bitsPerSample%8 != 0 {
		return nil, fmt.Errorf("unsupported bit depth %d", bitsPerSample)
	}
	frameSize := int(channels) * bytesPerSample
	if frameSize == 0 || len(dataChunk)%frameSize != 0 {
		return nil, errors.New("WAV data chunk is not aligned to full sample frames")
	}

	mono := make([]float64, 0, len(dataChunk)/frameSize)
	for offset := 0; offset < len(dataChunk); offset += frameSize {
		var sum float64
		for ch := 0; ch < int(channels); ch++ {
			sampleOffset := offset + ch*bytesPerSample
			sum += decodeSample(dataChunk[sampleOffset:sampleOffset+bytesPerSample], bitsPerSample)
		}
		mono = append(mono, sum/float64(channels))
	}

	return &Audio{
		Channels:      int(channels),
		SampleRate:    int(sampleRate),
		BitsPerSample: int(bitsPerSample),
		Mono:          mono,
	}, nil
}

func decodeSample(sample []byte, bits uint16) float64 {
	switch bits {
	case 8:
		return (float64(sample[0]) - 128) / 128
	case 16:
		v := int16(binary.LittleEndian.Uint16(sample))
		return float64(v) / float64(math.MaxInt16)
	case 24:
		v := int32(sample[0]) | int32(sample[1])<<8 | int32(sample[2])<<16
		if v&0x800000 != 0 {
			v |= ^0xffffff
		}
		return float64(v) / 8388607
	case 32:
		v := int32(binary.LittleEndian.Uint32(sample))
		return float64(v) / float64(math.MaxInt32)
	default:
		return 0
	}
}
