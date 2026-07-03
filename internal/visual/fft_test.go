package visual

import (
	"math"
	"testing"
)

func TestFFTImpulse(t *testing.T) {
	input := make([]float64, 64)
	input[0] = 1
	mags := FFT(input)
	if len(mags) != 32 {
		t.Fatalf("expected 32 bins, got %d", len(mags))
	}
	if mags[0] <= 0 {
		t.Fatal("dc bin should be non-zero")
	}
}

func TestSpectrumBarsShape(t *testing.T) {
	mags := make([]float64, 512)
	for i := range mags {
		mags[i] = 1.0 / float64(i+1)
	}
	bars := SpectrumBars(mags, 44100, 64)
	if len(bars) != 64 {
		t.Fatalf("expected 64 bars, got %d", len(bars))
	}
	if bars[0] <= bars[len(bars)-1] {
		t.Fatal("low frequencies should dominate in test spectrum")
	}
}

func TestAnalyzeWindow(t *testing.T) {
	samples := make([]float64, 2048)
	for i := range samples {
		samples[i] = math.Sin(float64(i) * 0.15)
	}
	bars := AnalyzeWindow(samples, 44100, 32)
	if len(bars) != 32 {
		t.Fatalf("expected 32 bars, got %d", len(bars))
	}
}

func TestSmootherAttackDecay(t *testing.T) {
	s := NewSmoother(8, 0.9, 0.2, 0.5)
	out := s.Step([]float64{1, 1, 1, 1, 1, 1, 1, 1}, 1.0/60.0)
	for _, v := range out {
		if v <= 0 {
			t.Fatal("expected rising values")
		}
	}
}
