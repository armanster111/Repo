package visual

import "math"

// FFT computes magnitudes for real input using a radix-2 Cooley-Tukey transform.
// Input length must be a power of two; shorter inputs are zero-padded.
func FFT(input []float64) []float64 {
	n := nextPow2(len(input))
	if n < 2 {
		return nil
	}
	re := make([]float64, n)
	im := make([]float64, n)
	copy(re, input)
	fftInPlace(re, im)
	mags := make([]float64, n/2)
	for i := range mags {
		mags[i] = math.Hypot(re[i], im[i]) / float64(n)
	}
	return mags
}

func fftInPlace(re, im []float64) {
	n := len(re)
	for i, j := 1, 0; i < n; i++ {
		bit := n >> 1
		for j&bit != 0 {
			j &= ^bit
			bit >>= 1
		}
		j |= bit
		if i < j {
			re[i], re[j] = re[j], re[i]
			im[i], im[j] = im[j], im[i]
		}
	}
	for length := 2; length <= n; length <<= 1 {
		angle := -2 * math.Pi / float64(length)
		wlenRe := math.Cos(angle)
		wlenIm := math.Sin(angle)
		for i := 0; i < n; i += length {
			wRe, wIm := 1.0, 0.0
			for j := 0; j < length/2; j++ {
				uRe := re[i+j]
				uIm := im[i+j]
				vRe := re[i+j+length/2]*wRe - im[i+j+length/2]*wIm
				vIm := re[i+j+length/2]*wIm + im[i+j+length/2]*wRe
				re[i+j] = uRe + vRe
				im[i+j] = uIm + vIm
				re[i+j+length/2] = uRe - vRe
				im[i+j+length/2] = uIm - vIm
				nextWRe := wRe*wlenRe - wIm*wlenIm
				wIm = wRe*wlenIm + wIm*wlenRe
				wRe = nextWRe
			}
		}
	}
}

func nextPow2(n int) int {
	if n <= 1 {
		return 2
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

// HanningWindow applies a Hann window in-place.
func HanningWindow(samples []float64) {
	n := len(samples)
	if n == 0 {
		return
	}
	for i := range samples {
		w := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
		samples[i] *= w
	}
}

// SpectrumBars maps FFT magnitudes to logarithmically spaced visual bars.
func SpectrumBars(magnitudes []float64, sampleRate, barCount int) []float64 {
	if len(magnitudes) == 0 || sampleRate <= 0 || barCount <= 0 {
		return nil
	}
	bars := make([]float64, barCount)
	nyquist := float64(sampleRate) / 2
	minHz := 40.0
	maxHz := math.Min(16000, nyquist)
	if maxHz <= minHz {
		maxHz = nyquist
	}
	for b := 0; b < barCount; b++ {
		t0 := float64(b) / float64(barCount)
		t1 := float64(b+1) / float64(barCount)
		fLow := minHz * math.Pow(maxHz/minHz, t0)
		fHigh := minHz * math.Pow(maxHz/minHz, t1)
		binLow := int(fLow / nyquist * float64(len(magnitudes)))
		binHigh := int(fHigh / nyquist * float64(len(magnitudes)))
		if binLow < 0 {
			binLow = 0
		}
		if binHigh <= binLow {
			binHigh = binLow + 1
		}
		if binHigh > len(magnitudes) {
			binHigh = len(magnitudes)
		}
		var sum float64
		for i := binLow; i < binHigh; i++ {
			sum += magnitudes[i]
		}
		bars[b] = sum / float64(binHigh-binLow)
	}
	normalizeBars(bars)
	for i := range bars {
		bars[i] = math.Pow(bars[i], 0.82)
	}
	return bars
}

// AnalyzeWindow runs FFT spectrum analysis on a PCM window.
func AnalyzeWindow(samples []float64, sampleRate, barCount int) []float64 {
	if len(samples) == 0 {
		return nil
	}
	n := nextPow2(len(samples))
	buf := make([]float64, n)
	copy(buf, samples)
	HanningWindow(buf)
	mags := FFT(buf)
	return SpectrumBars(mags, sampleRate, barCount)
}
