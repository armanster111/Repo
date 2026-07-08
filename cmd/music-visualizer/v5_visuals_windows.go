//go:build windows

package main

import (
	"math"
	"time"

	"github.com/armanster111/music-visualizer/internal/visual"
)

func drawSpectrogram(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	maxR := float64(min(bounds.right-bounds.left, bounds.bottom-bounds.top)) / 2
	palette := currentPalette()
	history := app.specHistory
	if len(history) == 0 {
		history = [][]float64{bars}
	}
	for ring, frame := range history {
		age := float64(len(history) - ring)
		alpha := age / float64(len(history)+1)
		for i, energy := range frame {
			if i >= len(bars) {
				break
			}
			angle := float64(i)/float64(len(frame))*2*math.Pi - math.Pi/2
			r := maxR * (0.15 + alpha*0.85) * (0.3 + energy*0.7)
			x := cx + int32(math.Cos(angle)*r)
			y := cy + int32(math.Sin(angle)*r)
			size := int32(2 + energy*6)
			col := dimColor(gradientColor(i, len(frame), 220), 0.3+alpha*0.7)
			fill(hdc, rect{left: x - size, top: y - size, right: x + size, bottom: y + size}, col)
		}
	}
	// current frame ring
	for i, energy := range bars {
		angle := float64(i)/float64(len(bars))*2*math.Pi - math.Pi/2 + float64(time.Now().UnixMilli())/2000
		r := maxR * (0.25 + energy*0.75)
		x := cx + int32(math.Cos(angle)*r)
		y := cy + int32(math.Sin(angle)*r)
		w := int32(3 + energy*8)
		fill(hdc, rect{left: x - w, top: y - 2, right: x + w, bottom: y + 2}, palette.accent)
	}
}

func drawOscilloscope(hdc uintptr, bounds rect, bars []float64) {
	width := bounds.right - bounds.left
	height := bounds.bottom - bounds.top
	midY := bounds.top + height/2
	palette := currentPalette()

	var waveform []float64
	if app.pcmCache != nil && len(app.pcmCache.Mono) > 0 {
		idx := app.pcmCache.SampleIndex(app.position().Seconds())
		waveform = visual.WaveformSlice(app.pcmCache.Mono, idx, int(width))
	} else {
		waveform = make([]float64, width)
		for i := range waveform {
			t := float64(i) / float64(width)
			waveform[i] = bars[int(t*float64(len(bars)-1))] * math.Sin(t*12*math.Pi)
		}
	}
	if len(waveform) < 2 {
		return
	}

	// Direct2D anti-aliased path when available
	if d2dReady() {
		pts := make([]d2dPoint, len(waveform))
		for i, v := range waveform {
			x := bounds.left + int32(i)
			y := midY - int32(v*float64(height)*0.42)
			pts[i] = d2dPoint{x: float32(x), y: float32(y)}
		}
		d2dDrawPolyline(hdc, pts, palette.accent, 2)
		return
	}

	pen, _, _ := procCreatePen.Call(0, 2, palette.accent)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	for i, v := range waveform {
		x := bounds.left + int32(i)
		y := midY - int32(v*float64(height)*0.42)
		if i == 0 {
			procMoveToEx.Call(hdc, uintptr(x), uintptr(y), 0)
		} else {
			procLineTo.Call(hdc, uintptr(x), uintptr(y))
		}
	}
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(pen)
}

func drawLissajous(hdc uintptr, bounds rect, bars []float64) {
	width := bounds.right - bounds.left
	height := bounds.bottom - bounds.top
	cx := bounds.left + width/2
	cy := bounds.top + height/2
	scale := float64(min(width, height)) * 0.38
	palette := currentPalette()

	var points []visual.StereoSample
	if app.pcmCache != nil {
		idx := app.pcmCache.SampleIndex(app.position().Seconds())
		left, right := app.pcmCache.StereoWindowAt(idx, 512)
		points = visual.LissajousPoints(left, right, 0, 256)
	} else if app.desktop != nil {
		left, right := app.desktop.stereoSnapshot()
		points = visual.LissajousPoints(left, right, max(0, len(left)-256), 256)
	}
	if len(points) < 2 {
		// fallback: use bars as fake L/R
		for i, b := range bars {
			if i >= 128 {
				break
			}
			angle := float64(i) / float64(len(bars)) * 2 * math.Pi
			points = append(points, visual.StereoSample{
				Left:  b * math.Cos(angle),
				Right: b * math.Sin(angle),
			})
		}
	}

	if d2dReady() {
		pts := make([]d2dPoint, len(points))
		for i, p := range points {
			pts[i] = d2dPoint{
				x: float32(cx) + float32(p.Left*scale),
				y: float32(cy) + float32(p.Right*scale),
			}
		}
		d2dDrawPolyline(hdc, pts, palette.accent2, 1.5)
		return
	}

	for i, p := range points {
		x := cx + int32(p.Left*scale)
		y := cy + int32(p.Right*scale)
		size := int32(2)
		col := dimColor(palette.accent2, 0.4+float64(i)/float64(len(points))*0.6)
		fill(hdc, rect{left: x - size, top: y - size, right: x + size, bottom: y + size}, col)
	}
}
