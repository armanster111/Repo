//go:build windows

package main

import (
	"unsafe"

	"github.com/armanster111/music-visualizer/internal/visual"
)

const (
	biRGB        = 0
	dibRGBColors = 0
)

type bitmapInfoHeader struct {
	size          uint32
	width         int32
	height        int32
	planes        uint16
	bitCount      uint16
	compression   uint32
	sizeImage     uint32
	xPelsPerMeter int32
	yPelsPerMeter int32
	clrUsed       uint32
	clrImportant  uint32
}

type shaderEngine struct {
	canvas *visual.Canvas
	fluid  *visual.FluidState
	fluidT float64
}

func newShaderEngine() *shaderEngine {
	return &shaderEngine{}
}

func (e *shaderEngine) ensure(w, h int) {
	if w < 8 {
		w = 8
	}
	if h < 8 {
		h = 8
	}
	if e.canvas == nil || e.canvas.W != w || e.canvas.H != h {
		e.canvas = visual.NewCanvas(w, h)
		e.fluid = visual.NewFluidState(w/2, h/2)
	}
}

func (e *shaderEngine) render(mode visualMode, bars []float64) *visual.Canvas {
	if e.canvas == nil {
		return nil
	}
	t := visual.Now()
	e.fluidT = t
	switch mode {
	case modeFluid:
		visual.StepFluid(e.fluid, e.canvas, bars, t)
		e.canvas.ApplyBloom(0.85)
	case modeGalaxy:
		visual.DrawGalaxy(e.canvas, bars, t)
		e.canvas.ApplyBloom(0.55)
	case modeChromatic:
		visual.DrawChromaticBurst(e.canvas, bars, t)
		e.canvas.ApplyBloom(0.7)
		e.canvas.ApplyChromaticShift(2)
	default:
		return nil
	}
	return e.canvas
}

func drawShaderMode(hdc uintptr, bounds rect, bars []float64, mode visualMode) bool {
	w := int(bounds.right - bounds.left)
	h := int(bounds.bottom - bounds.top)
	if app.shader == nil {
		app.shader = newShaderEngine()
	}
	app.shader.ensure(w, h)
	c := app.shader.render(mode, bars)
	if c == nil {
		return false
	}
	blitCanvas(hdc, bounds, c)
	return true
}

func blitCanvas(hdc uintptr, bounds rect, c *visual.Canvas) {
	if c == nil || len(c.Pix) == 0 {
		return
	}
	// BGRA bottom-up DIB for SetDIBitsToDevice
	bgra := make([]byte, c.W*c.H*4)
	for y := 0; y < c.H; y++ {
		srcY := y
		dstRow := (c.H - 1 - y) * c.W * 4
		for x := 0; x < c.W; x++ {
			p := c.Pix[srcY*c.W+x]
			off := dstRow + x*4
			bgra[off] = byte(p)
			bgra[off+1] = byte(p >> 8)
			bgra[off+2] = byte(p >> 16)
			bgra[off+3] = 255
		}
	}
	var bih bitmapInfoHeader
	bih.size = uint32(unsafe.Sizeof(bih))
	bih.width = int32(c.W)
	bih.height = int32(c.H)
	bih.planes = 1
	bih.bitCount = 32
	bih.compression = biRGB
	procSetDIBitsToDevice.Call(
		hdc,
		uintptr(bounds.left), uintptr(bounds.top),
		uintptr(c.W), uintptr(c.H),
		0, 0, 0, uintptr(c.H),
		uintptr(unsafe.Pointer(&bgra[0])),
		uintptr(unsafe.Pointer(&bih)),
		dibRGBColors,
	)
}

func isShaderMode(mode visualMode) bool {
	switch mode {
	case modeFluid, modeGalaxy, modeChromatic:
		return true
	default:
		return false
	}
}

func captureCanvasFrame(c *visual.Canvas) *visual.Canvas {
	if c == nil {
		return nil
	}
	dup := visual.NewCanvas(c.W, c.H)
	copy(dup.Pix, c.Pix)
	return dup
}

func shaderPreviewColor(mode visualMode) uintptr {
	switch mode {
	case modeFluid:
		return rgb(40, 120, 200)
	case modeGalaxy:
		return rgb(80, 60, 200)
	case modeChromatic:
		return rgb(200, 60, 140)
	case modeClassic:
		return rgb(60, 180, 120)
	case modeAurora:
		return rgb(50, 140, 200)
	case modeNeonCity:
		return rgb(200, 80, 160)
	case modeSupernova:
		return rgb(220, 140, 60)
	case modeLiquid:
		return rgb(60, 160, 200)
	default:
		return rgb(60, 60, 80)
	}
}

func formatShaderStatus(mode visualMode) string {
	switch mode {
	case modeFluid:
		return "GPU Fluid"
	case modeGalaxy:
		return "GPU Galaxy"
	case modeChromatic:
		return "GPU Chromatic"
	default:
		return ""
	}
}

