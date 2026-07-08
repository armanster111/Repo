//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/armanster111/music-visualizer/internal/presets"
)

func presetsPath() string {
	return filepath.Join(filepath.Dir(settingsPath()), "presets.json")
}

func packsDir() string {
	return filepath.Join(filepath.Dir(settingsPath()), "packs")
}

func (s *appState) loadPresetsFromDisk() {
	list, err := presets.Load(presetsPath())
	if err != nil {
		s.presets = defaultPresets()
		return
	}
	if len(list) == 0 {
		s.presets = defaultPresets()
		return
	}
	s.presets = list
}

func defaultPresets() []presets.Preset {
	return []presets.Preset{
		{Name: "Neon Classic", Mode: int(modeClassic), Theme: int(themeNeon), VisualIntensity: 1},
		{Name: "Cyber Halo", Mode: int(modeHalo), Theme: int(themeCyberpunk), VisualIntensity: 1.2},
		{Name: "Ocean Wave", Mode: int(modeWave), Theme: int(themeOcean), VisualIntensity: 0.9},
		{Name: "Lava Fire", Mode: int(modeFire), Theme: int(themeLava), VisualIntensity: 1.1},
		{Name: "Party Plasma", Mode: int(modePlasma), Theme: int(themeCyberpunk), VisualIntensity: 1.4, Karaoke: true},
		{Name: "Bass Drop", Mode: int(modeBassDrop), Theme: int(themeCyberpunk), VisualIntensity: 1.5},
		{Name: "Vortex Dream", Mode: int(modeVortex), Theme: int(themeOcean), VisualIntensity: 1.2},
		{Name: "Comet Storm", Mode: int(modeComets), Theme: int(themeNeon), VisualIntensity: 1.3},
		{Name: "Prism Break", Mode: int(modePrism), Theme: int(themeLava), VisualIntensity: 1.1},
		{Name: "Rain Chamber", Mode: int(modeRain), Theme: int(themeOcean), VisualIntensity: 0.95, Ambient: true},
		{Name: "Supernova Cinema", Mode: int(modeSupernova), Theme: int(themeNeon), VisualIntensity: 1.4},
		{Name: "GPU Galaxy", Mode: int(modeGalaxy), Theme: int(themeCyberpunk), VisualIntensity: 1.2},
	}
}

func (s *appState) presetsList() []presets.Preset {
	if len(s.presets) == 0 {
		return defaultPresets()
	}
	return s.presets
}

func (s *appState) saveCurrentPreset() {
	p := presets.Preset{
		Name:               fmt.Sprintf("Preset %d", len(s.presets)+1),
		Mode:               int(s.mode),
		Theme:              int(s.theme),
		EqBass:             s.eqBass,
		EqMid:              s.eqMid,
		EqTreble:           s.eqTreble,
		DesktopSensitivity: s.desktopSensitivity,
		VisualIntensity:    s.visualIntensity,
		Karaoke:            s.karaokeMode,
		Ambient:            s.ambientMode,
	}
	s.presets = append(s.presets, p)
	_ = presets.Save(presetsPath(), s.presets)
	s.presetIndex = len(s.presets) - 1
	s.saveSettings()
	s.status = "Saved preset: " + p.Name
	invalidate()
}

func (s *appState) cyclePreset() {
	list := s.presetsList()
	if len(list) == 0 {
		return
	}
	s.presetIndex = (s.presetIndex + 1) % len(list)
	s.applyPreset(list[s.presetIndex])
}

func (s *appState) applyPreset(p presets.Preset) {
	s.mode = visualMode(p.Mode % int(modeCount))
	s.theme = theme(p.Theme % int(themeCount))
	s.eqBass = p.EqBass
	s.eqMid = p.EqMid
	s.eqTreble = p.EqTreble
	if p.DesktopSensitivity > 0 {
		s.desktopSensitivity = p.DesktopSensitivity
		if s.desktop != nil {
			s.desktop.setSensitivity(s.desktopSensitivity)
		}
	}
	if p.VisualIntensity > 0 {
		s.visualIntensity = p.VisualIntensity
	}
	s.karaokeMode = p.Karaoke
	s.ambientMode = p.Ambient
	s.applyEqualizer()
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) exportPresets() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Desktop", "music-visualizer-presets.json")
	if err := presets.Save(path, s.presetsList()); err != nil {
		s.status = "Preset export failed: " + err.Error()
	} else {
		s.status = "Exported presets: " + path
	}
	invalidate()
}

func (s *appState) importPresets() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Desktop", "music-visualizer-presets.json")
	list, err := presets.Load(path)
	if err != nil {
		s.status = "Preset import failed: " + err.Error()
		invalidate()
		return
	}
	s.presets = append(s.presets, list...)
	_ = presets.Save(presetsPath(), s.presets)
	s.status = fmt.Sprintf("Imported %d presets.", len(list))
	s.saveSettings()
	invalidate()
}

func (s *appState) loadPluginPacks() {
	_ = os.MkdirAll(packsDir(), 0755)
	entries, err := os.ReadDir(packsDir())
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		pack, err := presets.LoadPack(filepath.Join(packsDir(), e.Name()))
		if err != nil || len(pack.Presets) == 0 {
			continue
		}
		s.presets = append(s.presets, pack.Presets...)
	}
	if len(s.presets) > 0 {
		_ = presets.Save(presetsPath(), s.presets)
	}
}

func (s *appState) maybeAutoPreset(bars []float64) {
	if !s.autoPreset || len(s.presets) == 0 {
		return
	}
	mood := string(detectMoodFromBars(bars))
	for i, p := range s.presets {
		if mood == "Energetic" && (p.Mode == int(modeFire) || p.Mode == int(modePlasma)) {
			if s.presetIndex != i {
				s.presetIndex = i
				s.applyPreset(p)
			}
			return
		}
		if mood == "Chill" && p.Ambient {
			if s.presetIndex != i {
				s.presetIndex = i
				s.applyPreset(p)
			}
			return
		}
	}
}

func (s *appState) presetsToSettings() []presetData {
	out := make([]presetData, len(s.presets))
	for i, p := range s.presets {
		out[i] = presetData{
			Name: p.Name, Mode: p.Mode, Theme: p.Theme,
			EqBass: p.EqBass, EqMid: p.EqMid, EqTreble: p.EqTreble,
			DesktopSensitivity: p.DesktopSensitivity, VisualIntensity: p.VisualIntensity,
			Karaoke: p.Karaoke, Ambient: p.Ambient,
		}
	}
	return out
}

func (s *appState) presetsFromSettings(data []presetData) {
	s.presets = nil
	for _, p := range data {
		s.presets = append(s.presets, presets.Preset{
			Name: p.Name, Mode: p.Mode, Theme: p.Theme,
			EqBass: p.EqBass, EqMid: p.EqMid, EqTreble: p.EqTreble,
			DesktopSensitivity: p.DesktopSensitivity, VisualIntensity: p.VisualIntensity,
			Karaoke: p.Karaoke, Ambient: p.Ambient,
		})
	}
}

func (s *appState) installDefaultPack() {
	packPath := filepath.Join(packsDir(), "builtin-party.json")
	_ = os.MkdirAll(packsDir(), 0755)
	if _, err := os.Stat(packPath); err == nil {
		return
	}
	pack := presets.Pack{
		ID: "builtin-party", Name: "Party Pack", Version: "1",
		Description: "High-energy visual presets",
		Presets: []presets.Preset{
			{Name: "Streamer Halo", Mode: int(modeHalo), Theme: int(themeCyberpunk), VisualIntensity: 1.3},
			{Name: "Aurora Dream", Mode: int(modeAurora), Theme: int(themeOcean), VisualIntensity: 1, Ambient: true},
			{Name: "Kaleidoscope Rave", Mode: int(modeKaleidoscope), Theme: int(themeNeon), VisualIntensity: 1.5},
		},
	}
	data, _ := json.MarshalIndent(pack, "", "  ")
	_ = os.WriteFile(packPath, data, 0644)
}
