package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanEmptyDir(t *testing.T) {
	dir := t.TempDir()
	idx, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(idx.Entries))
	}
}

func TestMostPlayed(t *testing.T) {
	idx := &Index{
		Entries: []Entry{
			{Path: "a.mp3", Title: "A", PlayCount: 2},
			{Path: "b.mp3", Title: "B", PlayCount: 5},
		},
	}
	top := idx.MostPlayed(1)
	if len(top) != 1 || top[0].Title != "B" {
		t.Fatalf("unexpected top: %+v", top)
	}
}

func TestIsAudio(t *testing.T) {
	if !isAudio("song.flac") {
		t.Fatal("flac should be audio")
	}
	if isAudio("notes.txt") {
		t.Fatal("txt should not be audio")
	}
}

func TestScanSkipsRecycleBin(t *testing.T) {
	root := t.TempDir()
	_ = os.Mkdir(filepath.Join(root, "$Recycle.Bin"), 0755)
	_ = os.WriteFile(filepath.Join(root, "track.mp3"), []byte("x"), 0644)
	idx, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(idx.Entries))
	}
}
