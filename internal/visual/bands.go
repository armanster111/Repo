package visual

import "math"

// SplitBands returns bass, mid, and treble energy from normalized bars.
func SplitBands(bars []float64) (bass, mid, treble float64) {
	if len(bars) == 0 {
		return 0, 0, 0
	}
	third := len(bars) / 3
	if third == 0 {
		third = 1
	}
	for i, v := range bars {
		switch {
		case i < third:
			bass += v
		case i < third*2:
			mid += v
		default:
			treble += v
		}
	}
	count := float64(third)
	return bass / count, mid / count, treble / count
}

// BeatPulse returns a multiplier from recent energy spikes.
func BeatPulse(history []float64, current float64) float64 {
	if len(history) == 0 {
		return 1
	}
	var avg float64
	for _, v := range history {
		avg += v
	}
	avg /= float64(len(history))
	if current > avg*1.35 {
		return 1 + math.Min(0.45, (current-avg)*1.8)
	}
	return 1
}
