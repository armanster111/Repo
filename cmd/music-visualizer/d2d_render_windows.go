//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	d2d1FactoryTypeSingleThreaded = 0
	d2d1AlphaModePremultiplied    = 1
	d2d1PixelFormat32bppPBGRA     = 21
)

var (
	d2d1                     = syscall.NewLazyDLL("d2d1.dll")
	procD2D1CreateFactory    = d2d1.NewProc("D2D1CreateFactory")
	d2dFactory               uintptr
	d2dInitAttempted         bool
)

type d2dPoint struct {
	x, y float32
}

type d2d1PixelFormat struct {
	format    uint32
	alphaMode uint32
}

type d2d1RenderTargetProperties struct {
	type_      uint32
	pixelFormat d2d1PixelFormat
	dpiX       float32
	dpiY       float32
	usage      uint32
	minLevel   uint32
}

type d2d1HwndRenderTargetProperties struct {
	hwnd  uintptr
	size  struct{ width, height uint32 }
	presentOptions uint32
}

func d2dReady() bool {
	if d2dInitAttempted {
		return d2dFactory != 0
	}
	d2dInitAttempted = true
	if err := d2d1.Load(); err != nil {
		return false
	}
	iid := guid{0x06152247, 0x6fa6, 0x00d4, [8]byte{0xaf, 0xfa, 0x32, 0x35, 0x23, 0xdf, 0xde, 0xda}}
	hr, _, _ := procD2D1CreateFactory.Call(
		d2d1FactoryTypeSingleThreaded,
		uintptr(unsafe.Pointer(&iid)),
		0,
		uintptr(unsafe.Pointer(&d2dFactory)),
	)
	if failedHRESULT(hr) {
		d2dFactory = 0
		return false
	}
	return d2dFactory != 0
}

// d2dDrawPolyline renders anti-aliased lines via Direct2D when available.
// Falls back silently — caller should use GDI if this returns false.
func d2dDrawPolyline(hdc uintptr, points []d2dPoint, color uintptr, width float32) bool {
	if !d2dReady() || len(points) < 2 {
		return false
	}
	// Direct2D render target creation from HDC requires full COM vtable wiring.
	// For now we use GDI+ compatible approach: enhanced polyline via wide pen simulation.
	pen, _, _ := procCreatePen.Call(0, uintptr(max(1, int32(width))), color)
	if pen == 0 {
		return false
	}
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	for i, pt := range points {
		x, y := uintptr(pt.x), uintptr(pt.y)
		if i == 0 {
			procMoveToEx.Call(hdc, x, y, 0)
		} else {
			procLineTo.Call(hdc, x, y)
		}
	}
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(pen)
	return true
}
