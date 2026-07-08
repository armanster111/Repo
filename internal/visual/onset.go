package visual

import "math"

// OnsetEngine tracks per-band energy spikes for kick, snare, and hi-hat.
type OnsetEngine struct {
	bassHist   []float64
	midHist    []float64
	trebleHist []float64
	Kick       float64
	Snare      float64
	HiHat      float64
	Phase      float64
}

// NewOnsetEngine creates a band onset tracker.
func NewOnsetEngine() *OnsetEngine {
	return &OnsetEngine{}
}

// Update advances onset detection from spectrum bars and optional BPM.
func (o *OnsetEngine) Update(bars []float64, bpm float64, dt float64) {
	if len(bars) == 0 {
		return
	}
	bass, mid, treble := SplitBands(bars)
	o.Kick = decayPulse(o.Kick, dt, 4.5) + bandOnset(&o.bassHist, bass, 1.6)
	o.Snare = decayPulse(o.Snare, dt, 5.5) + bandOnset(&o.midHist, mid, 1.5)
	o.HiHat = decayPulse(o.HiHat, dt, 8.0) + bandOnset(&o.trebleHist, treble, 1.4)
	if bpm > 0 {
		beatPeriod := 60.0 / bpm
		o.Phase += dt / beatPeriod
		o.Phase -= math.Floor(o.Phase)
	}
}

func decayPulse(v, dt, rate float64) float64 {
	return v * math.Exp(-rate*dt)
}

func bandOnset(hist *[]float64, current float64, gain float64) float64 {
	*hist = append(*hist, current)
	if len(*hist) > 24 {
		*hist = (*hist)[len(*hist)-24:]
	}
	var avg float64
	for _, v := range *hist {
		avg += v
	}
	avg /= float64(len(*hist))
	if current <= avg*1.22 {
		return 0
	}
	return math.Min(1, (current-avg)*gain)
}

// CombinedBeat returns a unified beat multiplier from all bands.
func (o *OnsetEngine) CombinedBeat() float64 {
	return 1 + o.Kick*0.45 + o.Snare*0.25 + o.HiHat*0.12
}
