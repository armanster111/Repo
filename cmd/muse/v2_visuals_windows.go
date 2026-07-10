//go:build windows

package main

import (
	"math"
	"time"
)

func drawAurora(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	rows := 14
	cols := len(bars)
	cellW := max(int32(3), (bounds.right-bounds.left)/int32(cols))
	cellH := max(int32(3), (bounds.bottom-bounds.top)/int32(rows))
	now := float64(time.Now().UnixMilli()) / 900
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			energy := bars[x]
			wave := math.Sin(float64(x)*0.35+now+float64(y)*0.2) * 0.5
			level := clamp((energy*0.75+wave+0.5)*0.65, 0, 1)
			left := bounds.left + int32(x)*cellW
			top := bounds.top + int32(y)*cellH
			fill(hdc, rect{left: left, top: top, right: left + cellW - 1, bottom: top + cellH - 1},
				rgb(byte(20+level*40), byte(80+level*140), byte(120+level*120)))
		}
	}
}

func drawMandala(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	maxR := min(bounds.right-bounds.left, bounds.bottom-bounds.top) / 2
	now := float64(time.Now().UnixMilli()) / 1200
	for ring := 1; ring <= 8; ring++ {
		radius := int32(float64(maxR) * float64(ring) / 9)
		for i, energy := range bars {
			if i%8 != ring%8 {
				continue
			}
			angle := float64(i)/float64(len(bars))*math.Pi*2 + now
			x := cx + int32(math.Cos(angle)*float64(radius)*(0.7+energy*0.35))
			y := cy + int32(math.Sin(angle)*float64(radius)*(0.7+energy*0.35))
			size := int32(3 + energy*8)
			ellipse(hdc, x-size, y-size, x+size, y+size, gradientColor(i, len(bars), 255))
		}
	}
}

func drawStarfield(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	now := float64(time.Now().UnixMilli()) / 500
	for i, energy := range bars {
		angle := float64(i)/float64(len(bars))*math.Pi*2 + now*0.05
		layer := float64(i%5) / 5
		radius := (0.12 + layer*0.75 + energy*0.2) * float64(min(bounds.right-bounds.left, bounds.bottom-bounds.top)) / 2
		x := cx + int32(math.Cos(angle)*radius)
		y := cy + int32(math.Sin(angle)*radius)
		size := int32(2 + energy*6)
		ellipse(hdc, x-size, y-size, x+size, y+size, rgb(byte(180+energy*70), byte(200+energy*40), 255))
	}
}

func drawKaleidoscope(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	segments := 8
	now := float64(time.Now().UnixMilli()) / 700
	for seg := 0; seg < segments; seg++ {
		base := float64(seg) / float64(segments) * math.Pi * 2
		for i, energy := range bars {
			if i%6 != seg%6 {
				continue
			}
			angle := base + float64(i)/float64(len(bars))*math.Pi/float64(segments) + now*0.1
			radius := (0.15 + energy*0.8) * float64(min(bounds.right-bounds.left, bounds.bottom-bounds.top)) / 2
			x1 := cx + int32(math.Cos(angle)*radius*0.4)
			y1 := cy + int32(math.Sin(angle)*radius*0.4)
			x2 := cx + int32(math.Cos(angle)*radius)
			y2 := cy + int32(math.Sin(angle)*radius)
			line(hdc, x1, y1, x2, y2, gradientColor(i, len(bars), 255), 2)
		}
	}
}
