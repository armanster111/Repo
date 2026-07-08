//go:build windows

package main

import (
	"math"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/armanster111/music-visualizer/internal/analysis"
	"github.com/armanster111/music-visualizer/internal/audio"
	"github.com/armanster111/music-visualizer/internal/visual"
	"github.com/armanster111/music-visualizer/internal/wav"
)

type inputSource int

const (
	inputPlayer inputSource = iota
	inputDesktop
	inputMic
)

type overlayAspect int

const (
	overlayAspectFree overlayAspect = iota
	overlayAspect16x9
	overlayAspect4x3
	overlayAspect1x1
)

const specHistoryDepth = 128

func (s *appState) loadPCMForPath(path string) *visual.PCMCache {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".wav":
		data, err := wav.DecodeFile(path)
		if err != nil || data == nil {
			return nil
		}
		left, right := wavDecodeStereo(path)
		return &visual.PCMCache{Mono: data.Mono, Left: left, Right: right, SampleRate: data.SampleRate}
	case ".mp3":
		mono, sr, err := audio.DecodeFileMP3(path)
		if err != nil || len(mono) == 0 {
			return nil
		}
		return &visual.PCMCache{Mono: mono, Left: mono, Right: mono, SampleRate: sr}
	case ".flac":
		mono, sr, err := audio.DecodeFileFLAC(path)
		if err != nil || len(mono) == 0 {
			return nil
		}
		return &visual.PCMCache{Mono: mono, Left: mono, Right: mono, SampleRate: sr}
	case ".ogg", ".oga":
		mono, sr, err := audio.DecodeFileOGG(path)
		if err != nil || len(mono) == 0 {
			return nil
		}
		return &visual.PCMCache{Mono: mono, Left: mono, Right: mono, SampleRate: sr}
	case ".aiff", ".aif":
		mono, sr, err := audio.DecodeFileAIFF(path)
		if err != nil || len(mono) == 0 {
			return nil
		}
		return &visual.PCMCache{Mono: mono, Left: mono, Right: mono, SampleRate: sr}
	default:
		return nil
	}
}

func wavDecodeStereo(path string) (left, right []float64) {
	data, err := wav.DecodeFile(path)
	if err != nil || data == nil {
		return nil, nil
	}
	return data.Mono, data.Mono
}

func (s *appState) liveFFTBars() []float64 {
	if s.pcmCache == nil || len(s.pcmCache.Mono) == 0 {
		return nil
	}
	windowSize := 2048
	if s.pcmCache.SampleRate > 0 {
		windowSize = visual.NextPow2(s.pcmCache.SampleRate / fps * 2)
		if windowSize < 512 {
			windowSize = 512
		}
		if windowSize > 4096 {
			windowSize = 4096
		}
	}
	idx := s.pcmCache.SampleIndex(s.position().Seconds())
	window := s.pcmCache.WindowAt(idx, windowSize)
	if window == nil {
		return nil
	}
	bars := visual.AnalyzeWindow(window, s.pcmCache.SampleRate, barCount)
	return bars
}

func (s *appState) pushSpecHistory(bars []float64) {
	if s.mode != modeSpectrogram || len(bars) == 0 {
		return
	}
	if len(s.specHistory) >= specHistoryDepth {
		s.specHistory = s.specHistory[1:]
	}
	frame := make([]float64, len(bars))
	copy(frame, bars)
	s.specHistory = append(s.specHistory, frame)
}

func (s *appState) moodAdjustedPalette(base palette) palette {
	if !s.moodReactive || s.mood == "" {
		return base
	}
	switch analysis.Mood(s.mood) {
	case analysis.MoodEnergetic:
		return palette{
			background: base.background,
			panel:      base.panel,
			text:       base.text,
			dim:        base.dim,
			accent:     blendColor(base.accent, rgb(255, 80, 40), 0.35),
			accent2:    blendColor(base.accent2, rgb(255, 200, 60), 0.35),
		}
	case analysis.MoodChill:
		return palette{
			background: base.background,
			panel:      base.panel,
			text:       base.text,
			dim:        base.dim,
			accent:     blendColor(base.accent, rgb(80, 160, 220), 0.4),
			accent2:    blendColor(base.accent2, rgb(120, 200, 180), 0.35),
		}
	case analysis.MoodBright:
		return palette{
			background: base.background,
			panel:      base.panel,
			text:       base.text,
			dim:        base.dim,
			accent:     blendColor(base.accent, rgb(255, 240, 120), 0.3),
			accent2:    blendColor(base.accent2, rgb(200, 255, 200), 0.3),
		}
	case analysis.MoodDark:
		return palette{
			background: dimColor(base.background, 0.75),
			panel:      dimColor(base.panel, 0.8),
			text:       base.text,
			dim:        base.dim,
			accent:     blendColor(base.accent, rgb(120, 40, 180), 0.45),
			accent2:    blendColor(base.accent2, rgb(80, 20, 120), 0.4),
		}
	case analysis.MoodGroove:
		return palette{
			background: base.background,
			panel:      base.panel,
			text:       base.text,
			dim:        base.dim,
			accent:     blendColor(base.accent, rgb(60, 220, 120), 0.35),
			accent2:    blendColor(base.accent2, rgb(220, 60, 160), 0.3),
		}
	default:
		return base
	}
}

func blendColor(a, b uintptr, t float64) uintptr {
	ar, ag, ab := byte(a&0xff), byte((a>>8)&0xff), byte((a>>16)&0xff)
	br, bg, bb := byte(b&0xff), byte((b>>8)&0xff), byte((b>>16)&0xff)
	r := byte(float64(ar)*(1-t) + float64(br)*t)
	g := byte(float64(ag)*(1-t) + float64(bg)*t)
	bv := byte(float64(ab)*(1-t) + float64(bb)*t)
	return rgb(r, g, bv)
}

func (s *appState) bpmPulseMultiplier() float64 {
	if s.bpm <= 0 {
		return 1
	}
	beatPeriod := 60.0 / s.bpm
	phase := math.Mod(s.position().Seconds(), beatPeriod) / beatPeriod
	pulse := 0.5 + 0.5*math.Cos(phase*2*math.Pi)
	return 1 + pulse*0.18
}

func (s *appState) applyBPMPulse(bars []float64) []float64 {
	if s.bpm <= 0 || len(bars) == 0 {
		return bars
	}
	pulse := s.bpmPulseMultiplier()
	third := len(bars) / 3
	if third < 1 {
		third = 1
	}
	out := make([]float64, len(bars))
	for i, v := range bars {
		mult := 1.0
		if i < third {
			mult = pulse
		}
		out[i] = math.Min(1, v*mult)
	}
	return out
}


func (s *appState) cycleInputSource() {
	if s.desktop == nil {
		s.desktop = newDesktopInput()
	}
	switch s.inputSource {
	case inputPlayer:
		s.inputSource = inputDesktop
		s.desktop.setCaptureKind(captureLoopback)
		if err := s.desktop.start(); err != nil {
			s.inputSource = inputPlayer
			s.flashStatus("Desktop audio failed: "+err.Error(), 3*time.Second)
			return
		}
		s.desktopMode = true
		s.flashStatus("Input: Desktop audio (loopback)", 2*time.Second)
	case inputDesktop:
		s.stopDesktopInput()
		s.inputSource = inputMic
		s.desktop.setCaptureKind(captureMic)
		if err := s.desktop.start(); err != nil {
			s.inputSource = inputPlayer
			s.flashStatus("Mic input failed: "+err.Error(), 3*time.Second)
			return
		}
		s.desktopMode = true
		s.flashStatus("Input: Microphone", 2*time.Second)
	case inputMic:
		s.stopDesktopInput()
		s.inputSource = inputPlayer
		s.flashStatus("Input: Player audio", 2*time.Second)
	}
	s.saveSettings()
	invalidate()
}

func (s *appState) inputSourceName() string {
	switch s.inputSource {
	case inputDesktop:
		if s.desktop != nil && s.desktop.errText() != "" {
			return "Desktop error"
		}
		return "Desktop"
	case inputMic:
		if s.desktop != nil && s.desktop.errText() != "" {
			return "Mic error"
		}
		return "Mic"
	default:
		return "Player"
	}
}

func (s *appState) cycleOverlayAspect() {
	s.overlayAspect = (s.overlayAspect + 1) % 4
	if s.overlayHWND != 0 {
		s.applyOverlayAspect()
	}
	names := []string{"Free", "16:9", "4:3", "1:1"}
	s.flashStatus("Overlay aspect: "+names[s.overlayAspect], 2*time.Second)
	s.saveSettings()
	invalidate()
}

func (s *appState) applyOverlayAspect() {
	if s.overlayHWND == 0 {
		return
	}
	var bounds winRect
	procGetWindowRect.Call(s.overlayHWND, uintptr(unsafe.Pointer(&bounds)))
	w := bounds.right - bounds.left
	h := bounds.bottom - bounds.top
	switch s.overlayAspect {
	case overlayAspect16x9:
		h = w * 9 / 16
	case overlayAspect4x3:
		h = w * 3 / 4
	case overlayAspect1x1:
		if w > h {
			w = h
		} else {
			h = w
		}
	}
	procSetWindowPos.Call(s.overlayHWND, 0, uintptr(bounds.left), uintptr(bounds.top), uintptr(w), uintptr(h), swpShowWindow)
}

func (s *appState) toggleLiveFFT() {
	s.liveFFT = !s.liveFFT
	state := "off"
	if s.liveFFT {
		state = "on"
	}
	s.flashStatus("Live FFT "+state, 2*time.Second)
	s.saveSettings()
	invalidate()
}

func (s *appState) toggleMoodReactive() {
	s.moodReactive = !s.moodReactive
	state := "off"
	if s.moodReactive {
		state = "on"
	}
	s.flashStatus("Mood-reactive visuals "+state, 2*time.Second)
	s.saveSettings()
	invalidate()
}

func (s *appState) toggleChromaKey() {
	s.chromaKey = !s.chromaKey
	state := "off"
	if s.chromaKey {
		state = "on"
	}
	s.flashStatus("Chroma key "+state, 2*time.Second)
	s.saveSettings()
	invalidate()
}

func drawVisualizationBlend(hdc uintptr, bounds rect, bars []float64, modeA, modeB visualMode, blend float64) {
	if blend >= 1 || modeA == modeB {
		drawVisualization(hdc, bounds, bars, modeB)
		return
	}
	// Fade from previous mode to new mode using full bar heights.
	drawVisualizationSoft(hdc, bounds, bars, modeA, 1-blend)
	drawVisualizationSoft(hdc, bounds, bars, modeB, blend)
}

func drawWithEffectsBlend(hdc uintptr, bounds rect, bars []float64, peaks []float64, trails [][]float64, modeA, modeB visualMode, blend float64) {
	if blend >= 1 || modeA == modeB {
		drawWithEffects(hdc, bounds, bars, peaks, trails, modeB)
		return
	}
	drawWithEffects(hdc, bounds, bars, peaks, trails, modeA)
	if blend > 0.4 {
		drawVisualizationSoft(hdc, bounds, bars, modeB, blend)
	}
}
