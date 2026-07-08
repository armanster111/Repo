package visual

import "math"

// DrawNebula renders volumetric-style nebula clouds driven by spectrum energy.
func DrawNebula(c *Canvas, bars []float64, t float64, kick float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	step := 2
	if c.W*c.H > 500000 {
		step = 3
	}
	for y := 0; y < c.H; y += step {
		for x := 0; x < c.W; x += step {
			nx := float64(x)/float64(c.W)*2 - 1
			ny := float64(y)/float64(c.H)*2 - 1
			v := fbm(nx*1.8+t*0.08, ny*1.8-t*0.06, 4)
			v += fbm(nx*3.2-t*0.12, ny*3.2+t*0.09, 3) * 0.5
			band := int((nx+1)*0.5*float64(len(bars)-1))
			if band < 0 {
				band = 0
			}
			if band >= len(bars) {
				band = len(bars) - 1
			}
			e := bars[band] * (1 + kick*0.8)
			intensity := v * (0.25 + e*1.4)
			if intensity < 0.04 {
				continue
			}
			hue := float64(band)/float64(len(bars)) + t*0.03 + v*0.15
			col := HSL(hue, 0.75+ e*0.2, 0.25+intensity*0.55)
			for dy := 0; dy < step; dy++ {
				for dx := 0; dx < step; dx++ {
					c.Plot(x+dx, y+dy, col)
				}
			}
		}
	}
}

// DrawSynesthesia renders flowing color ribbons tied to each frequency band.
func DrawSynesthesia(c *Canvas, bars []float64, t float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	c.Clear(0x00000410)
	rows := len(bars)
	rowH := iMax(2, c.H/rows)
	for i, e := range bars {
		y0 := i * rowH
		y1 := y0 + rowH
		if y1 > c.H {
			y1 = c.H
		}
		for x := 0; x < c.W; x++ {
			wave := math.Sin(float64(x)*0.018+t*2.4+float64(i)*0.35) * e * float64(rowH) * 0.45
			cy := (y0+y1)/2 + int(wave)
			for dy := -2; dy <= 2; dy++ {
				col := SpectrumColor(i, rows, e)
				c.Add(x, cy+dy, byte(col>>16), byte(col>>8), byte(col))
			}
		}
	}
}

// DrawFractal renders a Julia set with bass-driven zoom and rotation.
func DrawFractal(c *Canvas, bars []float64, t float64, kick float64, zoom *float64) {
	if c == nil {
		return
	}
	var bass float64
	third := len(bars) / 3
	if third < 1 {
		third = 1
	}
	for i := 0; i < third; i++ {
		bass += bars[i]
	}
	bass /= float64(third)
	if zoom != nil {
		*zoom = *zoom*0.96 + (0.8+bass*0.6+kick*0.5)*0.04
	}
	zScale := 1.8
	if zoom != nil {
		zScale = *zoom
	}
	cx := -0.745 + math.Sin(t*0.15)*0.08 + bass*0.12
	cy := 0.186 + math.Cos(t*0.12)*0.08 + kick*0.1
	step := 2
	maxIter := 40
	for py := 0; py < c.H; py += step {
		for px := 0; px < c.W; px += step {
			x0 := (float64(px)/float64(c.W)*2.5 - 1.25) * zScale
			y0 := (float64(py)/float64(c.H)*2.5 - 1.25) * zScale
			x, y := x0, y0
			iter := 0
			for x*x+y*y <= 4 && iter < maxIter {
				xt := x*x - y*y + cx
				y = 2*x*y + cy
				x = xt
				iter++
			}
			if iter == maxIter {
				continue
			}
			frac := float64(iter) / float64(maxIter)
			col := HSL(frac*0.65+t*0.02+bass*0.2, 0.85, 0.2+frac*0.55)
			for dy := 0; dy < step; dy++ {
				for dx := 0; dx < step; dx++ {
					c.Plot(px+dx, py+dy, col)
				}
			}
		}
	}
}

// DrawTerrain renders a scrolling 3D frequency landscape.
func DrawTerrain(c *Canvas, bars []float64, t float64, scroll *float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	c.Clear(0x00020812)
	if scroll != nil {
		*scroll += 0.6 + t*0.001
	}
	off := 0.0
	if scroll != nil {
		off = *scroll
	}
	horizon := c.H * 2 / 5
	for row := 0; row < c.H-horizon; row++ {
		depth := float64(row) / float64(c.H-horizon)
		scale := 0.3 + depth*1.8
		yScreen := horizon + row
		for x := 0; x < c.W; x++ {
			worldX := (float64(x)/float64(c.W)*2-1)*scale + off*0.02
			idx := int(math.Mod(worldX*float64(len(bars))+float64(len(bars)*2), float64(len(bars))))
			if idx < 0 {
				idx = 0
			}
			h := bars[idx] * (1-depth*0.5)
			peak := horizon + int(h*float64(c.H)*0.35*(1-depth*0.3))
			if yScreen >= peak {
				col := HSL(float64(idx)/float64(len(bars)), 0.7, 0.15+depth*0.35+h*0.3)
				c.Plot(x, yScreen, col)
			}
		}
	}
	// sky gradient
	for y := 0; y < horizon; y++ {
		fade := 1 - float64(y)/float64(horizon)
		for x := 0; x < c.W; x++ {
			c.Plot(x, y, rgb(byte(8+fade*20), byte(12+fade*30), byte(40+fade*80)))
		}
	}
}

// DrawHyperspace renders a warp-speed star tunnel.
func DrawHyperspace(c *Canvas, bars []float64, t float64, stars [][3]float64) [][3]float64 {
	if c == nil {
		return stars
	}
	c.Clear(0x00000208)
	if len(stars) < 180 {
		stars = make([][3]float64, 180)
		for i := range stars {
			stars[i] = [3]float64{
				(math.Sin(float64(i)*1.7)*0.5 + 0.5),
				(math.Cos(float64(i)*2.3)*0.5 + 0.5),
				float64(i) / 180,
			}
		}
	}
	var speed float64 = 0.018
	for _, b := range bars[:len(bars)/4] {
		speed += b * 0.004
	}
	cx, cy := float64(c.W)/2, float64(c.H)/2
	for i := range stars {
		s := &stars[i]
		s[2] -= speed
		if s[2] <= 0 {
			s[0] = math.Mod(float64(i)*0.37+t*0.1, 1)
			s[1] = math.Mod(float64(i)*0.53+t*0.07, 1)
			s[2] = 1
		}
		z := s[2]
		px := int(cx + (s[0]-0.5)*float64(c.W)/z)
		py := int(cy + (s[1]-0.5)*float64(c.H)/z)
		size := int((1 - z) * 6)
		if size < 1 {
			size = 1
		}
		band := i % len(bars)
		e := bars[band]
		col := SpectrumColor(band, len(bars), e*(1-z))
		for dy := -size; dy <= size; dy++ {
			for dx := -size; dx <= size; dx++ {
				c.Add(px+dx, py+dy, byte(col>>16), byte(col>>8), byte(col))
			}
		}
	}
	return stars
}

// DrawWaterfall renders a scrolling spectrogram history.
func DrawWaterfall(c *Canvas, bars []float64, history *[][]float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	*history = append(*history, append([]float64(nil), bars...))
	maxRows := c.H
	if len(*history) > maxRows {
		*history = (*history)[len(*history)-maxRows:]
	}
	c.Clear(0x00000408)
	colW := iMax(1, c.W/len(bars))
	for row, frame := range *history {
		y := c.H - len(*history) + row
		if y < 0 || y >= c.H {
			continue
		}
		for i, e := range frame {
			x0 := i * colW
			col := HSL(float64(i)/float64(len(frame)), 0.85, 0.12+e*0.65)
			c.FillRect(x0, y, x0+colW, y+1, col)
		}
	}
}

func fbm(x, y float64, octaves int) float64 {
	var sum, amp, freq float64 = 0, 0.5, 1
	for o := 0; o < octaves; o++ {
		sum += amp * noise2(x*freq, y*freq)
		amp *= 0.5
		freq *= 2.1
	}
	return sum
}

func noise2(x, y float64) float64 {
	return math.Sin(x*1.7+y*0.9)*0.5 + math.Sin(x*0.6-y*1.3)*0.3 + math.Cos(x*1.1+y*1.1)*0.2
}

// DrawAuroraStorm renders flowing northern-lights curtains with bass-driven shimmer.
func DrawAuroraStorm(c *Canvas, bars []float64, t float64, kick float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	step := 2
	if c.W*c.H > 500000 {
		step = 3
	}
	var bass float64
	third := len(bars) / 3
	if third < 1 {
		third = 1
	}
	for i := 0; i < third; i++ {
		bass += bars[i]
	}
	bass /= float64(third)
	for y := 0; y < c.H; y += step {
		for x := 0; x < c.W; x += step {
			nx := float64(x) / float64(c.W)
			ny := float64(y) / float64(c.H)
			wave := math.Sin(nx*8+t*1.2)*0.15 + math.Sin(nx*14-t*0.8)*0.08
			curtain := math.Exp(-math.Pow(ny-0.35-wave, 2)*12) * (0.4 + bass*0.8 + kick*0.6)
			noise := fbm(nx*2+t*0.05, ny*3-t*0.04, 3)
			intensity := curtain * (0.5 + noise*0.5)
			if intensity < 0.03 {
				continue
			}
			band := int(nx * float64(len(bars)-1))
			if band < 0 {
				band = 0
			}
			if band >= len(bars) {
				band = len(bars) - 1
			}
			e := bars[band]
			hue := 0.35 + nx*0.25 + t*0.02 + e*0.15
			col := HSL(hue, 0.7+e*0.25, 0.2+intensity*0.5)
			for dy := 0; dy < step; dy++ {
				for dx := 0; dx < step; dx++ {
					c.Plot(x+dx, y+dy, col)
				}
			}
		}
	}
}

// DrawPulseGrid renders a perspective cyberpunk grid pulsing with the beat.
func DrawPulseGrid(c *Canvas, bars []float64, t float64, kick float64) {
	if c == nil || len(bars) == 0 {
		return
	}
	c.Clear(0x00000408)
	cx := float64(c.W) / 2
	horizon := c.H * 2 / 5
	var bass float64
	quarter := len(bars) / 4
	if quarter < 1 {
		quarter = 1
	}
	for i := 0; i < quarter; i++ {
		bass += bars[i]
	}
	bass /= float64(quarter)
	pulse := 1 + kick*0.8 + bass*0.4
	for row := 0; row < c.H-horizon; row++ {
		depth := float64(row) / float64(c.H-horizon)
		y := horizon + row
		spacing := int(8 + depth*40*pulse)
		if spacing < 2 {
			spacing = 2
		}
		if row%spacing == 0 {
			bright := 0.2 + depth*0.5 + kick*0.4
			col := HSL(0.75+t*0.01, 0.9, bright)
			for x := 0; x < c.W; x++ {
				c.Plot(x, y, col)
			}
		}
	}
	// vertical perspective lines
	for i := -20; i <= 20; i++ {
		xBase := cx + float64(i)*18*pulse
		band := (i + 20) % len(bars)
		e := bars[band]
		for row := 0; row < c.H-horizon; row++ {
			depth := float64(row) / float64(c.H-horizon)
			x := int(xBase + float64(i)*depth*depth*80)
			y := horizon + row
			bright := 0.15 + depth*0.4 + e*0.35
			col := HSL(0.55+float64(band)/float64(len(bars))*0.3, 0.85, bright)
			c.Plot(x, y, col)
			c.Plot(x+1, y, col)
		}
	}
	// sky glow
	for y := 0; y < horizon; y++ {
		fade := 1 - float64(y)/float64(horizon)
		for x := 0; x < c.W; x++ {
			c.Plot(x, y, rgb(byte(4+fade*12), byte(8+fade*20), byte(20+fade*60)))
		}
	}
}
