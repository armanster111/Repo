//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/armanster111/music-visualizer/internal/audio"
	"github.com/armanster111/music-visualizer/internal/lyrics"
	"github.com/armanster111/music-visualizer/internal/metadata"
	"github.com/armanster111/music-visualizer/internal/visual"
	"github.com/armanster111/music-visualizer/internal/wav"
)

type panelView int

const (
	viewQueue panelView = iota
	viewRecent
	viewFavorites
	viewSearch
	viewLibrary
	viewAlbums
	viewArtists
	viewSmart
)

type customTheme struct {
	Name       string `json:"name"`
	Background uint32 `json:"background"`
	Panel      uint32 `json:"panel"`
	Text       uint32 `json:"text"`
	Dim        uint32 `json:"dim"`
	Accent     uint32 `json:"accent"`
	Accent2    uint32 `json:"accent2"`
}

type trackMeta struct {
	Title  string
	Artist string
	Album  string
}

func (s *appState) loadTrackMedia(path string) {
	info, art, _ := metadata.Read(path)
	s.meta = trackMeta{Title: info.Title, Artist: info.Artist, Album: info.Album}
	s.lyricLines = lyrics.Load(path)
	s.artGrid = metadata.ArtGridFromBytes(art, 12, 8)
	s.fetchLyricsAsync()
}

func (s *appState) buildFramesForPath(path string) ([]visual.Frame, time.Duration) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".wav":
		audioData, err := wav.DecodeFile(path)
		if err != nil {
			return nil, 0
		}
		frames := visual.BuildFrames(audioData.Mono, audioData.SampleRate, fps, barCount)
		return frames, time.Duration(audioData.DurationSeconds() * float64(time.Second))
	case ".mp3":
		mono, sampleRate, err := audio.DecodeFileMP3(path)
		if err != nil || len(mono) == 0 {
			return nil, 0
		}
		frames := visual.BuildFrames(mono, sampleRate, fps, barCount)
		duration := time.Duration(float64(len(mono)) / float64(sampleRate) * float64(time.Second))
		return frames, duration
	default:
		return nil, 0
	}
}

func (s *appState) cyclePanelView() {
	s.panel = (s.panel + 1) % 4
	s.updateStatus()
	invalidate()
}

func (s *appState) showRecentView() { s.panel = viewRecent; invalidate() }
func (s *appState) showFavoritesView() { s.panel = viewFavorites; invalidate() }
func (s *appState) showSearchView() { s.panel = viewSearch; invalidate() }

func (s *appState) toggleVisualOnly() {
	s.visualOnly = !s.visualOnly
	if s.visualOnly {
		procShowWindow.Call(s.hwnd, swShowMaximized)
	}
	invalidate()
}

func (s *appState) cyclePlaybackSpeed() {
	switch s.playbackSpeed {
	case 750:
		s.playbackSpeed = 1000
	case 1000:
		s.playbackSpeed = 1250
	case 1250:
		s.playbackSpeed = 1500
	default:
		s.playbackSpeed = 750
	}
	s.applyPlaybackSpeed()
	s.updateStatus()
	invalidate()
}

func (s *appState) applyPlaybackSpeed() {
	_ = mciSend(fmt.Sprintf("set musicvisualizer_track speed to %d", s.playbackSpeed))
}

func (s *appState) cycleEQBand() {
	switch s.eqFocus {
	case 0:
		s.eqBass += 200
		if s.eqBass > 1000 {
			s.eqBass = -1000
		}
	case 1:
		s.eqMid += 200
		if s.eqMid > 1000 {
			s.eqMid = -1000
		}
	default:
		s.eqTreble += 200
		if s.eqTreble > 1000 {
			s.eqTreble = -1000
		}
	}
	s.applyEqualizer()
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) focusNextEQBand() {
	s.eqFocus = (s.eqFocus + 1) % 3
	s.updateStatus()
	invalidate()
}

func (s *appState) applyEqualizer() {
	_ = mciSend(fmt.Sprintf("setaudio musicvisualizer_track bass to %d", clampInt(s.eqBass, -1000, 1000)))
	_ = mciSend(fmt.Sprintf("setaudio musicvisualizer_track treble to %d", clampInt(s.eqTreble, -1000, 1000)))
}

func (s *appState) cycleDesktopSensitivity() {
	switch s.desktopSensitivity {
	case 1.0:
		s.desktopSensitivity = 1.5
	case 1.5:
		s.desktopSensitivity = 2.0
	case 2.0:
		s.desktopSensitivity = 0.75
	default:
		s.desktopSensitivity = 1.0
	}
	if s.desktop != nil {
		s.desktop.setSensitivity(s.desktopSensitivity)
	}
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) cycleDesktopDevice() {
	s.desktopDevice++
	if s.desktop != nil {
		if err := s.desktop.setDeviceIndex(s.desktopDevice); err != nil {
			s.desktopDevice = 0
			_ = s.desktop.setDeviceIndex(0)
			s.status = "Desktop device reset to default output."
		}
	}
	if s.desktopMode {
		s.stopDesktopInput()
		_ = s.desktop.start()
	}
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) saveCustomTheme() {
	if len(s.customThemes) >= 5 {
		s.customThemes = s.customThemes[1:]
	}
	p := currentPalette()
	s.customThemes = append(s.customThemes, customTheme{
		Name:       fmt.Sprintf("Custom %d", len(s.customThemes)+1),
		Background: uint32(p.background & 0xffffff),
		Panel:      uint32(p.panel & 0xffffff),
		Text:       uint32(p.text & 0xffffff),
		Dim:        uint32(p.dim & 0xffffff),
		Accent:     uint32(p.accent & 0xffffff),
		Accent2:    uint32(p.accent2 & 0xffffff),
	})
	s.theme = themeCustomBase + theme(len(s.customThemes)-1)
	s.saveSettings()
	s.updateStatus()
	invalidate()
}

func (s *appState) toggleDesktopAuto() {
	s.desktopAuto = !s.desktopAuto
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) maybeAutoDesktop() {
	if !s.desktopAuto || s.desktopMode || s.playing {
		return
	}
	if s.filePath == "" {
		_ = s.desktop.start()
		s.desktopMode = true
		s.updateStatus()
	}
}

func (s *appState) showOnboardingIfNeeded() {
	if s.seenOnboarding {
		return
	}
	s.seenOnboarding = true
	s.saveSettings()
	showMessage(s.hwnd, appTitle+" Pro",
		"Welcome to Music Visualizer Pro v2!\n\nLibrary scan starts automatically.\n1 OBS overlay | 3 settings | 7 rescan library\n8 cycle preset | 9 save preset | 6 party mode\nRemote control: http://localhost:8765\n\nO open | I desktop | V viz | Z visual-only",
		0)
}

func (s *appState) restoreWindowBounds() {
	if s.windowW < 400 || s.windowH < 300 {
		return
	}
	procSetWindowPos.Call(s.hwnd, 0, uintptr(s.windowX), uintptr(s.windowY), uintptr(s.windowW), uintptr(s.windowH), 0x0040)
}

func (s *appState) saveWindowBounds() {
	var rect winRect
	procGetWindowRect.Call(s.hwnd, uintptr(unsafe.Pointer(&rect)))
	s.windowX = rect.left
	s.windowY = rect.top
	s.windowW = rect.right - rect.left
	s.windowH = rect.bottom - rect.top
	s.saveSettings()
}

func (s *appState) appendSearchChar(ch rune) {
	if ch == '\b' {
		if len(s.searchQuery) > 0 {
			s.searchQuery = s.searchQuery[:len(s.searchQuery)-1]
		}
	} else if ch >= 32 && ch <= 126 && len(s.searchQuery) < 40 {
		s.searchQuery += string(ch)
	}
	s.panel = viewSearch
	invalidate()
}

func (s *appState) filteredPlaylist() []track {
	switch s.panel {
	case viewRecent:
		var out []track
		for _, path := range s.recent {
			out = append(out, makeTrack(path, s.favorites))
		}
		return out
	case viewFavorites:
		var out []track
		for path, ok := range s.favorites {
			if ok {
				out = append(out, makeTrack(path, s.favorites))
			}
		}
		sortTracks(out)
		return out
	case viewSearch:
		q := strings.ToLower(strings.TrimSpace(s.searchQuery))
		if q == "" {
			return s.playlist
		}
		var out []track
		for _, t := range s.playlist {
			if strings.Contains(strings.ToLower(t.Title), q) || strings.Contains(strings.ToLower(filepath.Base(t.Path)), q) {
				out = append(out, t)
			}
		}
		return out
	default:
		return s.playlist
	}
}

func (s *appState) panelTitle() string {
	switch s.panel {
	case viewRecent:
		return "Recent"
	case viewFavorites:
		return "Favorites"
	case viewSearch:
		return "Search: " + s.searchQuery
	case viewLibrary, viewAlbums, viewArtists, viewSmart:
		return s.libraryPanelTitle()
	default:
		return "Queue"
	}
}

func (s *appState) currentLyric() string {
	return lyrics.At(s.lyricLines, s.position())
}

func (s *appState) noteEnergy(energy float64) {
	s.energyHistory = append(s.energyHistory, energy)
	if len(s.energyHistory) > 24 {
		s.energyHistory = s.energyHistory[len(s.energyHistory)-24:]
	}
}

func (s *appState) beatMultiplier(bars []float64) float64 {
	var sum float64
	for _, b := range bars {
		sum += b
	}
	energy := sum / float64(max(1, len(bars)))
	s.noteEnergy(energy)
	return visual.BeatPulse(s.energyHistory, energy)
}

func drawAlbumArtBackground(hdc uintptr, bounds rect, grid *metadata.ArtGrid) {
	if grid == nil || len(grid.Colors) == 0 {
		return
	}
	cellW := max(int32(2), (bounds.right-bounds.left)/int32(grid.Cols))
	cellH := max(int32(2), (bounds.bottom-bounds.top)/int32(grid.Rows))
	for y := 0; y < grid.Rows; y++ {
		for x := 0; x < grid.Cols; x++ {
			c := grid.Colors[y*grid.Cols+x]
			fill(hdc, rect{
				left:   bounds.left + int32(x)*cellW,
				top:    bounds.top + int32(y)*cellH,
				right:  bounds.left + int32(x+1)*cellW,
				bottom: bounds.top + int32(y+1)*cellH,
			}, uintptr(c))
		}
	}
}

func paletteFromCustom(ct customTheme) palette {
	return palette{
		background: uintptr(ct.Background),
		panel:      uintptr(ct.Panel),
		text:       uintptr(ct.Text),
		dim:        uintptr(ct.Dim),
		accent:     uintptr(ct.Accent),
		accent2:    uintptr(ct.Accent2),
	}
}

func installToLocalAppData() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	target := filepath.Join(home, "AppData", "Local", "MusicVisualizerPro")
	_ = os.MkdirAll(target, 0755)
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(target, "MusicVisualizerPro.exe"), data, 0755)
}

type winRect struct {
	left, top, right, bottom int32
}
