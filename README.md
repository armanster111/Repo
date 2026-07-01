# Music Visualizer

A tiny native Windows music visualizer for WAV files.

## What it does

- Opens `.wav` audio files from a native file picker.
- Plays the selected file through Windows audio.
- Draws animated visualizer bars from the decoded waveform.
- Supports keyboard shortcuts:
  - `O`: open a WAV file
  - `Space`: restart playback
  - `Esc`: quit

## Download/build

This repo includes a built executable at:

```text
dist/music-visualizer.exe
```

To rebuild it from source:

```sh
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o dist/music-visualizer.exe ./cmd/music-visualizer
```

## Notes

The visualizer intentionally supports uncompressed PCM WAV files so it can stay dependency-free and ship as one small `.exe`.
