//go:build windows

package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func (s *appState) toggleDJMode() {
	s.djMode = !s.djMode
	if s.djMode {
		s.status = "DJ mode on — , . crossfader | ; load deck B"
	} else {
		s.djDeckBPath = ""
		s.djDeckBFrames = nil
		s.status = "DJ mode off."
	}
	s.saveSettings()
	invalidate()
}

func (s *appState) djNudgeCrossfader(delta float64) {
	if !s.djMode {
		return
	}
	s.djCrossfader = clamp(s.djCrossfader+delta, 0, 1)
	s.updateStatus()
	invalidate()
}

func (s *appState) loadDJDeckB() {
	if s.filePath == "" {
		s.status = "Open a track first, then press ; to copy it to deck B."
		invalidate()
		return
	}
	s.djDeckBPath = s.filePath
	if frames, _ := s.buildFramesForPath(s.filePath); len(frames) > 0 {
		s.djDeckBFrames = frames
	}
	s.status = "Deck B loaded: " + filepath.Base(s.djDeckBPath)
	invalidate()
}

func (s *appState) mixDJBars(deckA []float64) []float64 {
	if len(s.djDeckBFrames) == 0 {
		return deckA
	}
	pos := s.position()
	idx := int(pos.Seconds() * float64(fps))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.djDeckBFrames) {
		idx = len(s.djDeckBFrames) - 1
	}
	deckB := append([]float64(nil), s.djDeckBFrames[idx].Bars...)
	if len(deckB) != len(deckA) {
		return deckA
	}
	mix := s.djCrossfader
	out := make([]float64, len(deckA))
	for i := range deckA {
		out[i] = clamp(deckA[i]*(1-mix)+deckB[i]*mix, 0, 1)
	}
	return out
}

func (s *appState) togglePartyMode() {
	if s.partyMode {
		s.partyMode = false
		s.karaokeMode = false
		s.visualIntensity = 1.0
		s.status = "Party mode off."
		s.saveSettings()
		invalidate()
		return
	}
	s.partyMode = true
	s.visualIntensity = 1.45
	s.karaokeMode = true
	s.mode = modeKaleidoscope
	s.flashStatus("Party mode on — kaleidoscope + karaoke. Press 6 to turn off.", 3*time.Second)
	s.saveSettings()
	invalidate()
}

func djStatusText(enabled bool, cross float64, deckB string) string {
	if !enabled {
		return ""
	}
	name := "empty"
	if deckB != "" {
		name = strings.TrimSuffix(filepath.Base(deckB), filepath.Ext(deckB))
	}
	return fmt.Sprintf("DJ %.0f%% B:%s", cross*100, truncate(name, 12))
}
