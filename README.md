# Music Visualizer Pro

A native Windows music player and visualizer with playlists, MP3/WAV playback, desktop audio capture, themes, lyrics, album-art backgrounds, library scanning, presets, OBS overlay, DJ mode, and remote control.

## Quick start (Windows)

1. Download **`MusicVisualizerPro-Windows.zip`**
2. Extract it anywhere
3. Double-click **`MusicVisualizerPro.exe`**

Direct download:

- ZIP: `https://github.com/armanster111/Repo/raw/cursor/music-visualizer-exe-7b99/dist/MusicVisualizerPro-Windows.zip`
- EXE: `https://github.com/armanster111/Repo/raw/cursor/music-visualizer-exe-7b99/dist/MusicVisualizerPro.exe`

If Windows SmartScreen warns about an unsigned app, choose **More info → Run anyway**.

## v2 features

### Music library
- Auto-scans `Music`, `Downloads`, and your last playlist folder on startup
- **Library / Artists / Albums / Top Played** panels (`\` cycles views, **Lib** button)
- Click tracks in the right panel to play
- Play counts saved and used for smart "Top Played" list

### Visualizers (14 modes)
Classic, Mirror, Blocks, Wave, Halo, Spectrum, Fire, Particles, Tunnel, Plasma, **Aurora**, **Mandala**, **Starfield**, **Kaleidoscope**

### Presets & plugin packs
- **8** cycle preset | **9** save preset | **5** import from Desktop
- Presets saved to `%APPDATA%\MusicVisualizerPro\presets.json`
- Drop JSON packs in `%APPDATA%\MusicVisualizerPro\packs\`
- **0** auto-preset (mood-based switching)

### Lyrics
- Local `.lrc` sidecar files
- **Auto-fetch** synced lyrics online (LRCLIB) when a track loads
- **)** toggle karaoke overlay

### Streamer / OBS
- **1** toggle semi-transparent topmost overlay window (visualizer only)
- **2** streamer frame capture notes
- **6** party mode (kaleidoscope + karaoke + boosted visuals)

### Audio enhancements
- Software 3-band visual EQ (bass/mid/treble affects bars)
- **=** crossfade between tracks
- BPM + mood detection in status line

### DJ mode
- **~** toggle DJ mode
- **;** load current track to deck B
- **,** / **.** move crossfader

### Remote control
- Open **http://localhost:8765** on your PC or phone (same Wi‑Fi)
- Play/pause/next/prev from browser
- **4** export sync bundle (favorites, presets, play counts) to Desktop

### UI
- **3** settings panel
- Clickable playlist rows
- Volume slider (top right)
- **`** ambient mode (slow, dim fullscreen visuals)

## Core features

- Opens or drag-drops audio files and folders
- Plays `.wav`, `.mp3`, `.wma`, `.mid`, `.aiff`, `.au`, `.snd`, `.flac`, `.ogg` (library scan)
- **Desktop Input** — WASAPI loopback from Windows speakers (**I**)
- Playlists, shuffle, repeat, favorites, recent, search
- Album art blurred background, `.lrc` lyrics
- Built-in themes + 5 custom saved themes (**H**)
- System tray, EQ, playback speed, visual-only fullscreen
- Settings in `%APPDATA%\MusicVisualizerPro\settings.json`

## Keyboard shortcuts

| Key | Action |
|-----|--------|
| `O` | Open file |
| `Space` | Play/pause |
| `V` | Cycle visualizer |
| `G` | Next theme |
| `I` | Desktop input |
| `1` | OBS overlay |
| `2` | Streamer capture |
| `3` | Settings panel |
| `6` | Party mode |
| `7` | Rescan library |
| `8` / `9` | Cycle / save preset |
| `0` | Auto-preset |
| `\` | Cycle library panels |
| `)` | Karaoke lyrics |
| `` ` `` | Ambient mode |
| `~` | DJ mode |
| `=` | Crossfade |
| `,` / `.` | DJ crossfader |
| `;` | Load DJ deck B |
| `U` / `Y` / `/` | Recent / favorites / search |
| `Z` | Visual-only |
| `Esc` | Quit |

## Build from source (optional)

```bat
build-windows.bat
```

## Notes

- MP3/WAV get real waveform analysis; other formats use animated visualizers via MCI playback
- Desktop input visualizes Spotify, YouTube, games, etc.
- Place a matching `.lrc` next to a song for offline synced lyrics
