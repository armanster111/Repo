//go:build windows

package main

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/armanster111/music-visualizer/internal/visual"
)

var preloadMu sync.Mutex

func (s *appState) nextPlaylistIndex() int {
	if len(s.playlist) == 0 {
		return -1
	}
	if s.shuffle && len(s.playlist) > 1 {
		return s.pickShuffleIndex()
	}
	idx := s.currentIndex + 1
	if idx >= len(s.playlist) {
		if s.repeat == repeatAll {
			return 0
		}
		return -1
	}
	return idx
}

func (s *appState) tickGaplessPreload() {
	if !s.playing || len(s.playlist) <= 1 {
		return
	}
	nextIdx := s.nextPlaylistIndex()
	if nextIdx < 0 {
		return
	}
	path := s.playlist[nextIdx].Path
	if strings.EqualFold(s.preloadPath, path) && len(s.preloadFrames) > 0 {
		return
	}
	remaining := s.duration - s.position()
	trigger := 8 * time.Second
	if s.crossfade {
		trigger = 12 * time.Second
	}
	if remaining > trigger || s.preloadBusy {
		return
	}
	s.preloadBusy = true
	go func(p string) {
		frames, _ := s.buildFramesForPath(p)
		preloadMu.Lock()
		s.preloadPath = filepath.Clean(p)
		s.preloadFrames = frames
		s.preloadBusy = false
		preloadMu.Unlock()
	}(path)
}

func (s *appState) takePreloadedFrames(path string) []visual.Frame {
	preloadMu.Lock()
	defer preloadMu.Unlock()
	clean := filepath.Clean(path)
	if !strings.EqualFold(s.preloadPath, clean) || len(s.preloadFrames) == 0 {
		return nil
	}
	frames := s.preloadFrames
	s.preloadFrames = nil
	s.preloadPath = ""
	return frames
}
