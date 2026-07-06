package visual

import (
	"math"
	"time"
)

// FluidState holds a fluid simulation grid.
type FluidState struct {
	W, H  int
	VelX  []float64
	VelY  []float64
	Dens  []float64
}

// NewFluidState creates simulation buffers.
func NewFluidState(w, h int) *FluidState {
	n := w * h
	return &FluidState{
		W: w, H: h,
		VelX: make([]float64, n),
		VelY: make([]float64, n),
		Dens: make([]float64, n),
	}
}

// StepFluid advances fluid and renders to canvas.
func StepFluid(s *FluidState, c *Canvas, bars []float64, t float64) {
	if s == nil || c == nil {
		return
	}
	// Inject energy from spectrum along bottom
	for i, e := range bars {
		x := i * s.W / iMax(1, len(bars))
		for dx := 0; dx < s.W/len(bars)+1; dx++ {
			idx := (s.H-2)*s.W + clampi(x+dx, 0, s.W-1)
			s.Dens[idx] = math.Min(1, s.Dens[idx]+e*0.35)
			s.VelY[idx] -= e * 0.8
		}
	}
	// Advect + diffuse (simplified)
	next := make([]float64, len(s.Dens))
	copy(next, s.Dens)
	for y := 1; y < s.H-1; y++ {
		for x := 1; x < s.W-1; x++ {
			i := y*s.W + x
			up := s.Dens[i-s.W]
			down := s.Dens[i+s.W]
			left := s.Dens[i-1]
			right := s.Dens[i+1]
			next[i] = s.Dens[i]*0.92 + (up+down+left+right)*0.02
			next[i] += math.Sin(t*0.8+float64(x)*0.08)*0.01
			if next[i] < 0.001 {
				next[i] = 0
			}
		}
	}
	s.Dens = next
	c.Clear(0x00050812)
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			v := s.Dens[y*s.W+x]
			if v < 0.02 {
				continue
			}
			r := byte(20 + v*180)
			g := byte(40 + v*120)
			b := byte(100 + v*155)
			c.Plot(x, y, rgb(r, g, b))
		}
	}
}

// DrawGalaxy renders a pseudo-3D starfield with depth.
func DrawGalaxy(c *Canvas, bars []float64, t float64) {
	if c == nil {
		return
	}
	c.Clear(0x00020410)
	cx, cy := float64(c.W)/2, float64(c.H)/2
	stars := 220
	for i := 0; i < stars; i++ {
		angle := float64(i)/float64(stars)*math.Pi*2 + t*0.12
		band := bars[i%len(bars)]
		depth := 0.2 + float64(i%10)/10
		radius := (depth*0.45 + band*0.55) * float64(iMin(c.W, c.H)) * 0.48
		x := int(cx + math.Cos(angle)*radius)
		y := int(cy + math.Sin(angle)*radius*0.72)
		size := 1 + int(band*4*depth)
		col := rgb(byte(120+depth*100+band*80), byte(160+depth*60), byte(220+band*30))
		for dy := -size; dy <= size; dy++ {
			for dx := -size; dx <= size; dx++ {
				if dx*dx+dy*dy <= size*size {
					c.Plot(x+dx, y+dy, col)
				}
			}
		}
	}
}

// DrawChromaticBurst radial burst with spectrum-driven spikes.
func DrawChromaticBurst(c *Canvas, bars []float64, t float64) {
	if c == nil {
		return
	}
	c.Clear(0x00000000)
	cx, cy := float64(c.W)/2, float64(c.H)/2
	maxR := float64(iMin(c.W, c.H)) * 0.48
	for i, e := range bars {
		angle := float64(i)/float64(len(bars))*math.Pi*2 - math.Pi/2 + t*0.05
		r := maxR * (0.15 + e*0.85)
		x0 := int(cx)
		y0 := int(cy)
		x1 := int(cx + math.Cos(angle)*r)
		y1 := int(cy + math.Sin(angle)*r)
		drawLineCanvas(c, x0, y0, x1, y1, rgb(byte(80+e*175), byte(40+e*80), byte(120+e*135)))
	}
}

func drawLineCanvas(c *Canvas, x0, y0, x1, y1 int, col uint32) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		c.Plot(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func iMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func iMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Now returns time for shaders.
func Now() float64 {
	return float64(time.Now().UnixMilli()) / 1000
}
