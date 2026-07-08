package visual

import "math"

// DrawVortex renders a spiraling energy tunnel driven by bass.
func DrawVortex(c *Canvas, bars []float64, t float64, kick float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	c.Clear(0x00020814)
	cx, cy := float64(c.W)/2, float64(c.H)/2
	arms := 3
	for arm := 0; arm < arms; arm++ {
		offset := float64(arm) * math.Pi * 2 / float64(arms)
		for i, e := range bars {
			angle := float64(i)/float64(len(bars))*math.Pi*6 + t*1.4 + offset - kick*0.8
			r := (float64(i)/float64(len(bars)))*float64(iMin(c.W, c.H))*0.46*(0.4 + e*0.9 + kick*0.3)
			x := int(cx + math.Cos(angle)*r)
			y := int(cy + math.Sin(angle)*r)
			size := 1 + int(e*5+kick*4)
			col := rgb(byte(30+e*200), byte(80+e*120+kick*80), byte(180+e*75))
			for dy := -size; dy <= size; dy++ {
				for dx := -size; dx <= size; dx++ {
					if dx*dx+dy*dy <= size*size {
						c.Add(x+dx, y+dy, byte(col>>16), byte(col>>8), byte(col))
					}
				}
			}
		}
	}
}

// DrawComets renders streaking particles across the canvas.
func DrawComets(c *Canvas, bars []float64, t float64) {
	if c == nil {
		return
	}
	c.Clear(0x00000410)
	for i, e := range bars {
		if e < 0.08 {
			continue
		}
		seed := float64(i)*1.73 + t*0.6
		x0 := int(float64(c.W) * math.Mod(seed*0.17, 1))
		y0 := int(float64(c.H) * math.Mod(seed*0.31, 1))
		length := int(e * float64(c.W) * 0.35)
		angle := seed * 2.1
		x1 := x0 + int(math.Cos(angle)*float64(length))
		y1 := y0 + int(math.Sin(angle)*float64(length))
		col := rgb(byte(60+e*195), byte(100+e*100), byte(200+e*55))
		drawLineCanvas(c, x0, y0, x1, y1, col)
		c.Add(x1, y1, byte(col>>16), byte(col>>8), byte(col))
	}
}

// DrawBassDrop renders shockwave rings from kick hits.
func DrawBassDrop(c *Canvas, bars []float64, t float64, kick float64) {
	if c == nil {
		return
	}
	c.Clear(0x00000612)
	cx, cy := float64(c.W)/2, float64(c.H)/2
	maxR := float64(iMin(c.W, c.H)) * 0.5
	var bass float64
	for i := 0; i < len(bars)/3; i++ {
		bass += bars[i]
	}
	if len(bars) > 0 {
		bass /= float64(len(bars) / 3)
	}
	punch := math.Max(kick, bass*0.8)
	for ring := 0; ring < 8; ring++ {
		phase := math.Mod(t*0.8+float64(ring)*0.12, 1.0)
		r := maxR * phase * (0.5 + punch*0.5)
		alpha := 1 - phase
		if alpha < 0.05 {
			continue
		}
		steps := 64
		for s := 0; s < steps; s++ {
			angle := float64(s) / float64(steps) * math.Pi * 2
			x := int(cx + math.Cos(angle)*r)
			y := int(cy + math.Sin(angle)*r)
			c.Add(x, y, byte(40+punch*180*alpha), byte(20+punch*80*alpha), byte(100+punch*155*alpha))
		}
	}
	for i, e := range bars {
		angle := float64(i)/float64(len(bars))*math.Pi*2 - math.Pi/2
		r := maxR * 0.15 * (1 + punch*2) * e
		x := int(cx + math.Cos(angle)*r)
		y := int(cy + math.Sin(angle)*r)
		c.Add(x, y, 200, 80, 255)
	}
}

// DrawRain renders spectrum-driven vertical light columns.
func DrawRain(c *Canvas, bars []float64, t float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	c.Clear(0x00000408)
	colW := iMax(2, c.W/len(bars))
	for i, e := range bars {
		x := i * colW
		drops := int(e * 18)
		for d := 0; d < drops; d++ {
			phase := math.Mod(t*1.5+float64(i)*0.07+float64(d)*0.13, 1.0)
			y := int(phase * float64(c.H))
			h := int(8 + e*24)
			col := rgb(byte(40+e*180), byte(120+e*100), byte(180+e*75))
			c.FillRect(x, y, x+colW-1, y+h, col)
		}
	}
}

// DrawPrism renders refracted rainbow beams from a central prism.
func DrawPrism(c *Canvas, bars []float64, t float64) {
	if c == nil {
		return
	}
	c.Clear(0x00000208)
	cx, cy := c.W/2, c.H/2
	c.FillRect(cx-4, cy-int(float64(c.H)*0.2), cx+4, cy+int(float64(c.H)*0.2), 0x00CCCCCC)
	for i, e := range bars {
		if e < 0.06 {
			continue
		}
		angle := (float64(i)/float64(len(bars)) - 0.5) * math.Pi * 0.9
		length := int(float64(c.H) * 0.45 * e)
		hue := float64(i) / float64(len(bars))
		r := byte(80 + hue*175)
		g := byte(60 + (1-hue)*140)
		b := byte(100 + e*155)
		x1 := cx + int(math.Sin(angle+ t*0.05)*float64(length))
		y1 := cy - int(math.Cos(angle)*float64(length)*0.8)
		drawLineCanvas(c, cx, cy, x1, y1, rgb(r, g, b))
		c.Add(x1, y1, r, g, b)
	}
}
