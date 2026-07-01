# Music Visualizer Pro

A native Windows music player and visualizer with playlists, MP3/WAV playback, themes, and track seeking.

## What it does

- Opens or drag-drops audio files and folders.
- Plays Windows-supported audio through MCI, including `.wav`, `.mp3`, `.wma`, `.mid`, `.aiff`, `.au`, and `.snd`.
- Builds a queue from the selected folder and supports previous/next, shuffle, repeat one/all, favorites, and recent files.
- Draws real waveform visualizers for PCM WAV files and animated visualizers for other formats.
- Includes ten visualizer modes: Classic, Mirror, Blocks, Wave, Halo, Spectrum, Fire, Particles, Tunnel, and Plasma.
- Includes four themes: Neon, Lava, Cyberpunk, and Ocean.
- Shows a clickable progress bar for shifting through the track.
- Includes volume up/down, mute, fullscreen, mini-player, media key support, sleep timer, and snapshot export.
- Saves settings in `%APPDATA%\MusicVisualizerPro\settings.json`.
- Supports keyboard shortcuts:
  - `O`: open an audio file
  - Drag/drop: open files or folders
  - `Space`: pause/resume playback
  - `B` / `N`: previous/next track
  - `V`: cycle visualizer modes
  - `G`: cycle themes
  - `P`: toggle shuffle
  - `Q`: cycle repeat mode
  - `F`: favorite/unfavorite current track
  - `D`: load the current folder as a playlist
  - `M`: mute/unmute
  - `Up` / `Down`: volume up/down
  - `Left` / `Right`: seek backward/forward
  - `T`: toggle 10-second and 30-second seek jumps
  - `R`: restart the current track
  - `S`: cycle sleep timer
  - `C`: export a visualizer snapshot to the Desktop
  - `X`: toggle mini-player mode
  - `F11` / `L`: toggle fullscreen/maximized mode
  - `Esc`: quit

## Download/build

Safest option: download the source-only package, then build it on your own Windows computer:

```text
dist/music-visualizer-source-package.zip
```

After extracting it on Windows, double-click:

```text
build-windows.bat
```

If Go is not installed, the script will offer to install it with Windows Package Manager (`winget`). After the build finishes, it creates a `Music Visualizer Pro` shortcut on your Desktop.

I removed the prebuilt `.exe` download from the current branch because unsigned executables downloaded from GitHub can be blocked by Windows Defender or Edge. Build locally from source instead.

To rebuild it from source:

```sh
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o dist/music-visualizer.exe ./cmd/music-visualizer
```

## Notes

The executable stays dependency-free. WAV files get real waveform analysis; MP3 and other Windows-supported formats play normally and use generated animated visualizers because decoding compressed audio would require a bundled decoder.
