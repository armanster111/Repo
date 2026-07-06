//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const panelPageSize = 8

var framesLoadMu sync.Mutex

func hasFFTForPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".wav", ".mp3", ".flac", ".ogg", ".oga":
		return true
	default:
		return false
	}
}

func (s *appState) loadFramesAsync(path string) {
	if !hasFFTForPath(path) {
		return
	}
	clean := filepath.Clean(path)
	framesLoadMu.Lock()
	if s.framesLoading && strings.EqualFold(s.framesLoadingPath, clean) {
		framesLoadMu.Unlock()
		return
	}
	s.framesLoading = true
	s.framesLoadingPath = clean
	framesLoadMu.Unlock()

	go func(p string) {
		frames, dur := s.buildFramesForPath(p)
		framesLoadMu.Lock()
		s.framesLoading = false
		framesLoadMu.Unlock()
		if !strings.EqualFold(filepath.Clean(s.filePath), p) {
			return
		}
		if len(frames) == 0 {
			return
		}
		s.frames = frames
		if dur > 0 {
			s.duration = dur
		}
		s.flashStatus("Spectrum analysis ready.", 2*time.Second)
		invalidate()
	}(clean)
}

func (s *appState) clampPanelRow(total int) {
	maxStart := total - panelPageSize
	if maxStart < 0 {
		maxStart = 0
	}
	if s.panelRow > maxStart {
		s.panelRow = maxStart
	}
	if s.panelRow < 0 {
		s.panelRow = 0
	}
}

func (s *appState) scrollPlaylist(delta int) {
	s.panelRow += delta
	if s.panelRow < 0 {
		s.panelRow = 0
	}
	invalidate()
}

func (s *appState) playlistScrollMax() int {
	list := s.filteredPlaylist()
	groups := s.libraryGroupsForPanel()
	entries := s.libraryEntriesForPanel()
	total := len(list)
	if total == 0 {
		total = len(groups)
	}
	if total == 0 {
		total = len(entries)
	}
	maxStart := total - panelPageSize
	if maxStart < 0 {
		return 0
	}
	return maxStart
}

func (s *appState) handleSliderDrag(x, y int32) {
	switch s.sliderDrag {
	case "volume":
		s.handleVolumeSlider(x, y)
	case "intensity":
		s.handleIntensitySlider(x, y)
	}
}

func (s *appState) beginSliderDrag(x, y int32) bool {
	if s.showSettings && pointInRect(x, y, s.intensitySlider) {
		s.sliderDrag = "intensity"
		procSetCapture.Call(s.hwnd)
		s.handleIntensitySlider(x, y)
		return true
	}
	if pointInRect(x, y, s.volumeSliderBounds) {
		s.sliderDrag = "volume"
		procSetCapture.Call(s.hwnd)
		s.handleVolumeSlider(x, y)
		return true
	}
	return false
}

func (s *appState) endSliderDrag() {
	if s.sliderDrag != "" {
		s.sliderDrag = ""
		procReleaseCapture.Call()
	}
}

func (s *appState) exportSnapshotPNG() {
	frame := s.renderExportFrame(960, 540)
	if frame == nil {
		s.status = "No visualizer snapshot to export yet."
		invalidate()
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, "Desktop")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		dir = home
	}
	path := filepath.Join(dir, "music-visualizer-"+time.Now().Format("20060102-150405")+".png")
	if err := frame.WritePNG(path); err != nil {
		s.status = "Snapshot failed: " + err.Error()
	} else {
		s.flashStatus("Saved PNG snapshot: "+path, 4*time.Second)
	}
	invalidate()
}

func drawScrollHint(hdc uintptr, left, bottom int32, palette palette, start, shown, total int) {
	if total <= panelPageSize {
		return
	}
	procSetTextColor.Call(hdc, palette.dim)
	end := start + shown
	if end > total {
		end = total
	}
	textOut(hdc, left+12, bottom-16, formatScrollRange(start+1, end, total))
}

func formatScrollRange(from, to, total int) string {
	return itoa(from) + "-" + itoa(to) + " of " + itoa(total) + "  (scroll)"
}
