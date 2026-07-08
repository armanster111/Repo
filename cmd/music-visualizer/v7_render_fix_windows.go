//go:build windows

package main

import (
	"math"
)

var (
	procCreateRectRgn  = gdi32.NewProc("CreateRectRgn")
	procSelectClipRgn  = gdi32.NewProc("SelectClipRgn")
)

func withClipRgn(hdc uintptr, r rect, fn func()) {
	hrgn, _, _ := procCreateRectRgn.Call(
		uintptr(r.left), uintptr(r.top), uintptr(r.right), uintptr(r.bottom),
	)
	if hrgn == 0 {
		fn()
		return
	}
	defer procDeleteObject.Call(hrgn)
	old, _, _ := procSelectClipRgn.Call(hdc, hrgn)
	fn()
	procSelectClipRgn.Call(hdc, old)
}

func drawChromeHeader(hdc uintptr, width int32, palette palette, chrome uiChromeLayout) {
	headerBottom := chrome.vizTop
	if headerBottom < chrome.buttonTop+chrome.buttonH+8 {
		headerBottom = chrome.buttonTop + chrome.buttonH + 8
	}
	fill(hdc, rect{left: 0, top: 0, right: width, bottom: headerBottom}, palette.background)
}

func (s *appState) displayStatus() string {
	if s.statusFlash != "" && !s.statusFlashUntil.IsZero() {
		return s.statusFlash
	}
	return s.status
}

func (s *appState) barsChanged(bars []float64) bool {
	if len(bars) != len(s.lastPaintBars) {
		s.lastPaintBars = append([]float64(nil), bars...)
		return true
	}
	var delta float64
	for i, v := range bars {
		delta += math.Abs(v - s.lastPaintBars[i])
	}
	if delta < 0.008 {
		return false
	}
	copy(s.lastPaintBars, bars)
	return true
}

func (s *appState) needsAnimation() bool {
	if s.gifRecording || s.mp4Recording {
		return true
	}
	if s.modeBlend < 1 {
		return true
	}
	if s.ultra != nil && s.ultra.showcase {
		return true
	}
	if s.desktopMode && s.desktop != nil && s.desktop.isActive() {
		return true
	}
	if s.playing {
		return true
	}
	if s.uiDirty {
		return true
	}
	// Subtle ambient motion when idle with no file.
	if s.filePath == "" && !s.desktopMode {
		return true
	}
	return false
}

func (s *appState) markUIDirty() {
	s.uiDirty = true
}

func invalidateIfNeeded() {
	if app.hwnd == 0 {
		return
	}
	if app.needsAnimation() {
		procInvalidateRect.Call(app.hwnd, 0, 0)
	}
}
