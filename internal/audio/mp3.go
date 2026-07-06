package audio

import (
	"io"
	"math"
	"os"

	"github.com/hajimehoshi/go-mp3"
)

// DecodeFileMP3 decodes an MP3 file into normalized mono samples.
func DecodeFileMP3(path string) ([]float64, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	return DecodeMP3(f)
}

// DecodeMP3 decodes MP3 PCM into mono float samples in [-1, 1].
func DecodeMP3(r io.Reader) ([]float64, int, error) {
	dec, err := mp3.NewDecoder(r)
	if err != nil {
		return nil, 0, err
	}
	sampleRate := dec.SampleRate()
	pcm, err := io.ReadAll(dec)
	if err != nil {
		return nil, 0, err
	}
	if len(pcm) < 2 {
		return nil, sampleRate, nil
	}

	samples := make([]float64, 0, len(pcm)/4)
	for i := 0; i+3 < len(pcm); i += 4 {
		left := int16(pcm[i]) | int16(pcm[i+1])<<8
		right := int16(pcm[i+2]) | int16(pcm[i+3])<<8
		samples = append(samples, (float64(left)+float64(right))/2/float64(math.MaxInt16))
	}
	return samples, sampleRate, nil
}
