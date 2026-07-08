package audio

import (
	"strings"
	"testing"
)

func TestDecodeAIFFRejectsBadInput(t *testing.T) {
	_, _, err := DecodeAIFF(nil)
	if err == nil {
		t.Fatal("expected error for nil reader")
	}
	_, _, err = DecodeAIFF(strings.NewReader("not aiff"))
	if err == nil {
		t.Fatal("expected error for invalid data")
	}
}
