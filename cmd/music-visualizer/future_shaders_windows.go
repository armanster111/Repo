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
	canvas        *visual.Canvas
	fade          *visual.Canvas
	fluid         *visual.FluidState
	fluidT        float64
	waterfall     [][]float64
	fractalZoom   float64
	terrainScroll float64
	stars         [][3]float64
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
	scale := 1.0
	if app.perf != nil {
		scale = app.perf.canvasScale()
	}
	w = int(float64(w) * scale)
	h = int(float64(h) * scale)
	if e.canvas == nil || e.canvas.W != w || e.canvas.H != h {
		e.canvas = visual.NewCanvas(w, h)
		e.fade = visual.NewCanvas(w, h)
		e.fluid = visual.NewFluidState(w, h)
		e.waterfall = nil
		e.stars = nil
		e.fractalZoom = 1.0
		e.terrainScroll = 0
	}
}

func (e *shaderEngine) render(mode visualMode, bars []float64) *visual.Canvas {
	if e.canvas == nil {
		return nil
	}
	t := visual.Now()
	e.fluidT = t
	kick := 0.0
	if app.onset != nil {
		kick = app.onset.Kick
	}
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
	case modeVortex:
		visual.DrawVortex(e.canvas, bars, t, kick)
		e.canvas.ApplyBloom(0.65)
	case modeComets:
		visual.DrawComets(e.canvas, bars, t)
		e.canvas.ApplyBloom(0.5)
	case modeBassDrop:
		visual.DrawBassDrop(e.canvas, bars, t, kick)
		e.canvas.ApplyBloom(0.75)
	case modeRain:
		visual.DrawRain(e.canvas, bars, t)
		e.canvas.ApplyBloom(0.4)
	case modePrism:
		visual.DrawPrism(e.canvas, bars, t)
		e.canvas.ApplyBloom(0.6)
		e.canvas.ApplyChromaticShift(1)
	case modeNebula:
		e.canvas.FadeFromPrevious(e.fade, 0.88)
		visual.DrawNebula(e.canvas, bars, t, kick)
		e.canvas.ApplyPremiumBloom(0.75)
		e.canvas.ApplyVignette(0.45)
	case modeSynesthesia:
		e.canvas.FadeFromPrevious(e.fade, 0.82)
		visual.DrawSynesthesia(e.canvas, bars, t)
		e.canvas.ApplyPremiumBloom(0.65)
		e.canvas.ApplyVignette(0.35)
	case modeFractal:
		visual.DrawFractal(e.canvas, bars, t, kick, &e.fractalZoom)
		e.canvas.ApplyPremiumBloom(0.7)
		e.canvas.ApplyVignette(0.4)
	case modeTerrain:
		visual.DrawTerrain(e.canvas, bars, t, &e.terrainScroll)
		e.canvas.ApplyBloom(0.55)
		e.canvas.ApplyVignette(0.35)
	case modeHyperspace:
		e.stars = visual.DrawHyperspace(e.canvas, bars, t, e.stars)
		e.canvas.ApplyPremiumBloom(0.8)
		e.canvas.ApplyVignette(0.5)
	case modeWaterfall:
		visual.DrawWaterfall(e.canvas, bars, &e.waterfall)
		e.canvas.ApplyBloom(0.35)
	case modeAuroraStorm:
		e.canvas.FadeFromPrevious(e.fade, 0.9)
		visual.DrawAuroraStorm(e.canvas, bars, t, kick)
		e.canvas.ApplyPremiumBloom(0.85)
		e.canvas.ApplyVignette(0.4)
	case modePulseGrid:
		visual.DrawPulseGrid(e.canvas, bars, t, kick)
		e.canvas.ApplyPremiumBloom(0.7)
		e.canvas.ApplyChromaticShift(1)
	default:
		return nil
	}
	if e.fade != nil && usesFadePersistence(mode) {
		e.fade.CopyFrom(e.canvas)
	}
	return e.canvas
}

func usesFadePersistence(mode visualMode) bool {
	switch mode {
	case modeNebula, modeSynesthesia, modeAuroraStorm:
		return true
	default:
		return false
	}
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
	case modeFluid, modeGalaxy, modeChromatic,
		modeVortex, modeComets, modeBassDrop, modeRain, modePrism,
		modeNebula, modeSynesthesia, modeFractal, modeTerrain,
		modeHyperspace, modeWaterfall, modeAuroraStorm, modePulseGrid:
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
	case modeVortex:
		return rgb(60, 140, 220)
	case modeComets:
		return rgb(100, 80, 240)
	case modeBassDrop:
		return rgb(200, 60, 180)
	case modeRain:
		return rgb(40, 180, 160)
	case modePrism:
		return rgb(220, 120, 80)
	case modeNebula:
		return rgb(80, 40, 180)
	case modeSynesthesia:
		return rgb(200, 80, 160)
	case modeFractal:
		return rgb(60, 180, 220)
	case modeTerrain:
		return rgb(40, 120, 80)
	case modeHyperspace:
		return rgb(100, 60, 240)
	case modeWaterfall:
		return rgb(40, 160, 200)
	case modeAuroraStorm:
		return rgb(40, 200, 120)
	case modePulseGrid:
		return rgb(200, 60, 220)
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
	case modeVortex:
		return "GPU Vortex"
	case modeComets:
		return "GPU Comets"
	case modeBassDrop:
		return "GPU Bass Drop"
	case modeRain:
		return "GPU Rain"
	case modePrism:
		return "GPU Prism"
	case modeNebula:
		return "GPU Nebula"
	case modeSynesthesia:
		return "GPU Synesthesia"
	case modeFractal:
		return "GPU Fractal"
	case modeTerrain:
		return "GPU Terrain"
	case modeHyperspace:
		return "GPU Hyperspace"
	case modeWaterfall:
		return "GPU Waterfall"
	case modeAuroraStorm:
		return "GPU Aurora Storm"
	case modePulseGrid:
		return "GPU Pulse Grid"
	default:
		return ""
	}
}
