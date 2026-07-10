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

	"github.com/armanster111/muse/internal/visual"
)

const (
	wsExLayered      = 0x00080000
	wsExTopmost      = 0x00000008
	wsExToolwindow   = 0x00000080
	wsPopup          = 0x80000000
	wsCaption        = 0x00C00000
	wsSysMenu        = 0x00080000
	wsMinimizeBox    = 0x00020000
	lwaAlpha         = 0x00000002
	swpShowWindow    = 0x0040
	htClient         = 1
	htCaption        = 2
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
	case wmKeyDown:
		if wParam == vkEscape {
			app.closeOverlay()
			return 0
		}
		return 0
	case wmNCHitTest:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
		if ret == htClient {
			return htCaption
		}
		return ret
	case wmMove, wmSize:
		app.saveOverlayBounds()
		return 0
	case wmPaint:
		drawOverlay(hwnd)
		return 0
	case wmDestroy:
		app.overlayHWND = 0
		app.saveOverlayBounds()
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
	s.saveSettings()
}

func (s *appState) openOverlay() {
	if s.overlayHWND != 0 {
		s.overlayMode = true
		procShowWindow.Call(s.overlayHWND, swShowDefault)
		s.saveSettings()
		return
	}
	instance, _, _ := procGetModuleHandleW.Call(0)
	className, _ := syscall.UTF16PtrFromString("MuseOverlay")
	title, _ := syscall.UTF16PtrFromString("Muse Overlay")
	wc := wndClassEx{
		size:      uint32(unsafe.Sizeof(wndClassEx{})),
		wndProc:   syscall.NewCallback(overlayWndProc),
		instance:  instance,
		className: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	x, y, w, h := int32(100), int32(100), int32(900), int32(500)
	if s.overlayW >= 300 && s.overlayH >= 200 {
		x, y, w, h = s.overlayX, s.overlayY, s.overlayW, s.overlayH
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		wsExLayered|wsExTopmost|wsExToolwindow,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		wsPopup|wsCaption|wsSysMenu|wsMinimizeBox,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		s.status = "Overlay window failed."
		invalidate()
		return
	}
	alpha := s.overlayAlpha
	if alpha < 80 {
		alpha = 220
	}
	procSetLayeredWindowAttributes.Call(hwnd, 0, uintptr(alpha), lwaAlpha)
	procSetTimer.Call(hwnd, timerID+1, 1000/fps, 0)
	s.overlayHWND = hwnd
	s.overlayMode = true
	procShowWindow.Call(hwnd, swShowDefault)
	s.status = "OBS overlay — drag to move, [ ] opacity, Esc to close."
	s.saveSettings()
	invalidate()
}

func (s *appState) closeOverlay() {
	if s.overlayHWND != 0 {
		s.saveOverlayBounds()
		procDestroyWindow.Call(s.overlayHWND)
		s.overlayHWND = 0
	}
	s.overlayMode = false
	s.saveSettings()
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
	bars := app.targetBars()
	if app.ultra != nil {
		beat := app.beatMultiplier(bars)
		bars = app.ultra.processBars(bars, app.visualIntensity, beat)
		app.currentBars = bars
		drawWithEffects(hdc, rect{left: 8, top: 8, right: width - 8, bottom: height - 8}, bars, app.ultra.peakBars(), app.ultra.trailFrames(), app.mode)
		return
	}
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
	bounds := rect{left: 8, top: 8, right: width - 8, bottom: height - 8}
	drawVisualization(hdc, bounds, app.currentBars, app.mode)
}

func (s *appState) toggleGifRecord() {
	if s.gifRecording {
		s.finishGifRecord()
		return
	}
	s.gifRecording = true
	s.gifFrameCount = 0
	s.gifFrames = s.gifFrames[:0]
	s.flashStatus("Recording GIF... press 2 to stop (max 120 frames).", 3*time.Second)
}

func (s *appState) captureGifFrame() {
	if !s.gifRecording {
		return
	}
	frame := s.renderExportFrame(480, 270)
	if frame == nil {
		return
	}
	dup := visual.NewCanvas(frame.W, frame.H)
	copy(dup.Pix, frame.Pix)
	s.gifFrames = append(s.gifFrames, dup)
	s.gifFrameCount++
	if s.gifFrameCount >= 120 {
		s.finishGifRecord()
	}
}

func (s *appState) finishGifRecord() {
	if !s.gifRecording {
		return
	}
	s.gifRecording = false
	count := len(s.gifFrames)
	frames := s.gifFrames
	s.gifFrames = nil
	s.gifFrameCount = 0
	if count == 0 {
		s.flashStatus("GIF recording cancelled.", 2*time.Second)
		return
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Desktop")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		dir = home
	}
	path := filepath.Join(dir, fmt.Sprintf("muse-%s.gif", time.Now().Format("20060102-150405")))
	delay := time.Second / time.Duration(fps)
	if err := visual.EncodeGIF(path, frames, delay); err != nil {
		s.flashStatus("GIF export failed: "+err.Error(), 3*time.Second)
		return
	}
	s.flashStatus(fmt.Sprintf("Saved GIF (%d frames): %s", count, path), 4*time.Second)
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
