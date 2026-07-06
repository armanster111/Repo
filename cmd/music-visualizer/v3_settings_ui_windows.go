//go:build windows

package main

import "time"

type settingsRow struct {
	label  string
	action string
	bounds rect
}

func (s *appState) drawSettingsPanel(hdc uintptr, width, height int32, palette palette) {
	if !s.showSettings {
		s.settingsRows = nil
		return
	}
	left := int32(28)
	top := int32(108)
	right := width - 320
	bottom := top + 250
	panel := rect{left: left, top: top, right: right, bottom: bottom}
	fill(hdc, panel, dimColor(palette.panel, 0.92))
	procSetTextColor.Call(hdc, palette.text)
	textOut(hdc, left+12, top+10, "Settings — click a row to toggle (3 to close)")
	s.settingsRows = s.settingsRows[:0]
	rows := []struct {
		label  string
		action string
		on     bool
	}{
		{"Crossfade", "crossfade", s.crossfade},
		{"Karaoke lyrics", "karaoke", s.karaokeMode},
		{"Ambient mode", "ambient", s.ambientMode},
		{"OBS overlay", "overlay", s.overlayMode},
		{"DJ mode", "dj", s.djMode},
		{"Auto preset", "autopreset", s.autoPreset},
		{"Remote control", "remote", s.remoteEnabled},
		{"Party mode", "party", s.partyMode},
	}
	for i, row := range rows {
		y := top + 34 + int32(i*24)
		r := rect{left: left + 8, top: y - 2, right: right - 8, bottom: y + 18}
		s.settingsRows = append(s.settingsRows, settingsRow{label: row.label, action: row.action, bounds: r})
		state := "Off"
		color := palette.dim
		if row.on {
			state = "On"
			color = palette.accent2
		}
		procSetTextColor.Call(hdc, color)
		textOut(hdc, left+16, y, row.label+": "+state)
	}
	procSetTextColor.Call(hdc, palette.dim)
	textOut(hdc, left+12, bottom-42, "Mood: "+s.mood+"  BPM: "+formatBPM(s.bpm)+"  Intensity: "+formatFloat(s.visualIntensity))
	textOut(hdc, left+12, bottom-22, "Overlay opacity: [ / ]   Double-click visualizer for cinema")
}

func (s *appState) handleSettingsClick(x, y int32) bool {
	if !s.showSettings {
		return false
	}
	for _, row := range s.settingsRows {
		if !pointInRect(x, y, row.bounds) {
			continue
		}
		switch row.action {
		case "crossfade":
			s.toggleCrossfade()
		case "karaoke":
			s.toggleKaraoke()
		case "ambient":
			s.toggleAmbient()
		case "overlay":
			s.toggleOverlay()
		case "dj":
			s.toggleDJMode()
		case "autopreset":
			s.toggleAutoPreset()
		case "remote":
			s.toggleRemoteControl()
		case "party":
			s.togglePartyMode()
		}
		return true
	}
	return false
}

func (s *appState) toggleRemoteControl() {
	if s.remoteEnabled {
		s.stopRemoteControl()
		s.remoteEnabled = false
		s.flashStatus("Remote control off.", 2*time.Second)
	} else {
		s.startRemoteControl()
		s.remoteEnabled = true
		s.flashStatus("Remote: http://localhost:8765", 3*time.Second)
	}
	s.saveSettings()
	invalidate()
}
