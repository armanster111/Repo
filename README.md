# Music Visualizer Pro

A native Windows music visualizer for WAV files with multiple display modes and track seeking.

## What it does

- Opens `.wav` audio files from a native file picker.
- Plays the selected file through Windows audio.
- Draws animated visualizers from the decoded waveform.
- Includes five visualizer modes: Classic, Mirror, Blocks, Wave, and Halo.
- Shows a clickable progress bar for shifting through the track.
- Supports keyboard shortcuts:
  - `O`: open a WAV file
  - `Space`: pause/resume playback
  - `V`: cycle visualizer modes
  - `Left` / `Right`: seek backward/forward
  - `T`: toggle 10-second and 30-second seek jumps
  - `R`: restart the current track
  - `Esc`: quit

## Download/build

This repo includes a built executable at:

```text
dist/music-visualizer.exe
```

For a one-file download package, use:

```text
dist/music-visualizer-windows-package.zip
```

To rebuild it from source:

```sh
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o dist/music-visualizer.exe ./cmd/music-visualizer
```

## Notes

The visualizer intentionally supports uncompressed PCM WAV files so it can stay dependency-free and ship as one small `.exe`.
