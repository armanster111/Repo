package visual

// PCMCache holds decoded audio for live FFT and waveform modes.
type PCMCache struct {
	Mono       []float64
	Left       []float64
	Right      []float64
	SampleRate int
}

// WindowAt extracts a mono PCM window centered at sample index pos.
func (p *PCMCache) WindowAt(pos, windowSize int) []float64 {
	if p == nil || len(p.Mono) == 0 || windowSize <= 0 {
		return nil
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= len(p.Mono) {
		pos = len(p.Mono) - 1
	}
	start := pos - windowSize/2
	if start < 0 {
		start = 0
	}
	end := start + windowSize
	if end > len(p.Mono) {
		end = len(p.Mono)
		start = end - windowSize
		if start < 0 {
			start = 0
		}
	}
	buf := make([]float64, windowSize)
	copy(buf[windowSize-(end-start):], p.Mono[start:end])
	return buf
}

// StereoWindowAt extracts left/right windows centered at pos.
func (p *PCMCache) StereoWindowAt(pos, windowSize int) (left, right []float64) {
	if p == nil || windowSize <= 0 {
		return nil, nil
	}
	if len(p.Left) > 0 && len(p.Right) > 0 {
		left = windowSlice(p.Left, pos, windowSize)
		right = windowSlice(p.Right, pos, windowSize)
		return left, right
	}
	mono := p.WindowAt(pos, windowSize)
	return mono, mono
}

func windowSlice(samples []float64, pos, windowSize int) []float64 {
	if len(samples) == 0 {
		return nil
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= len(samples) {
		pos = len(samples) - 1
	}
	start := pos - windowSize/2
	if start < 0 {
		start = 0
	}
	end := start + windowSize
	if end > len(samples) {
		end = len(samples)
		start = end - windowSize
		if start < 0 {
			start = 0
		}
	}
	buf := make([]float64, windowSize)
	copy(buf[windowSize-(end-start):], samples[start:end])
	return buf
}

// SampleIndex converts playback position in seconds to a sample index.
func (p *PCMCache) SampleIndex(seconds float64) int {
	if p == nil || p.SampleRate <= 0 {
		return 0
	}
	idx := int(seconds * float64(p.SampleRate))
	if idx < 0 {
		return 0
	}
	if idx >= len(p.Mono) {
		return len(p.Mono) - 1
	}
	return idx
}
