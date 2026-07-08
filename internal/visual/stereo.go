package visual

import "math"

// StereoSample holds left/right PCM in [-1, 1].
type StereoSample struct {
	Left  float64
	Right float64
}

// Downmix averages stereo channels to mono.
func Downmix(left, right []float64) []float64 {
	n := len(left)
	if len(right) < n {
		n = len(right)
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = (left[i] + right[i]) / 2
	}
	return out
}

// AnalyzeStereoWindow runs FFT on mono downmix and returns pan (-1..1) and width.
func AnalyzeStereoWindow(left, right []float64, sampleRate, barCount int) ([]float64, float64, float64) {
	mono := Downmix(left, right)
	bars := AnalyzeWindow(mono, sampleRate, barCount)
	if bars == nil {
		return nil, 0, 0
	}
	var lEnergy, rEnergy float64
	n := min(len(left), len(right))
	for i := 0; i < n; i++ {
		lEnergy += left[i] * left[i]
		rEnergy += right[i] * right[i]
	}
	total := lEnergy + rEnergy
	if total < 1e-9 {
		return bars, 0, 0
	}
	pan := (rEnergy - lEnergy) / total
	width := 1 - math.Abs(pan)
	return bars, pan, width
}

// WaveformSlice returns normalized waveform points from mono PCM.
func WaveformSlice(samples []float64, start, count int) []float64 {
	if len(samples) == 0 || count <= 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}
	if start >= len(samples) {
		start = len(samples) - 1
	}
	end := start + count
	if end > len(samples) {
		end = len(samples)
	}
	out := make([]float64, count)
	avail := end - start
	for i := 0; i < count; i++ {
		if avail <= 0 {
			break
		}
		idx := start + i*avail/count
		if idx >= end {
			idx = end - 1
		}
		out[i] = samples[idx]
	}
	return out
}

// LissajousPoints returns paired L/R samples for oscilloscope-style display.
func LissajousPoints(left, right []float64, start, count int) []StereoSample {
	n := min(len(left), len(right))
	if n == 0 || count <= 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}
	if start >= n {
		start = n - 1
	}
	end := start + count
	if end > n {
		end = n
	}
	avail := end - start
	out := make([]StereoSample, count)
	for i := 0; i < count; i++ {
		if avail <= 0 {
			break
		}
		idx := start + i*avail/count
		if idx >= end {
			idx = end - 1
		}
		out[i] = StereoSample{Left: left[idx], Right: right[idx]}
	}
	return out
}
