package analysis

import (
	"math"
	"time"

	"github.com/armanster111/muse/internal/visual"
)

// Mood describes detected audio character.
type Mood string

const (
	MoodChill     Mood = "Chill"
	MoodGroove    Mood = "Groove"
	MoodEnergetic Mood = "Energetic"
	MoodBright    Mood = "Bright"
	MoodDark      Mood = "Dark"
)

// DetectMood estimates mood from frequency band energy.
func DetectMood(bars []float64) Mood {
	bass, mid, treble := visual.SplitBands(bars)
	total := bass + mid + treble
	if total < 0.05 {
		return MoodChill
	}
	bass /= total
	mid /= total
	treble /= total
	switch {
	case bass > 0.45 && mid > 0.3:
		return MoodEnergetic
	case treble > 0.4:
		return MoodBright
	case bass > 0.5:
		return MoodDark
	case mid > 0.42:
		return MoodGroove
	default:
		return MoodChill
	}
}

// EstimateBPM returns a rough beats-per-minute from energy spikes.
func EstimateBPM(history []float64, sampleInterval time.Duration) float64 {
	if len(history) < 6 {
		return 0
	}
	var avg float64
	for _, v := range history {
		avg += v
	}
	avg /= float64(len(history))
	threshold := avg * 1.25
	var beats []int
	for i, v := range history {
		if v >= threshold {
			beats = append(beats, i)
		}
	}
	if len(beats) < 2 {
		return 0
	}
	var intervals []float64
	for i := 1; i < len(beats); i++ {
		intervals = append(intervals, float64(beats[i]-beats[i-1]))
	}
	mean := average(intervals)
	if mean <= 0 {
		return 0
	}
	secondsPerBeat := mean * sampleInterval.Seconds()
	if secondsPerBeat <= 0 {
		return 0
	}
	bpm := 60 / secondsPerBeat
	if bpm < 60 {
		bpm *= 2
	}
	if bpm > 180 {
		bpm /= 2
	}
	return math.Round(bpm)
}

// ApplyVisualEQ adjusts bar heights using a 3-band software EQ.
func ApplyVisualEQ(bars []float64, bass, mid, treble int) []float64 {
	if len(bars) == 0 {
		return bars
	}
	out := make([]float64, len(bars))
	third := len(bars) / 3
	if third == 0 {
		third = 1
	}
	bassGain := 1 + float64(clampInt(bass, -1000, 1000))/2000
	midGain := 1 + float64(clampInt(mid, -1000, 1000))/2000
	trebleGain := 1 + float64(clampInt(treble, -1000, 1000))/2000
	for i, v := range bars {
		switch {
		case i < third:
			out[i] = clamp(v*bassGain, 0, 1)
		case i < third*2:
			out[i] = clamp(v*midGain, 0, 1)
		default:
			out[i] = clamp(v*trebleGain, 0, 1)
		}
	}
	return out
}

func average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
