package visual

import "testing"

func TestBuildFramesCreatesNormalizedBars(t *testing.T) {
	samples := []float64{0.1, -0.1, 0.4, -0.4, 0.8, -0.8, 0.2, -0.2}

	frames := BuildFrames(samples, 8, 2, 2)
	if len(frames) != 2 {
		t.Fatalf("len(frames) = %d, want 2", len(frames))
	}

	for _, frame := range frames {
		if len(frame.Bars) != 2 {
			t.Fatalf("len(frame.Bars) = %d, want 2", len(frame.Bars))
		}
		for _, bar := range frame.Bars {
			if bar < 0 || bar > 1 {
				t.Fatalf("bar = %f, want normalized value in [0, 1]", bar)
			}
		}
	}
}

func TestBuildFramesRejectsEmptyInput(t *testing.T) {
	if frames := BuildFrames(nil, 44100, 30, 48); frames != nil {
		t.Fatalf("BuildFrames() = %#v, want nil", frames)
	}
}
