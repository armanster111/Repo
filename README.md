# Music Visualizer Pro

A native Windows music player and visualizer with playlists, MP3/WAV playback, desktop audio capture, themes, lyrics, album-art backgrounds, and track seeking.

## What it does

- Opens or drag-drops audio files and folders.
- Plays Windows-supported audio through MCI, including `.wav`, `.mp3`, `.wma`, `.mid`, `.aiff`, `.au`, and `.snd`.
- Decodes **MP3 and WAV** for real waveform visualizers.
- **Desktop Input mode** listens to default Windows speaker output (WASAPI loopback).
- Builds a queue from folders with **search**, **recent**, and **favorites** views.
- Shows **album art** as a blurred background and **`.lrc` lyrics** when available.
- Includes ten visualizer modes with **beat pulse** and **bass/mid/treble** coloring.
- Includes built-in themes plus up to **5 saved custom themes** (`H`).
- **System tray** icon for quick access.
- **Equalizer** bass/treble via Windows MCI (`K` cycles bands).
- **Playback speed** control (`J`).
- **Visual-only fullscreen** mode (`Z`).
- Remembers window size/position and settings in `%APPDATA%\MusicVisualizerPro\settings.json`.

## Keyboard shortcuts

- `O` open audio | drag/drop files or folders
- `Space` pause/play | `B`/`N` previous/next
- `V` visualizer | `G` theme | `H` save custom theme
- `I` desktop input | `A` auto desktop when idle | `]` sensitivity | `W` output device
- `U` recent | `Y` favorites | `/` search (type to filter)
- `P` shuffle | `Q` repeat | `F` favorite | `D` folder playlist
- `M` mute | `Up`/`Down` volume | `Left`/`Right` seek | `T` seek jump size
- `J` playback speed | `K` EQ band | `R` restart | `S` sleep timer
- `C` snapshot | `X` mini | `Z` visual-only | `F11`/`L` fullscreen
- `Esc` quit

## Download/build

Safest option: download the source-only package, then build on your Windows PC:

```text
dist/music-visualizer-source-package.zip
```

1. Extract the ZIP on Windows.
2. Double-click `build-windows.bat` (installs Go via winget if needed).
3. Open `dist\music-visualizer.exe` or the Desktop shortcut.
4. Optional: run `install-windows.bat` to install into `%LOCALAPPDATA%\MusicVisualizerPro`.

Unsigned `.exe` downloads may be blocked by Defender. Building locally avoids that. See `docs/CODE_SIGNING.md` for signing with your own certificate.

## Notes

- MP3/WAV get real waveform analysis; other formats still play through Windows and use animated visualizers.
- Desktop input visualizes whatever your PC is playing (YouTube, Spotify, games, etc.).
- Place a matching `.lrc` file next to a song for synced lyrics.
