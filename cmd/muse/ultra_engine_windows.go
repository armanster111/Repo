//go:build windows

package main

import (
	"math"
	"time"

	"github.com/armanster111/muse/internal/metadata"
	"github.com/armanster111/muse/internal/visual"
)

type ultraEngine struct {
	smoother   *visual.Smoother
	trails     *visual.TrailBuffer
	fluxHist   []float64
	prevBars   []float64
	showcase   bool
	showcaseAt time.Time
	artTint    palette
	artTintOn  bool
}

func newUltraEngine() *ultraEngine {
	return &ultraEngine{
		smoother:   visual.NewSmoother(barCount, 0.85, 0.28, 0.35),
		trails:     visual.NewTrailBuffer(6, barCount),
		showcaseAt: time.Now(),
	}
}

func (e *ultraEngine) processBars(target []float64, intensity, beat float64) []float64 {
	if len(target) == 0 {
		return target
	}
	dt := 1.0 / float64(fps)
	onset, hist := visual.OnsetBeat(e.prevBars, target, e.fluxHist)
	e.fluxHist = hist
	e.prevBars = append([]float64(nil), target...)

	combined := beat * onset
	out := make([]float64, len(target))
	for i, v := range target {
		out[i] = math.Min(1, v*combined*intensity)
	}
	smoothed := e.smoother.Step(out, dt)
	e.trails.Push(smoothed)
	return smoothed
}

func (e *ultraEngine) peakBars() []float64 {
	return e.smoother.Peaks()
}

func (e *ultraEngine) trailFrames() [][]float64 {
	return e.trails.Frames()
}

func (e *ultraEngine) tickShowcase() {
	if !e.showcase {
		return
	}
	if time.Since(e.showcaseAt) < 12*time.Second {
		return
	}
	e.showcaseAt = time.Now()
	app.prevMode = app.mode
	app.modeBlend = 0
	showcaseModes := []visualMode{
		modeKaleidoscope, modeSupernova, modeNeonCity, modeOrbit,
		modeLiquid, modeHalo, modeAurora, modeWaveform3D,
	}
	for i, m := range showcaseModes {
		if m == app.mode {
			app.mode = showcaseModes[(i+1)%len(showcaseModes)]
			app.updateStatus()
			return
		}
	}
	app.mode = showcaseModes[0]
	app.updateStatus()
}

func (e *ultraEngine) exitShowcase() {
	if !e.showcase {
		return
	}
	e.showcase = false
	app.visualOnly = false
	procShowWindow.Call(app.hwnd, swShowNormal)
	app.status = "Cinema showcase off."
	app.updateStatus()
	invalidate()
}

func (e *ultraEngine) toggleShowcase() {
	if e.showcase {
		e.exitShowcase()
		return
	}
	e.showcase = true
	app.visualOnly = true
	procShowWindow.Call(app.hwnd, swShowMaximized)
	app.mode = modeSupernova
	e.showcaseAt = time.Now()
	app.status = "Cinema showcase — auto-cycles premium visuals. Esc or F9 to exit."
	invalidate()
}

func (e *ultraEngine) syncArtTint(grid *metadata.ArtGrid) {
	if grid == nil || len(grid.Colors) == 0 {
		e.artTintOn = false
		return
	}
	var r, g, b int
	for _, c := range grid.Colors {
		r += int((c >> 16) & 0xff)
		g += int((c >> 8) & 0xff)
		b += int(c & 0xff)
	}
	n := float64(len(grid.Colors))
	r8 := byte(math.Min(255, float64(r)/n*0.35))
	g8 := byte(math.Min(255, float64(g)/n*0.35))
	b8 := byte(math.Min(255, float64(b)/n*0.35))
	e.artTint = palette{
		background: rgb(r8/3, g8/3, b8/3),
		panel:      rgb(r8/2, g8/2, b8/2),
		text:       rgb(245, 245, 255),
		dim:        rgb(r8+40, g8+40, b8+60),
		accent:     rgb(r8+120, g8+80, b8+140),
		accent2:    rgb(r8+60, g8+140, b8+120),
	}
	e.artTintOn = true
}

func drawWithEffects(hdc uintptr, bounds rect, bars []float64, peaks []float64, trails [][]float64, mode visualMode) {
	if isShaderMode(mode) {
		drawShaderMode(hdc, bounds, bars, mode)
		return
	}
	palette := currentPalette()
	for i, trail := range trails {
		if i == 0 {
			continue
		}
		fade := 0.12 + float64(len(trails)-i)*0.08
		faded := make([]float64, len(trail))
		for j, v := range trail {
			faded[j] = v * fade
		}
		drawVisualizationSoft(hdc, bounds, faded, mode, 0.35)
	}
	drawVisualizationSoft(hdc, bounds, bars, mode, 1)
	if len(peaks) == len(bars) {
		drawPeakGlow(hdc, bounds, peaks, palette)
	}
}

func drawVisualizationSoft(hdc uintptr, bounds rect, bars []float64, mode visualMode, alpha float64) {
	if alpha < 1 {
		// bloom pass: draw slightly expanded bounds
		expand := int32(4 * alpha)
		softBounds := rect{
			left: bounds.left - expand, top: bounds.top - expand,
			right: bounds.right + expand, bottom: bounds.bottom + expand,
		}
		drawVisualization(hdc, softBounds, bars, mode)
	}
	drawVisualization(hdc, bounds, bars, mode)
}

func drawPeakGlow(hdc uintptr, bounds rect, peaks []float64, palette palette) {
	if len(peaks) == 0 {
		return
	}
	width := bounds.right - bounds.left
	gap := int32(3)
	barW := max(int32(2), (width-gap*int32(len(peaks)-1))/int32(len(peaks)))
	for i, peak := range peaks {
		if peak < 0.15 {
			continue
		}
		h := int32(peak * float64(bounds.bottom-bounds.top) * 0.08)
		left := bounds.left + int32(i)*(barW+gap)
		fill(hdc, rect{left: left, top: bounds.bottom - h - 2, right: left + barW, bottom: bounds.bottom - 2},
			dimColor(palette.accent2, 0.5+peak*0.4))
	}
}
