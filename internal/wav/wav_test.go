package wav

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestDecodePCM16StereoDownmixesToMono(t *testing.T) {
	data := makeTestWAV(t, 2, 44100, 16, []int16{
		32767, 32767,
		32767, -32768,
		-32768, -32768,
	})

	audio, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if audio.Channels != 2 {
		t.Fatalf("Channels = %d, want 2", audio.Channels)
	}
	if audio.SampleRate != 44100 {
		t.Fatalf("SampleRate = %d, want 44100", audio.SampleRate)
	}
	if audio.BitsPerSample != 16 {
		t.Fatalf("BitsPerSample = %d, want 16", audio.BitsPerSample)
	}
	if len(audio.Mono) != 3 {
		t.Fatalf("len(Mono) = %d, want 3", len(audio.Mono))
	}

	if audio.Mono[0] < 0.99 {
		t.Fatalf("Mono[0] = %f, want near 1", audio.Mono[0])
	}
	if audio.Mono[1] > 0.01 || audio.Mono[1] < -0.01 {
		t.Fatalf("Mono[1] = %f, want near 0", audio.Mono[1])
	}
	if audio.Mono[2] > -0.99 {
		t.Fatalf("Mono[2] = %f, want near -1", audio.Mono[2])
	}
}

func TestDecodeRejectsNonWaveData(t *testing.T) {
	_, err := Decode(bytes.NewReader([]byte("not wave data")))
	if err == nil {
		t.Fatal("Decode() error = nil, want error")
	}
}

func makeTestWAV(t *testing.T, channels, sampleRate, bitsPerSample uint16, samples []int16) []byte {
	t.Helper()

	var pcm bytes.Buffer
	for _, sample := range samples {
		if err := binary.Write(&pcm, binary.LittleEndian, sample); err != nil {
			t.Fatal(err)
		}
	}

	byteRate := uint32(sampleRate) * uint32(channels) * uint32(bitsPerSample/8)
	blockAlign := channels * (bitsPerSample / 8)

	var out bytes.Buffer
	out.WriteString("RIFF")
	_ = binary.Write(&out, binary.LittleEndian, uint32(36+pcm.Len()))
	out.WriteString("WAVE")
	out.WriteString("fmt ")
	_ = binary.Write(&out, binary.LittleEndian, uint32(16))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, channels)
	_ = binary.Write(&out, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&out, binary.LittleEndian, byteRate)
	_ = binary.Write(&out, binary.LittleEndian, blockAlign)
	_ = binary.Write(&out, binary.LittleEndian, bitsPerSample)
	out.WriteString("data")
	_ = binary.Write(&out, binary.LittleEndian, uint32(pcm.Len()))
	out.Write(pcm.Bytes())
	return out.Bytes()
}
