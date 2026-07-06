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
