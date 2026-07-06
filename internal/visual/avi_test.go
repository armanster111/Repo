package visual

import (
	"os"
	"testing"
)

func TestEncodeMJPEGAVI(t *testing.T) {
	frames := []*Canvas{
		makeTestFrame(64, 48, 0x00FF0000),
		makeTestFrame(64, 48, 0x0000FF00),
		makeTestFrame(64, 48, 0x000000FF),
	}
	path := t.TempDir() + "/test.avi"
	if err := EncodeMJPEGAVI(path, frames, 30); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "AVI " {
		t.Fatalf("unexpected header: %q", string(data[:12]))
	}
}

func makeTestFrame(w, h int, col uint32) *Canvas {
	c := NewCanvas(w, h)
	c.Clear(col)
	return c
}
