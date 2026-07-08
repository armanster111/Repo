package visual

import "testing"

func TestOnsetEngineKick(t *testing.T) {
	o := NewOnsetEngine()
	bars := make([]float64, 64)
	for i := range bars {
		bars[i] = 0.1
	}
	o.Update(bars, 120, 1.0/60.0)
	if o.Kick < 0 {
		t.Fatal("kick should be non-negative")
	}
	for i := 0; i < 10; i++ {
		bars[i] = 1.0
	}
	o.Update(bars, 120, 1.0/60.0)
	if o.Kick <= 0 {
		t.Fatal("expected kick spike on bass burst")
	}
}

func TestCombinedBeat(t *testing.T) {
	o := NewOnsetEngine()
	o.Kick = 0.5
	if o.CombinedBeat() <= 1 {
		t.Fatal("combined beat should exceed 1 with kick energy")
	}
}
