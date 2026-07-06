package lyrics

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAtReturnsActiveLine(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "song.mp3")
	lrc := filepath.Join(dir, "song.lrc")
	_ = os.WriteFile(audio, []byte("x"), 0644)
	_ = os.WriteFile(lrc, []byte("[00:01.00]Hello\n[00:03.00]World\n"), 0644)

	lines := Load(audio)
	if got := At(lines, 2*time.Second); got != "Hello" {
		t.Fatalf("At() = %q, want Hello", got)
	}
}
