//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/armanster111/music-visualizer/internal/visual"
)

func (s *appState) toggleMp4Record() {
	if s.mp4Recording {
		s.finishVideoRecord()
		return
	}
	s.mp4Recording = true
	s.mp4FrameCount = 0
	s.videoFrames = s.videoFrames[:0]
	s.flashStatus("Recording video... press * to stop.", 3*time.Second)
}

func (s *appState) captureMp4Frame() {
	if !s.mp4Recording {
		return
	}
	frame := s.renderExportFrame(640, 360)
	if frame == nil {
		return
	}
	dup := visual.NewCanvas(frame.W, frame.H)
	copy(dup.Pix, frame.Pix)
	s.videoFrames = append(s.videoFrames, dup)
	s.mp4FrameCount++
	if s.mp4FrameCount >= 600 {
		s.finishVideoRecord()
	}
}

func (s *appState) finishVideoRecord() {
	if !s.mp4Recording {
		return
	}
	s.mp4Recording = false
	count := len(s.videoFrames)
	frames := s.videoFrames
	s.videoFrames = nil
	s.mp4FrameCount = 0
	if count == 0 {
		s.flashStatus("Video recording cancelled.", 2*time.Second)
		return
	}
	home, _ := os.UserHomeDir()
	outDir := filepath.Join(home, "Desktop")
	if info, err := os.Stat(outDir); err != nil || !info.IsDir() {
		outDir = home
	}
	outPath := filepath.Join(outDir, fmt.Sprintf("music-visualizer-%s.avi", time.Now().Format("20060102-150405")))
	if err := visual.EncodeMJPEGAVI(outPath, frames, fps); err != nil {
		s.flashStatus("Video export failed: "+err.Error(), 3*time.Second)
		return
	}
	s.flashStatus(fmt.Sprintf("Saved video (%d frames): %s", count, outPath), 4*time.Second)
}

func (s *appState) renderExportFrame(w, h int) *visual.Canvas {
	bars := s.targetBars()
	if s.ultra != nil {
		beat := s.beatMultiplier(bars)
		bars = s.ultra.processBars(bars, s.visualIntensity, beat)
	} else {
		beat := s.beatMultiplier(bars)
		for i := range bars {
			bars[i] = min(1.0, bars[i]*beat*s.visualIntensity)
		}
	}
	if isShaderMode(s.mode) {
		if s.shader == nil {
			s.shader = newShaderEngine()
		}
		s.shader.ensure(w, h)
		return s.shader.render(s.mode, bars)
	}
	c := visual.NewCanvas(w, h)
	c.Clear(0x00000000)
	if len(bars) == 0 {
		return c
	}
	gap := 4
	barW := max(2, (w-gap*(len(bars)-1))/len(bars))
	for i, v := range bars {
		hh := int(float64(h) * clamp(v, 0, 1) * 0.9)
		x0 := i * (barW + gap)
		col := uint32(byte(80+i*3))<<16 | uint32(byte(120+i*2))<<8 | uint32(byte(200+i))
		c.FillRect(x0, h-hh, x0+barW, h, col)
	}
	return c
}
