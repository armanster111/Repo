package visual

import "math"

// HzToMel converts frequency in Hz to the mel scale.
func HzToMel(hz float64) float64 {
	return 2595 * math.Log10(1+hz/700)
}

// MelToHz converts mel back to Hz.
func MelToHz(mel float64) float64 {
	return 700 * (math.Pow(10, mel/2595) - 1)
}

// AWeight approximates IEC A-weighting at frequency hz.
func AWeight(hz float64) float64 {
	if hz <= 0 {
		return 0
	}
	f2 := hz * hz
	numerator := 12194 * 12194 * f2 * f2
	denom := (f2 + 20.6*20.6) *
		math.Sqrt((f2+107.7*107.7)*(f2+737.9*737.9)) *
		(f2 + 12194*12194)
	if denom == 0 {
		return 1
	}
	w := numerator / denom
	return w / 0.794 // normalize so 1 kHz ≈ 1.0
}

// SpectrumMelBars maps FFT magnitudes to mel-spaced bars with perceptual weighting.
func SpectrumMelBars(magnitudes []float64, sampleRate, barCount int) []float64 {
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
	minMel := HzToMel(minHz)
	maxMel := HzToMel(maxHz)
	for b := 0; b < barCount; b++ {
		t0 := float64(b) / float64(barCount)
		t1 := float64(b+1) / float64(barCount)
		fLow := MelToHz(minMel + (maxMel-minMel)*t0)
		fHigh := MelToHz(minMel + (maxMel-minMel)*t1)
		centerHz := (fLow + fHigh) / 2
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
		bars[b] = (sum / float64(binHigh-binLow)) * AWeight(centerHz)
	}
	perceptualNormalize(bars)
	return bars
}

// perceptualNormalize converts magnitudes to a perceptual dB-like scale.
func perceptualNormalize(bars []float64) {
	const floor = 1e-6
	var peak float64
	for i, v := range bars {
		if v < floor {
			v = floor
		}
		db := 20 * math.Log10(v)
		bars[i] = db
		if db > peak {
			peak = db
		}
	}
	if peak <= -60 {
		return
	}
	minDB := peak - 48
	for i, db := range bars {
		norm := (db - minDB) / (peak - minDB)
		if norm < 0 {
			norm = 0
		}
		bars[i] = math.Pow(norm, 0.78)
	}
}

// SpectrumBars maps FFT magnitudes to logarithmically spaced visual bars.
func SpectrumBars(magnitudes []float64, sampleRate, barCount int) []float64 {
	return SpectrumMelBars(magnitudes, sampleRate, barCount)
}
