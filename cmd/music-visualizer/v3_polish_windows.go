//go:build windows

package main

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

func init() {
	// math/rand auto-seeded on modern Go; shuffle uses global rand.
	_ = rand.Intn(1)
}

func (s *appState) flashStatus(msg string, d time.Duration) {
	s.statusFlash = msg
	s.statusFlashUntil = time.Now().Add(d)
	s.status = msg
	invalidate()
}

func (s *appState) tickStatusFlash() {
	if s.statusFlash != "" && time.Now().After(s.statusFlashUntil) {
		s.statusFlash = ""
		s.updateStatus()
		invalidate()
	}
}

func (s *appState) noteMouseActivity() {
	s.lastMouseMove = time.Now()
	if s.cinemaHideUI {
		s.cinemaHideUI = false
		invalidate()
	}
}

func (s *appState) tickCinemaUI() {
	if s.ultra == nil || !s.ultra.showcase {
		s.cinemaHideUI = false
		return
	}
	hide := time.Since(s.lastMouseMove) > 3500*time.Millisecond
	if hide != s.cinemaHideUI {
		s.cinemaHideUI = hide
		invalidate()
	}
}

func (s *appState) cinemaUIVisible() bool {
	if s.ultra == nil || !s.ultra.showcase {
		return true
	}
	return !s.cinemaHideUI
}

func (s *appState) pickShuffleIndex() int {
	if len(s.playlist) <= 1 {
		return s.currentIndex
	}
	next := s.currentIndex
	for tries := 0; tries < 8 && next == s.currentIndex; tries++ {
		next = rand.Intn(len(s.playlist))
	}
	return next
}

func (s *appState) cycleOverlayOpacity(delta int) {
	if !s.overlayMode {
		return
	}
	v := int(s.overlayAlpha) + delta
	if v < 80 {
		v = 80
	}
	if v > 255 {
		v = 255
	}
	s.overlayAlpha = byte(v)
	if s.overlayHWND != 0 {
		procSetLayeredWindowAttributes.Call(s.overlayHWND, 0, uintptr(s.overlayAlpha), lwaAlpha)
	}
	s.saveSettings()
	s.flashStatus(fmt.Sprintf("Overlay opacity: %d%%", int(float64(s.overlayAlpha)/255*100)), 2*time.Second)
}

func (s *appState) saveOverlayBounds() {
	if s.overlayHWND == 0 {
		return
	}
	var r winRect
	procGetWindowRect.Call(s.overlayHWND, uintptr(unsafe.Pointer(&r)))
	s.overlayX = r.left
	s.overlayY = r.top
	s.overlayW = r.right - r.left
	s.overlayH = r.bottom - r.top
	s.saveSettings()
}

func (s *appState) restoreOverlayIfNeeded() {
	if !s.overlayMode {
		return
	}
	s.openOverlay()
}

func (s *appState) handleVisualizerDoubleClick(x, y, width, height int32) bool {
	top := int32(112)
	bottom := height - 104
	if app.mini || app.visualOnly {
		top = 28
		bottom = height - 28
	}
	area := rect{left: 28, top: top, right: width - 28, bottom: bottom}
	if !pointInRect(x, y, area) {
		return false
	}
	if app.ultra != nil {
		app.ultra.toggleShowcase()
	}
	return true
}