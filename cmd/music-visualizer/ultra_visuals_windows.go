//go:build windows

package main

import (
	"math"
	"time"
)

func drawNeonCity(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	width := bounds.right - bounds.left
	height := bounds.bottom - bounds.top
	skyline := bounds.bottom - height/4
	gap := int32(4)
	barW := max(int32(4), (width-gap*int32(len(bars)-1))/int32(len(bars)))
	palette := currentPalette()
	fill(hdc, rect{left: bounds.left, top: bounds.top, right: bounds.right, bottom: skyline}, dimColor(palette.background, 0.6))
	for i, energy := range bars {
		buildingH := int32((0.2 + energy*0.75) * float64(skyline-bounds.top))
		left := bounds.left + int32(i)*(barW+gap)
		fill(hdc, rect{left: left, top: skyline - buildingH, right: left + barW, bottom: skyline}, gradientColor(i, len(bars), 255))
		windows := int(energy * 6)
		for w := 0; w < windows; w++ {
			wy := skyline - buildingH + int32(w+1)*buildingH/7
			fill(hdc, rect{left: left + 2, top: wy, right: left + barW - 2, bottom: wy + 3}, palette.accent2)
		}
	}
	// reflection
	for i, energy := range bars {
		buildingH := int32(energy * float64(height) * 0.12)
		left := bounds.left + int32(i)*(barW+gap)
		fill(hdc, rect{left: left, top: skyline + 4, right: left + barW, bottom: skyline + 4 + buildingH},
			dimColor(gradientColor(i, len(bars), 180), 0.45))
	}
}

func drawSupernova(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	maxR := float64(min(bounds.right-bounds.left, bounds.bottom-bounds.top)) / 2
	now := float64(time.Now().UnixMilli()) / 600
	var energy float64
	for _, b := range bars {
		energy += b
	}
	energy /= float64(len(bars))
	for ring := 1; ring <= 8; ring++ {
		r := maxR * (float64(ring) / 8) * (0.55 + energy*0.5)
		alpha := byte(255 - ring*18)
		for i, b := range bars {
			if i%5 != ring%5 {
				continue
			}
			angle := float64(i)/float64(len(bars))*math.Pi*2 + now + float64(ring)*0.08
			x := cx + int32(math.Cos(angle)*r*(0.85+b*0.3))
			y := cy + int32(math.Sin(angle)*r*(0.85+b*0.3))
			size := int32(2 + b*10 + float64(ring)*0.4)
			ellipse(hdc, x-size, y-size, x+size, y+size, gradientColor(i, len(bars), alpha))
		}
	}
	line(hdc, cx-int32(maxR*0.2), cy, cx+int32(maxR*0.2), cy, currentPalette().accent, 1)
	line(hdc, cx, cy-int32(maxR*0.2), cx, cy+int32(maxR*0.2), currentPalette().accent2, 1)
}

func drawLiquid(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cols := len(bars)
	rows := 12
	cellW := max(int32(3), (bounds.right-bounds.left)/int32(cols))
	cellH := max(int32(3), (bounds.bottom-bounds.top)/int32(rows))
	now := float64(time.Now().UnixMilli()) / 750
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			energy := bars[x]
			drift := math.Sin(float64(x)*0.4+now+float64(y)*0.25) * 0.5
			level := clamp((energy*0.7+drift+0.5)*0.7, 0, 1)
			if float64(y)/float64(rows) > level+0.15 {
				continue
			}
			left := bounds.left + int32(x)*cellW
			top := bounds.top + int32(y)*cellH
			fill(hdc, rect{left: left, top: top, right: left + cellW - 1, bottom: top + cellH - 1},
				gradientColor(x+y, cols+rows, byte(180+level*75)))
		}
	}
}

func drawOrbit(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	now := float64(time.Now().UnixMilli()) / 900
	for i, energy := range bars {
		angle := float64(i)/float64(len(bars))*math.Pi*2 + now
		tilt := math.Sin(now*0.7) * 0.35
		radius := (0.18 + energy*0.72) * float64(min(bounds.right-bounds.left, bounds.bottom-bounds.top)) / 2
		x := cx + int32(math.Cos(angle)*radius)
		y := cy + int32(math.Sin(angle)*radius*(0.65+tilt))
		z := energy
		size := int32(4 + z*14)
		ellipse(hdc, x-size, y-size, x+size, y+size, gradientColor(i, len(bars), byte(160+z*95)))
		ellipse(hdc, x-size/2, y+size/2, x+size/2, y+size, dimColor(gradientColor(i, len(bars), 255), 0.35))
	}
}

func drawWaveform3D(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	width := bounds.right - bounds.left
	depth := 10
	layerH := (bounds.bottom - bounds.top) / int32(depth+2)
	for layer := depth; layer >= 0; layer-- {
		offset := int32(layer * 3)
		fade := 0.35 + float64(layer)/float64(depth)*0.65
		top := bounds.top + int32(layer)*layerH/2
		bottom := top + layerH*2
		gap := int32(3)
		barW := max(int32(2), (width-gap*int32(len(bars)-1))/int32(len(bars)))
		for i, energy := range bars {
			phase := math.Sin(float64(i)*0.25+float64(layer)*0.4) * 0.15
			value := clamp(energy*(0.7+fade*0.3)+phase, 0, 1)
			h := int32(value * float64(bottom-top))
			left := bounds.left + int32(i)*(barW+gap) + offset
			fill(hdc, rect{left: left, top: bottom - h, right: left + barW, bottom: bottom},
				dimColor(gradientColor(i, len(bars), 255), fade))
		}
	}
}
