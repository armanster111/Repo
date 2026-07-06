package visual

// BuildFramesFFT converts mono audio into FFT-based spectrum frames.
func BuildFramesFFT(samples []float64, sampleRate, fps, barCount int) []Frame {
	if len(samples) == 0 || sampleRate <= 0 || fps <= 0 || barCount <= 0 {
		return nil
	}
	windowSize := nextPow2(sampleRate / fps * 2)
	if windowSize < 512 {
		windowSize = 512
	}
	if windowSize > 4096 {
		windowSize = 4096
	}
	hop := max(1, sampleRate/fps)
	frameCount := int(float64(len(samples)-windowSize)/float64(hop)) + 1
	if frameCount < 1 {
		frameCount = 1
	}
	frames := make([]Frame, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		start := i * hop
		end := start + windowSize
		if start >= len(samples) {
			break
		}
		if end > len(samples) {
			end = len(samples)
		}
		window := make([]float64, windowSize)
		copy(window, samples[start:end])
		bars := AnalyzeWindow(window, sampleRate, barCount)
		if bars == nil {
			bars = make([]float64, barCount)
		}
		frames = append(frames, Frame{Bars: bars})
	}
	return frames
}
