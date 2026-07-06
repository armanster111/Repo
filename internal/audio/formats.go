package audio

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/jfreymuth/oggvorbis"
	"github.com/mewkiz/flac"
)

// DecodeFileFLAC decodes FLAC to mono samples.
func DecodeFileFLAC(path string) ([]float64, int, error) {
	stream, err := flac.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer stream.Close()
	info := stream.Info
	if info == nil {
		return nil, 0, os.ErrInvalid
	}
	sampleRate := int(info.SampleRate)
	channels := int(info.NChannels)
	var mono []float64
	for {
		frame, err := stream.ParseNext()
		if err != nil {
			break
		}
		samples := frame.Subframes[0].Samples
		if channels > 1 && len(frame.Subframes) > 1 {
			right := frame.Subframes[1].Samples
			for i := range samples {
				l := float64(samples[i])
				r := 0.0
				if i < len(right) {
					r = float64(right[i])
				}
				mono = append(mono, (l+r)/2/32768)
			}
		} else {
			for _, s := range samples {
				mono = append(mono, float64(s)/32768)
			}
		}
	}
	return mono, sampleRate, nil
}

// DecodeFileOGG decodes OGG Vorbis to mono samples.
func DecodeFileOGG(path string) ([]float64, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	return DecodeOGG(f)
}

// DecodeOGG decodes OGG from a reader.
func DecodeOGG(r io.Reader) ([]float64, int, error) {
	dec, err := oggvorbis.NewReader(r)
	if err != nil {
		return nil, 0, err
	}
	sampleRate := dec.SampleRate()
	buf := make([]float32, 4096)
	var mono []float64
	for {
		n, err := dec.Read(buf)
		if n > 0 {
			for i := 0; i < n; i += 2 {
				if i+1 < n {
					mono = append(mono, float64((buf[i]+buf[i+1])/2))
				} else {
					mono = append(mono, float64(buf[i]))
				}
			}
		}
		if err != nil {
			break
		}
	}
	// Normalize if loud
	var peak float64
	for _, v := range mono {
		peak = math.Max(peak, math.Abs(v))
	}
	if peak > 1 {
		for i := range mono {
			mono[i] /= peak
		}
	}
	return mono, sampleRate, nil
}

// DecodeAny attempts format-specific decode for visual analysis.
func DecodeAny(path string) ([]float64, int, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".flac":
		return DecodeFileFLAC(path)
	case ".ogg", ".oga":
		return DecodeFileOGG(path)
	case ".mp3":
		return DecodeFileMP3(path)
	default:
		return nil, 0, os.ErrNotExist
	}
}
