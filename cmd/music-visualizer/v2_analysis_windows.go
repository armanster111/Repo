//go:build windows

package main

import (
	"time"

	"github.com/armanster111/music-visualizer/internal/analysis"
	"github.com/armanster111/music-visualizer/internal/lyrics"
)

func (s *appState) fetchLyricsAsync() {
	if s.lyricsFetching || s.filePath == "" {
		return
	}
	if len(s.lyricLines) > 0 {
		return
	}
	artist := s.meta.Artist
	title := s.meta.Title
	if title == "" {
		title = s.filePath
	}
	s.lyricsFetching = true
		go func() {
		lines, err := lyrics.FetchOnline(artist, title)
		s.lyricsFetching = false
		if err != nil || len(lines) == 0 {
			if err != nil {
				s.flashStatus("Lyrics: "+err.Error(), 4*time.Second)
			}
			return
		}
		s.lyricLines = lines
		s.flashStatus("Synced lyrics loaded.", 2*time.Second)
		invalidate()
	}()
}

func detectMoodFromBars(bars []float64) analysis.Mood {
	return analysis.DetectMood(bars)
}

func (s *appState) updateAnalysis(bars []float64) {
	s.mood = string(analysis.DetectMood(bars))
	s.bpm = analysis.EstimateBPM(s.energyHistory, time.Second/time.Duration(fps))
	s.maybeAutoPreset(bars)
	s.maybeMoodMode(bars)
}

func (s *appState) maybeMoodMode(bars []float64) {
	if !s.moodReactive || s.ultra != nil && s.ultra.showcase {
		return
	}
	mood := analysis.DetectMood(bars)
	var target visualMode
	switch mood {
	case analysis.MoodEnergetic:
		target = modeFire
	case analysis.MoodChill:
		target = modeAurora
	case analysis.MoodBright:
		target = modeSupernova
	case analysis.MoodDark:
		target = modeTunnel
	case analysis.MoodGroove:
		target = modeLiquid
	default:
		return
	}
	if target == s.mode {
		return
	}
	// only switch occasionally to avoid flicker
	if len(s.energyHistory) > 0 && int(s.energyHistory[len(s.energyHistory)-1]*100)%17 != 0 {
		return
	}
	s.prevMode = s.mode
	s.modeBlend = 0
	s.mode = target
}

func (s *appState) applyVisualEQ(bars []float64) []float64 {
	return analysis.ApplyVisualEQ(bars, s.eqBass, s.eqMid, s.eqTreble)
}

func (s *appState) karaokeLine() string {
	if !s.karaokeMode {
		return ""
	}
	line := s.currentLyric()
	if line == "" {
		return ""
	}
	return "♪ " + line + " ♪"
}
