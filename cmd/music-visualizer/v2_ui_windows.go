//go:build windows

package main

import (
	"time"
)

func (s *appState) toggleCrossfade() {
	s.crossfade = !s.crossfade
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) tickCrossfade() {
	if !s.crossfade || !s.playing {
		s.crossfadeLevel = 1.0
		return
	}
	remaining := s.duration - s.position()
	if remaining < 2*time.Second && remaining > 0 {
		s.crossfadeLevel = clamp(remaining.Seconds()/2, 0.2, 1)
		return
	}
	if s.position() < 1500*time.Millisecond {
		s.crossfadeLevel = clamp(s.position().Seconds()/1.5, 0.2, 1)
		return
	}
	s.crossfadeLevel = 1.0
}

func (s *appState) applyCrossfadeVolume() {
	if !s.crossfade || s.muted {
		return
	}
	vol := int(float64(s.volume) * s.crossfadeLevel)
	_ = mciSend("setaudio musicvisualizer_track volume to " + itoa(vol))
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var digits []byte
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func (s *appState) toggleKaraoke() {
	s.karaokeMode = !s.karaokeMode
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) toggleAutoPreset() {
	s.autoPreset = !s.autoPreset
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) toggleSettingsPanel() {
	s.showSettings = !s.showSettings
	invalidate()
}

func (s *appState) drawSettingsPanel(hdc uintptr, width, height int32, palette palette) {
	if !s.showSettings {
		return
	}
	left := int32(28)
	top := int32(108)
	right := width - 320
	bottom := top + 220
	fill(hdc, rect{left: left, top: top, right: right, bottom: bottom}, dimColor(palette.panel, 0.92))
	procSetTextColor.Call(hdc, palette.text)
	lines := []string{
		"Settings (press 3 to close)",
		"Crossfade: " + boolText(s.crossfade) + "  |  Karaoke: " + boolText(s.karaokeMode),
		"Ambient: " + boolText(s.ambientMode) + "  |  Overlay: " + boolText(s.overlayMode),
		"DJ: " + boolText(s.djMode) + "  |  Auto preset: " + boolText(s.autoPreset),
		"Remote: http://localhost:8765  |  Enabled: " + boolText(s.remoteEnabled),
		"Mood: " + s.mood + "  |  BPM: " + formatBPM(s.bpm),
		"Visual intensity: " + formatFloat(s.visualIntensity),
		"4=export sync  5=import presets  6=party  7=library scan",
	}
	for i, line := range lines {
		procSetTextColor.Call(hdc, palette.dim)
		if i == 0 {
			procSetTextColor.Call(hdc, palette.text)
		}
		textOut(hdc, left+12, top+12+int32(i*22), line)
	}
}

func boolText(v bool) string {
	if v {
		return "On"
	}
	return "Off"
}

func formatBPM(v float64) string {
	if v <= 0 {
		return "-"
	}
	return itoa(int(v))
}

func formatFloat(v float64) string {
	return itoa(int(v * 100)) + "%"
}

func volumeSliderRect(width, height int32) rect {
	return rect{left: width - 260, top: 74, right: width - 40, bottom: 88}
}

func (s *appState) drawVolumeSlider(hdc uintptr, width, height int32, palette palette) {
	if app.mini || app.visualOnly {
		return
	}
	bar := volumeSliderRect(width, height)
	s.volumeSliderBounds = bar
	fill(hdc, bar, palette.panel)
	level := float64(s.volume) / 1000
	fill(hdc, rect{left: bar.left, top: bar.top, right: bar.left + int32(level*float64(bar.right-bar.left)), bottom: bar.bottom}, palette.accent2)
	procSetTextColor.Call(hdc, palette.dim)
	textOut(hdc, bar.left, bar.bottom+4, "Volume")
}

func (s *appState) handleVolumeSlider(x, y int32) bool {
	if !pointInRect(x, y, s.volumeSliderBounds) {
		return false
	}
	bar := s.volumeSliderBounds
	frac := float64(x-bar.left) / float64(bar.right-bar.left)
	s.volume = clampInt(int(frac*1000), 0, 1000)
	s.muted = false
	s.applyVolume()
	s.updateStatus()
	s.saveSettings()
	invalidate()
	return true
}

func (s *appState) handlePlaylistClick(x, y int32) bool {
	for i, r := range s.playlistRowBounds {
		if pointInRect(x, y, r) {
			list := s.filteredPlaylist()
			if i < len(list) {
				s.openPath(list[i].Path)
				return true
			}
			entries := s.libraryEntriesForPanel()
			groups := s.libraryGroupsForPanel()
			if len(groups) > 0 && i < len(groups) {
				s.panelGroupKey = groups[i]
				s.panelRow = 0
				invalidate()
				return true
			}
			if i < len(entries) {
				s.openLibraryEntry(entries[i])
				return true
			}
		}
	}
	return false
}
