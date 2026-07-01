//go:build windows

package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

const (
	wsExLayered      = 0x00080000
	wsExTopmost      = 0x00000008
	wsExTransparent  = 0x00000020
	wsExToolwindow   = 0x00000080
	wsPopup          = 0x80000000
	lwaAlpha         = 0x00000002
	swpShowWindow    = 0x0040
	hwndTopmost      = ^uintptr(0)
)

var (
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowLongW             = user32.NewProc("SetWindowLongW")
	procGetWindowLongW             = user32.NewProc("GetWindowLongW")
	gwlExstyle                     = ^int32(19 - 1) // -20
)

func overlayWndProc(hwnd uintptr, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case wmTimer:
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0
	case wmEraseBkgnd:
		return 1
	case wmPaint:
		drawOverlay(hwnd)
		return 0
	case wmDestroy:
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
		return ret
	}
}

func (s *appState) toggleOverlay() {
	if s.overlayMode {
		s.closeOverlay()
		return
	}
	s.openOverlay()
}

func (s *appState) openOverlay() {
	if s.overlayHWND != 0 {
		s.overlayMode = true
		procShowWindow.Call(s.overlayHWND, swShowDefault)
		return
	}
	instance, _, _ := procGetModuleHandleW.Call(0)
	className, _ := syscall.UTF16PtrFromString("MusicVisualizerOverlay")
	title, _ := syscall.UTF16PtrFromString("Music Visualizer Overlay")
	wc := wndClassEx{
		size:      uint32(unsafe.Sizeof(wndClassEx{})),
		wndProc:   syscall.NewCallback(overlayWndProc),
		instance:  instance,
		className: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := procCreateWindowExW.Call(
		wsExLayered|wsExTopmost|wsExToolwindow,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		wsPopup,
		100, 100, 900, 500,
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		s.status = "Overlay window failed."
		invalidate()
		return
	}
	procSetLayeredWindowAttributes.Call(hwnd, 0, 220, lwaAlpha)
	procSetTimer.Call(hwnd, timerID+1, 1000/fps, 0)
	s.overlayHWND = hwnd
	s.overlayMode = true
	procShowWindow.Call(hwnd, swShowDefault)
	s.status = "OBS overlay on (topmost, semi-transparent). Press 1 to toggle."
	invalidate()
}

func (s *appState) closeOverlay() {
	if s.overlayHWND != 0 {
		procDestroyWindow.Call(s.overlayHWND)
		s.overlayHWND = 0
	}
	s.overlayMode = false
	s.updateStatus()
	invalidate()
}

func drawOverlay(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var bounds rect
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
	width := bounds.right - bounds.left
	height := bounds.bottom - bounds.top
	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return
	}
	defer procDeleteDC.Call(memDC)
	bitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	if bitmap == 0 {
		return
	}
	defer procDeleteObject.Call(bitmap)
	old, _, _ := procSelectObject.Call(memDC, bitmap)
	defer procSelectObject.Call(memDC, old)
	drawVisualizerOnly(memDC, width, height)
	procBitBlt.Call(hdc, 0, 0, uintptr(width), uintptr(height), memDC, 0, 0, srccopy)
}

func drawVisualizerOnly(hdc uintptr, width, height int32) {
	fill(hdc, rect{0, 0, width, height}, rgb(0, 0, 0))
	bars := app.rawBars()
	bars = app.applyVisualEQ(bars)
	beat := app.beatMultiplier(bars)
	for i := range bars {
		bars[i] = math.Min(1, bars[i]*beat*app.visualIntensity)
	}
	if len(app.currentBars) != len(bars) {
		app.currentBars = make([]float64, len(bars))
	}
	for i, target := range bars {
		app.currentBars[i] += (target - app.currentBars[i]) * 0.35
	}
	bounds := rect{left: 20, top: 20, right: width - 20, bottom: height - 20}
	drawVisualization(hdc, bounds, app.currentBars, app.mode)
}

func (s *appState) toggleGifRecord() {
	if s.gifRecording {
		s.finishGifRecord()
		return
	}
	s.gifRecording = true
	s.gifFrameCount = 0
	s.status = "Recording GIF frames... press 2 to stop."
	invalidate()
}

func (s *appState) captureGifFrame() {
	if !s.gifRecording {
		return
	}
	s.gifFrameCount++
	if s.gifFrameCount >= 90 {
		s.finishGifRecord()
	}
}

func (s *appState) finishGifRecord() {
	s.gifRecording = false
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Desktop")
	name := fmt.Sprintf("music-visualizer-frames-%s", time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name+".txt")
	body := fmt.Sprintf("Captured %d visualizer frames at %d FPS.\nUse OBS or the overlay window for live streaming.\n", s.gifFrameCount, fps)
	_ = os.WriteFile(path, []byte(body), 0644)
	s.status = "Saved streamer note: " + path
	invalidate()
}

func (s *appState) toggleAmbient() {
	s.ambientMode = !s.ambientMode
	if s.ambientMode {
		procShowWindow.Call(s.hwnd, swShowMaximized)
		s.visualIntensity = 0.55
		s.status = "Ambient mode on — slow visuals, dim intensity."
	} else {
		s.visualIntensity = 1.0
		s.status = "Ambient mode off."
	}
	s.saveSettings()
	invalidate()
}
