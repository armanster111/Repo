//go:build windows

package main

import "time"

type settingsRow struct {
	label  string
	action string
	bounds rect
}

func (s *appState) drawSettingsPanel(hdc uintptr, width, height int32, palette palette, chrome uiChromeLayout) {
	if !s.showSettings {
		s.settingsRows = nil
		s.modePreviewBounds = nil
		s.uiStylePreviewBounds = nil
		return
	}
	left := chrome.marginL
	top := chrome.vizTop - 4
	right := width - chrome.marginR - 280
	if right < left+280 {
		right = width - chrome.marginR
	}
	bottom := top + 360
	panel := rect{left: left, top: top, right: right, bottom: bottom}
	fillPanel(hdc, panel, palette, chrome)
	procSetTextColor.Call(hdc, palette.text)
	textOut(hdc, left+12, top+10, "Settings — click rows to toggle (3 to close)")
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
		{"Live FFT", "livefft", s.liveFFT},
		{"Mood reactive", "mood", s.moodReactive},
		{"Chroma key", "chroma", s.chromaKey},
	}
	for i, row := range rows {
		y := top + 34 + int32(i*22)
		r := rect{left: left + 8, top: y - 2, right: right - 8, bottom: y + 16}
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

	sliderTop := top + 34 + int32(len(rows)*22) + 8
	s.intensitySlider = rect{left: left + 12, top: sliderTop, right: right - 12, bottom: sliderTop + 14}
	fillPanel(hdc, s.intensitySlider, palette, chrome)
	level := s.visualIntensity
	if level < 0.1 {
		level = 0.1
	}
	if level > 2 {
		level = 2
	}
	knob := s.intensitySlider.left + int32((level-0.1)/1.9*float64(s.intensitySlider.right-s.intensitySlider.left))
	fill(hdc, rect{left: s.intensitySlider.left, top: s.intensitySlider.top, right: knob, bottom: s.intensitySlider.bottom}, palette.accent)
	procSetTextColor.Call(hdc, palette.dim)
	textOut(hdc, left+12, sliderTop+18, "Visual intensity (drag): "+formatFloat(s.visualIntensity))

	previewTop := sliderTop + 40
	procSetTextColor.Call(hdc, palette.dim)
	textOut(hdc, left+12, previewTop, "Visualizer modes (click):")
	s.modePreviewBounds = s.modePreviewBounds[:0]
	previewModes := []visualMode{modeClassic, modeAurora, modeNeonCity, modeFluid, modeGalaxy, modeChromatic, modeSupernova, modeLiquid}
	box := int32(28)
	gap := int32(6)
	for i, m := range previewModes {
		x := left + 12 + int32(i)*(box+gap)
		y := previewTop + 18
		r := rect{left: x, top: y, right: x + box, bottom: y + box}
		s.modePreviewBounds = append(s.modePreviewBounds, r)
		col := shaderPreviewColor(m)
		if m == s.mode {
			fill(hdc, r, col)
			drawPanelBorder(hdc, r, palette.accent2, 2)
		} else {
			fill(hdc, r, dimColor(col, 0.55))
		}
	}

	styleTop := previewTop + 56
	s.drawUIStylePicker(hdc, left+12, styleTop, palette, chrome)

	procSetTextColor.Call(hdc, palette.dim)
	textOut(hdc, left+12, bottom-42, "Mood: "+s.mood+"  BPM: "+formatBPM(s.bpm)+"  F5 Live FFT  F6 Mood  F7 Chroma")
	textOut(hdc, left+12, bottom-22, "F8 UI | F10 overlay aspect | I=cycle input | 1=OBS overlay")
}

func (s *appState) handleIntensitySlider(x, y int32) bool {
	if !s.showSettings || !pointInRect(x, y, s.intensitySlider) {
		return false
	}
	frac := float64(x-s.intensitySlider.left) / float64(s.intensitySlider.right-s.intensitySlider.left)
	s.visualIntensity = clamp(0.1+frac*1.9, 0.1, 2.0)
	s.saveSettings()
	invalidate()
	return true
}

func (s *appState) handleModePreviewClick(x, y int32) bool {
	if !s.showSettings {
		return false
	}
	previewModes := []visualMode{modeClassic, modeAurora, modeNeonCity, modeFluid, modeGalaxy, modeChromatic, modeSupernova, modeLiquid}
	for i, r := range s.modePreviewBounds {
		if !pointInRect(x, y, r) {
			continue
		}
		if i < len(previewModes) {
			s.setMode(previewModes[i])
			s.flashStatus("Visualizer: "+modeName(s.mode), 2*time.Second)
			s.saveSettings()
			invalidate()
		}
		return true
	}
	return false
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
		case "livefft":
			s.toggleLiveFFT()
		case "mood":
			s.toggleMoodReactive()
		case "chroma":
			s.toggleChromaKey()
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
