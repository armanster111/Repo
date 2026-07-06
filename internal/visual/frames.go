package visual

import "math"

// Frame is one animation frame of normalized bar heights.
type Frame struct {
	Bars []float64
}

// BuildFrames converts mono audio samples into RMS bar frames.
func BuildFrames(samples []float64, sampleRate, fps, barCount int) []Frame {
	if len(samples) == 0 || sampleRate <= 0 || fps <= 0 || barCount <= 0 {
		return nil
	}

	samplesPerFrame := max(1, sampleRate/fps)
	frameCount := int(math.Ceil(float64(len(samples)) / float64(samplesPerFrame)))
	frames := make([]Frame, 0, frameCount)

	for frameIndex := 0; frameIndex < frameCount; frameIndex++ {
		start := frameIndex * samplesPerFrame
		end := min(len(samples), start+samplesPerFrame)
		if start >= end {
			break
		}

		bars := make([]float64, barCount)
		for bar := 0; bar < barCount; bar++ {
			segStart := start + (end-start)*bar/barCount
			segEnd := start + (end-start)*(bar+1)/barCount
			if segEnd <= segStart {
				segEnd = min(end, segStart+1)
			}
			bars[bar] = rms(samples[segStart:segEnd])
		}
		normalizeBars(bars)
		frames = append(frames, Frame{Bars: bars})
	}

	return frames
}

func rms(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, sample := range samples {
		sum += sample * sample
	}
	return math.Sqrt(sum / float64(len(samples)))
}

func normalizeBars(bars []float64) {
	var peak float64
	for _, value := range bars {
		peak = max(peak, value)
	}
	if peak == 0 {
		return
	}
	for i, value := range bars {
		bars[i] = math.Min(1, value/peak)
	}
}
