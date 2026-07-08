//go:build windows

package main

import (
	"time"
)

type perfEngine struct {
	lastTick   time.Time
	avgFrameMs float64
	quality    float64 // 1.0 full, 0.5 reduced
	skipTrails bool
	halfCanvas bool
}

func newPerfEngine() *perfEngine {
	return &perfEngine{quality: 1.0, lastTick: time.Now()}
}

func (p *perfEngine) tickFrame() {
	now := time.Now()
	dt := now.Sub(p.lastTick).Seconds()
	p.lastTick = now
	if dt <= 0 || dt > 0.5 {
		return
	}
	ms := dt * 1000
	if p.avgFrameMs == 0 {
		p.avgFrameMs = ms
	} else {
		p.avgFrameMs = p.avgFrameMs*0.9 + ms*0.1
	}
	target := 1000.0 / float64(fps)
	if p.avgFrameMs > target*1.35 {
		p.quality = clamp(p.quality-0.05, 0.45, 1.0)
		p.skipTrails = p.quality < 0.85
		p.halfCanvas = p.quality < 0.65
	} else if p.avgFrameMs < target*1.05 && p.quality < 1.0 {
		p.quality = clamp(p.quality+0.02, 0.45, 1.0)
		p.skipTrails = p.quality < 0.85
		p.halfCanvas = p.quality < 0.65
	}
}

func (p *perfEngine) canvasScale() float64 {
	if p.halfCanvas {
		return 0.75
	}
	if p.quality < 0.85 {
		return 0.9
	}
	return 1.0
}

func (p *perfEngine) statusText() string {
	if p.quality >= 0.95 {
		return ""
	}
	return "Adaptive quality " + formatFloat(p.quality*100) + "%"
}
