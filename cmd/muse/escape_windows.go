//go:build windows

package main

// handleEscapeKey exits layered fullscreen modes before quitting the app.
func (s *appState) handleEscapeKey() {
	if s.ultra != nil && s.ultra.showcase {
		s.ultra.exitShowcase()
		return
	}
	if s.visualOnly {
		s.visualOnly = false
		procShowWindow.Call(s.hwnd, swShowNormal)
		s.updateStatus()
		invalidate()
		return
	}
	if s.fullscreen {
		s.toggleFullscreen()
		return
	}
	if s.overlayMode {
		s.closeOverlay()
		return
	}
	procDestroyWindow.Call(s.hwnd)
}
