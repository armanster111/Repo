package analysis

import "testing"

func TestDetectMood(t *testing.T) {
	bars := make([]float64, 48)
	for i := range bars {
		if i < 16 {
			bars[i] = 0.9
		}
	}
	if DetectMood(bars) != MoodDark && DetectMood(bars) != MoodEnergetic {
		t.Fatalf("unexpected mood for bass-heavy bars")
	}
}

func TestApplyVisualEQ(t *testing.T) {
	bars := []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5}
	out := ApplyVisualEQ(bars, 1000, 0, 0)
	if out[0] <= bars[0] {
		t.Fatalf("expected bass boost, got %v", out[0])
	}
}
