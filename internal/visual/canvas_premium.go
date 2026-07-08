package visual

import "math"

// HSL converts hue [0,1], saturation [0,1], lightness [0,1] to RGB pixel.
func HSL(h, s, l float64) uint32 {
	h = math.Mod(h, 1)
	if h < 0 {
		h += 1
	}
	if s <= 0 {
		v := byte(l * 255)
		return rgb(v, v, v)
	}
	var r, g, b float64
	q := l * (1 + s)
	if l > 0.5 {
		q = l + s - l*s
	}
	p := 2*l - q
	r = hueToRGB(p, q, h+1.0/3)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3)
	return rgb(byte(r*255), byte(g*255), byte(b*255))
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 0.5:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	default:
		return p
	}
}

// SpectrumColor maps bar index and energy to a vivid HSL color.
func SpectrumColor(index, count int, energy float64) uint32 {
	hue := float64(index)/float64(iMax(1, count))
	sat := 0.65 + energy*0.35
	light := 0.35 + energy*0.45
	return HSL(hue, sat, light)
}

// FadeFromPrevious blends decayed previous frame for motion persistence.
func (c *Canvas) FadeFromPrevious(prev *Canvas, decay float64) {
	if prev == nil || prev.W != c.W || prev.H != c.H || decay <= 0 {
		return
	}
	for i := range c.Pix {
		p := prev.Pix[i]
		r := byte(float64(byte(p>>16)) * decay)
		g := byte(float64(byte(p>>8)) * decay)
		b := byte(float64(byte(p)) * decay)
		c.Pix[i] = rgb(r, g, b)
	}
}

// CopyFrom duplicates pixel buffer.
func (c *Canvas) CopyFrom(src *Canvas) {
	if src == nil || c == nil || len(c.Pix) != len(src.Pix) {
		return
	}
	copy(c.Pix, src.Pix)
}

// ApplyPremiumBloom runs a wider two-pass bloom for premium modes.
func (c *Canvas) ApplyPremiumBloom(strength float64) {
	c.ApplyBloom(strength * 0.6)
	c.ApplyBloom(strength * 0.45)
}

// ApplyVignette darkens edges for cinematic depth.
func (c *Canvas) ApplyVignette(strength float64) {
	if strength <= 0 {
		return
	}
	cx := float64(c.W) / 2
	cy := float64(c.H) / 2
	maxR := math.Hypot(cx, cy)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			i := c.idx(x, y)
			d := math.Hypot(float64(x)-cx, float64(y)-cy) / maxR
			fade := 1 - d*strength
			if fade < 0 {
				fade = 0
			}
			p := c.Pix[i]
			r := byte(float64(byte(p>>16)) * fade)
			g := byte(float64(byte(p>>8)) * fade)
			b := byte(float64(byte(p)) * fade)
			c.Pix[i] = rgb(r, g, b)
		}
	}
}
