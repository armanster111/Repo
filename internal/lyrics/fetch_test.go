package lyrics

import (
	"testing"
	"time"
)

func TestParseLRC(t *testing.T) {
	text := "[00:12.50]Hello world\n[00:15.00]Second line"
	lines := ParseLRC(text)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0].Text != "Hello world" {
		t.Fatalf("unexpected first line: %+v", lines[0])
	}
}

func TestAt(t *testing.T) {
	lines := []Line{{At: 0, Text: "a"}, {At: time.Second, Text: "b"}}
	if At(lines, 500*time.Millisecond) != "a" {
		t.Fatal("expected a")
	}
	if At(lines, 1500*time.Millisecond) != "b" {
		t.Fatal("expected b")
	}
}
