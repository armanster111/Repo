# Music Visualizer Ultra — 10x Edition

The most advanced standalone Windows music visualizer. **30 GPU-accelerated modes**, **kick/snare/hi-hat beat engine**, **live browser visualizer**, adaptive performance, cinema showcase, OBS overlay, mic input, library, lyrics, DJ mode, and LAN remote control. One `.exe`, no install.

## Download

**Latest build:** https://github.com/armanster111/Repo/raw/cursor/10x-upgrade-8f03/dist/MusicVisualizerPro-Windows.zip

Extract → double-click `MusicVisualizerPro.exe`

---

## What makes it 10x

| Pillar | What's new |
|--------|------------|
| **30 visual modes** | Vortex, Comets, Bass Drop, Rain, Prism + all previous modes |
| **Kick/snare engine** | Separate onset detection for bass, mid, treble — visuals punch on kicks |
| **Full-res GPU canvas** | Fluid, Galaxy, Chromatic, and 5 new shaders at native resolution with bloom |
| **Live browser viz** | Open **http://localhost:8765/viz** on any device on your LAN — real-time spectrum |
| **Adaptive performance** | Auto-scales quality when frames drop — stays smooth on any PC |
| **Mel-scale FFT** | Perceptual A-weighted analysis for all formats + desktop/mic input |
| **12 premium presets** | Bass Drop, Vortex Dream, Comet Storm, Supernova Cinema, and more (press 8) |
| **OBS-ready** | Overlay, chroma key (F7), aspect ratios (F10), 60 FPS |

---

## Quick controls

| Key | Action |
|-----|--------|
| **V** | Cycle 30 visualizers |
| **F5** | Toggle live FFT |
| **F6** | Toggle mood-reactive colors |
| **F7** | Chroma key (OBS green screen) |
| **F9** | Cinema showcase |
| **F10** | Overlay aspect ratio |
| **I** | Cycle input: Player → Desktop → Mic |
| **8** / **9** | Cycle / save preset |
| **1** | OBS overlay |
| **3** | Settings panel |

---

## Live browser visualizer

1. Launch the app (remote control is on by default)
2. Open **http://localhost:8765/viz** in Chrome/Edge/Firefox
3. Full-screen the browser tab on a second monitor or stream it via OBS Browser Source

API: `GET http://localhost:8765/api/bars` returns JSON with 64-band spectrum, kick/snare levels, BPM, mood, and track info.

---

## All 30 modes (press V)

Classic · Mirror · Blocks · Wave · Halo · Spectrum · Fire · Particles · Tunnel · Plasma · Aurora · Mandala · Starfield · Kaleidoscope · Neon City · Supernova · Liquid · Orbit · Waveform 3D · Fluid · Galaxy · Chromatic · Spectrogram · Oscilloscope · Lissajous · **Vortex** · **Comets** · **Bass Drop** · **Rain** · **Prism**

---

## For streamers

1. Press **F9** for cinema mode
2. Press **1** for OBS overlay window
3. Add Browser Source → `http://localhost:8765/viz` for a zero-CPU overlay option
4. **F7** enables green-screen chroma key on the native overlay

---

## Build from source

```bat
build-windows.bat
```

SmartScreen may warn on unsigned apps → **More info → Run anyway**.
