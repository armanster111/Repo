package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Preset captures a full look-and-feel configuration.
type Preset struct {
	Name               string  `json:"name"`
	Mode               int     `json:"mode"`
	Theme              int     `json:"theme"`
	EqBass             int     `json:"eq_bass"`
	EqMid              int     `json:"eq_mid"`
	EqTreble           int     `json:"eq_treble"`
	DesktopSensitivity float64 `json:"desktop_sensitivity"`
	VisualIntensity    float64 `json:"visual_intensity"`
	Karaoke            bool    `json:"karaoke"`
	Ambient            bool    `json:"ambient"`
}

// Pack is an importable bundle of presets (plugin pack).
type Pack struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Presets     []Preset `json:"presets"`
}

// Load reads presets from a JSON file.
func Load(path string) ([]Preset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var presets []Preset
	if err := json.Unmarshal(data, &presets); err == nil {
		return presets, nil
	}
	var pack Pack
	if err := json.Unmarshal(data, &pack); err != nil {
		return nil, err
	}
	return pack.Presets, nil
}

// Save writes presets to disk.
func Save(path string, presets []Preset) error {
	data, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadPack reads a plugin/preset pack manifest.
func LoadPack(path string) (Pack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Pack{}, err
	}
	var pack Pack
	if err := json.Unmarshal(data, &pack); err != nil {
		return Pack{}, err
	}
	return pack, nil
}
