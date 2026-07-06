package visual

import "math"

// Smoother applies attack/decay envelope following per bar.
type Smoother struct {
	Attack  float64
	Decay   float64
	PeakHold float64
	values  []float64
	peaks   []float64
	peakAge []float64
}

// NewSmoother creates a bar envelope follower.
func NewSmoother(barCount int, attack, decay, peakHold float64) *Smoother {
	return &Smoother{
		Attack:   attack,
		Decay:    decay,
		PeakHold: peakHold,
		values:   make([]float64, barCount),
		peaks:    make([]float64, barCount),
		peakAge:  make([]float64, barCount),
	}
}

// Step advances smoothing toward target bars (dt in seconds).
func (s *Smoother) Step(target []float64, dt float64) []float64 {
	if len(target) == 0 {
		return s.values
	}
	if len(s.values) != len(target) {
		s.values = make([]float64, len(target))
		s.peaks = make([]float64, len(target))
		s.peakAge = make([]float64, len(target))
	}
	for i, t := range target {
		rate := s.Decay
		if t > s.values[i] {
			rate = s.Attack
		}
		alpha := 1 - math.Exp(-rate*dt*60)
		s.values[i] += (t - s.values[i]) * alpha
		if s.values[i] > s.peaks[i] {
			s.peaks[i] = s.values[i]
			s.peakAge[i] = 0
		} else {
			s.peakAge[i] += dt
			if s.peakAge[i] > s.PeakHold {
				s.peaks[i] += (s.values[i] - s.peaks[i]) * 0.15
			}
		}
	}
	out := make([]float64, len(s.values))
	copy(out, s.values)
	return out
}

// Peaks returns current peak-hold values.
func (s *Smoother) Peaks() []float64 {
	out := make([]float64, len(s.peaks))
	copy(out, s.peaks)
	return out
}

// TrailBuffer stores faded history for motion trails.
type TrailBuffer struct {
	frames [][]float64
	depth  int
}

// NewTrailBuffer creates a trail buffer with depth history frames.
func NewTrailBuffer(depth, barCount int) *TrailBuffer {
	return &TrailBuffer{depth: depth, frames: make([][]float64, 0, depth)}
}

// Frames returns stored trail frames.
func (t *TrailBuffer) Frames() [][]float64 {
	return t.frames
}

// Push adds a frame.
func (t *TrailBuffer) Push(bars []float64) {
	if len(bars) == 0 {
		return
	}
	frame := make([]float64, len(bars))
	copy(frame, bars)
	t.frames = append([][]float64{frame}, t.frames...)
	if len(t.frames) > t.depth {
		t.frames = t.frames[:t.depth]
	}
}

// OnsetBeat detects beats via spectral flux.
func OnsetBeat(prev, current []float64, history []float64) (float64, []float64) {
	if len(prev) == 0 || len(current) == 0 {
		return 1, history
	}
	var flux float64
	n := min(len(prev), len(current))
	for i := 0; i < n; i++ {
		d := current[i] - prev[i]
		if d > 0 {
			flux += d
		}
	}
	flux /= float64(n)
	history = append(history, flux)
	if len(history) > 43 {
		history = history[len(history)-43:]
	}
	var avg float64
	for _, v := range history {
		avg += v
	}
	avg /= float64(len(history))
	mult := 1.0
	if flux > avg*1.4 {
		mult = 1 + math.Min(0.55, (flux-avg)*2.2)
	}
	return mult, history
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
