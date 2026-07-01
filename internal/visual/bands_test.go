package visual

import "testing"

func TestSplitBands(t *testing.T) {
	bars := []float64{1, 1, 1, 0, 0, 0}
	bass, mid, treble := SplitBands(bars)
	if bass <= treble {
		t.Fatalf("bass=%f treble=%f, want bass > treble", bass, treble)
	}
	if mid != 0.5 {
		t.Fatalf("mid=%f, want 0.5", mid)
	}
}
