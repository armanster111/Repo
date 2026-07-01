package presets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presets.json")
	in := []Preset{{Name: "Test", Mode: 1, Theme: 2, VisualIntensity: 1.1}}
	if err := Save(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Name != "Test" {
		t.Fatalf("unexpected %+v", out)
	}
}

func TestLoadPack(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pack.json")
	pack := Pack{ID: "demo", Name: "Demo", Presets: []Preset{{Name: "A"}}}
	data := []byte(`{"id":"demo","name":"Demo","presets":[{"name":"A"}]}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadPack(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != pack.ID || len(loaded.Presets) != 1 {
		t.Fatalf("unexpected %+v", loaded)
	}
}
