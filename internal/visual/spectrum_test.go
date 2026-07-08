package visual

import "testing"

func TestMelScaleMonotonic(t *testing.T) {
	m1 := HzToMel(100)
	m2 := HzToMel(1000)
	m3 := HzToMel(8000)
	if m1 >= m2 || m2 >= m3 {
		t.Fatal("mel scale should increase with frequency")
	}
	if MelToHz(HzToMel(440)) < 430 || MelToHz(HzToMel(440)) > 450 {
		t.Fatal("mel round-trip should be near original frequency")
	}
}

func TestAWeightPeak(t *testing.T) {
	w1k := AWeight(1000)
	w100 := AWeight(100)
	if w1k <= w100 {
		t.Fatal("A-weighting should emphasize mid frequencies over low")
	}
}

func TestSpectrumMelBars(t *testing.T) {
	mags := make([]float64, 512)
	for i := range mags {
		mags[i] = 1.0 / float64(i+1)
	}
	bars := SpectrumMelBars(mags, 44100, 64)
	if len(bars) != 64 {
		t.Fatalf("expected 64 bars, got %d", len(bars))
	}
	var sum float64
	for _, v := range bars {
		sum += v
	}
	if sum <= 0 {
		t.Fatal("mel bars should have energy")
	}
}

func TestPCMCacheWindow(t *testing.T) {
	cache := &PCMCache{
		Mono:       []float64{0, 0.5, 1, 0.5, 0, -0.5, -1, -0.5},
		SampleRate: 8,
	}
	win := cache.WindowAt(4, 4)
	if len(win) != 4 {
		t.Fatalf("expected window size 4, got %d", len(win))
	}
}

func TestDownmix(t *testing.T) {
	left := []float64{1, -1}
	right := []float64{-1, 1}
	mono := Downmix(left, right)
	if mono[0] != 0 || mono[1] != 0 {
		t.Fatalf("unexpected downmix: %v", mono)
	}
}
