# Music Visualizer Pro

A native Windows music player and visualizer with playlists, MP3/WAV playback, desktop audio capture, themes, lyrics, album-art backgrounds, and track seeking.

## Quick start (Windows)

1. Download **`MusicVisualizerPro-Windows.zip`**
2. Extract it anywhere (Downloads, Desktop, etc.)
3. Double-click **`MusicVisualizerPro.exe`**

No Go install. No build step. No separate installer required.

Direct download (latest branch build):

- ZIP: `https://github.com/armanster111/Repo/raw/cursor/music-visualizer-exe-7b99/dist/MusicVisualizerPro-Windows.zip`
- EXE: `https://github.com/armanster111/Repo/raw/cursor/music-visualizer-exe-7b99/dist/MusicVisualizerPro.exe`

Optional: run `install-windows.bat` to copy the app into `%LOCALAPPDATA%\MusicVisualizerPro` and create a Desktop shortcut.

If Windows SmartScreen warns about an unsigned app, choose **More info** → **Run anyway**. See `docs/CODE_SIGNING.md` if you want to sign releases yourself.

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

## Build from source (optional)

Only needed if you want to modify the code:

1. Install [Go](https://go.dev/dl/)
2. Run `build-windows.bat`
3. Open `dist\MusicVisualizerPro.exe`

CI also rebuilds the standalone `.exe` on every push via `.github/workflows/windows-release.yml`.

## Notes

- MP3/WAV get real waveform analysis; other formats still play through Windows and use animated visualizers.
- Desktop input visualizes whatever your PC is playing (YouTube, Spotify, games, etc.).
- Place a matching `.lrc` file next to a song for synced lyrics.
