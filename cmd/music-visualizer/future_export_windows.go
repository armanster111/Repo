//go:build windows

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/armanster111/music-visualizer/internal/visual"
)

func (s *appState) toggleMp4Record() {
	if s.mp4Recording {
		s.finishMp4Record()
		return
	}
	dir, err := os.MkdirTemp("", "mvp-mp4-*")
	if err != nil {
		s.flashStatus("MP4 temp folder failed.", 2*time.Second)
		return
	}
	s.mp4Recording = true
	s.mp4FrameDir = dir
	s.mp4FrameCount = 0
	s.flashStatus("Recording MP4 frames... press * to stop.", 3*time.Second)
}

func (s *appState) captureMp4Frame() {
	if !s.mp4Recording || s.mp4FrameDir == "" {
		return
	}
	frame := s.renderExportFrame(640, 360)
	if frame == nil {
		return
	}
	name := filepath.Join(s.mp4FrameDir, fmt.Sprintf("frame_%05d.png", s.mp4FrameCount))
	if err := writeCanvasPNG(name, frame); err != nil {
		return
	}
	s.mp4FrameCount++
	if s.mp4FrameCount >= 600 {
		s.finishMp4Record()
	}
}

func (s *appState) finishMp4Record() {
	if !s.mp4Recording {
		return
	}
	s.mp4Recording = false
	count := s.mp4FrameCount
	dir := s.mp4FrameDir
	s.mp4FrameDir = ""
	s.mp4FrameCount = 0
	if count == 0 {
		_ = os.RemoveAll(dir)
		s.flashStatus("MP4 recording cancelled.", 2*time.Second)
		return
	}
	home, _ := os.UserHomeDir()
	outDir := filepath.Join(home, "Desktop")
	if info, err := os.Stat(outDir); err != nil || !info.IsDir() {
		outDir = home
	}
	outPath := filepath.Join(outDir, fmt.Sprintf("music-visualizer-%s.mp4", time.Now().Format("20060102-150405")))
	pattern := filepath.Join(dir, "frame_%05d.png")
	cmd := exec.Command("ffmpeg",
		"-y", "-hide_banner", "-loglevel", "error",
		"-framerate", fmt.Sprintf("%d", fps),
		"-i", pattern,
		"-c:v", "libx264", "-pix_fmt", "yuv420p",
		outPath,
	)
	if err := cmd.Run(); err != nil {
		fallback := filepath.Join(outDir, fmt.Sprintf("music-visualizer-frames-%s", time.Now().Format("20060102-150405")))
		_ = os.Rename(dir, fallback)
		s.flashStatus(fmt.Sprintf("ffmpeg not found — saved %d PNG frames to %s", count, fallback), 4*time.Second)
		return
	}
	_ = os.RemoveAll(dir)
	s.flashStatus(fmt.Sprintf("Saved MP4 (%d frames): %s", count, outPath), 4*time.Second)
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
	// Raster fallback: draw bars into canvas for export
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

func writeCanvasPNG(path string, c *visual.Canvas) error {
	img := image.NewRGBA(image.Rect(0, 0, c.W, c.H))
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			p := c.Pix[y*c.W+x]
			img.SetRGBA(x, y, color.RGBA{
				R: byte(p >> 16),
				G: byte(p >> 8),
				B: byte(p),
				A: 255,
			})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
