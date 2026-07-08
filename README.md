# Music Visualizer Ultra

The most advanced standalone Windows music visualizer — **60 FPS**, **mel-scale FFT spectrum analysis**, **25 visual modes**, cinema showcase, OBS overlay, microphone input, library, lyrics, DJ mode, and LAN remote control. One `.exe`, no install.

## Download

1. **[MusicVisualizerPro-Windows.zip](https://github.com/armanster111/Repo/raw/cursor/music-visualizer-exe-7b99/dist/MusicVisualizerPro-Windows.zip)**
2. Extract → double-click **`MusicVisualizerPro.exe`**

Direct EXE: https://github.com/armanster111/Repo/raw/cursor/music-visualizer-exe-7b99/dist/MusicVisualizerPro.exe

---

## What makes it best-in-class

| Feature | Details |
|--------|---------|
| **60 FPS** | Silky-smooth rendering with double-buffered output |
| **Mel-scale FFT** | Perceptual frequency analysis (A-weighted, 64 mel bands) for all decodable formats |
| **25 visual modes** | Including Spectrogram, Oscilloscope, Lissajous, Neon City, Supernova, Liquid, and more |
| **Live FFT** | Real-time spectrum during playback — reacts to volume/EQ changes (F5) |
| **Mood-reactive** | Palette and mode adapt to detected mood; BPM-synced bass pulse (F6) |
| **Mode crossfade** | Smooth transitions when cycling visualizers (V) |
| **Microphone input** | WASAPI capture — press I to cycle Player → Desktop → Mic |
| **Motion trails & bloom** | Afterglow trails + peak-hold glow on every mode |
| **Beat detection** | Spectral-flux onset detection for punchy beat hits |
| **Cinema mode** | **F9** — fullscreen auto-cycling premium visuals every 12 seconds |
| **Album-reactive colors** | Palette tints from album art automatically |
| **Desktop audio** | WASAPI loopback with FFT — visualizes Spotify, YouTube, games |
| **OBS overlay** | **1** — transparent topmost window; **F10** cycles 16:9 / 4:3 / 1:1 aspect; F7 chroma key |
| **Online lyrics** | Auto-fetch + karaoke overlay |
| **Music library** | Artist/album/top-played views with click-to-play |
| **Presets & packs** | Save/share looks; drop JSON packs in `%APPDATA%\MusicVisualizerPro\packs\` |
| **DJ mode** | Dual-deck crossfader |
| **Remote** | http://localhost:8765 |

---

## Quick controls

| Key | Action |
|-----|--------|
| **F5** | Toggle live FFT |
| **F6** | Toggle mood-reactive visuals |
| **F7** | Toggle chroma key (OBS green screen) |
| **F9** | Cinema showcase (best demo mode) |
| **V** | Cycle 25 visualizers |
| **I** | Cycle input: Player → Desktop → Mic |
| **F10** | Cycle overlay aspect ratio (16:9 / 4:3 / 1:1) |
| **1** | OBS overlay |
| **3** | Settings panel |
| **7** | Rescan library |
| **8** / **9** | Cycle / save preset |
| **6** | Party mode |
| **Z** | Visual-only fullscreen |
| **Space** | Play/pause |

---

## Visual modes (press V)

Classic · Mirror · Blocks · Wave · Halo · Spectrum · Fire · Particles · Tunnel · Plasma · Aurora · Mandala · Starfield · Kaleidoscope · **Neon City** · **Supernova** · **Liquid** · **Orbit** · **Waveform 3D** · Fluid · Galaxy · Chromatic · **Spectrogram** · **Oscilloscope** · **Lissajous**

---

## For streamers

1. Press **F9** for cinema mode (looks incredible on stream)
2. Press **1** for OBS overlay window
3. Add Window Capture in OBS → pick "Music Visualizer Overlay"

---

## Build from source (optional)

```bat
build-windows.bat
```

---

SmartScreen may warn on unsigned apps → **More info → Run anyway**.
