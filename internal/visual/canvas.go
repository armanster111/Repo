package visual

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"math"
	"os"
	"time"
)

// Canvas is an RGBA framebuffer for multi-pass shader-style effects.
type Canvas struct {
	W, H int
	Pix  []uint32 // 0x00RRGGBB
}

// NewCanvas allocates a framebuffer.
func NewCanvas(w, h int) *Canvas {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return &Canvas{W: w, H: h, Pix: make([]uint32, w*h)}
}

func (c *Canvas) idx(x, y int) int {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return -1
	}
	return y*c.W + x
}

// Clear fills the canvas with a color.
func (c *Canvas) Clear(col uint32) {
	for i := range c.Pix {
		c.Pix[i] = col
	}
}

// Plot sets one pixel.
func (c *Canvas) Plot(x, y int, col uint32) {
	if i := c.idx(x, y); i >= 0 {
		c.Pix[i] = col
	}
}

// Add adds light to a pixel (additive blend).
func (c *Canvas) Add(x, y int, r, g, b byte) {
	i := c.idx(x, y)
	if i < 0 {
		return
	}
	p := c.Pix[i]
	cr := byte(p >> 16)
	cg := byte(p >> 8)
	cb := byte(p)
	c.Pix[i] = uint32(addByte(cr, r))<<16 | uint32(addByte(cg, g))<<8 | uint32(addByte(cb, b))
}

func addByte(a, b byte) byte {
	s := int(a) + int(b)
	if s > 255 {
		return 255
	}
	return byte(s)
}

// FillRect fills a rectangle.
func (c *Canvas) FillRect(x0, y0, x1, y1 int, col uint32) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c.Plot(x, y, col)
		}
	}
}

// ApplyBloom runs a cheap box-blur bloom pass.
func (c *Canvas) ApplyBloom(strength float64) {
	if strength <= 0 {
		return
	}
	tmp := make([]uint32, len(c.Pix))
	copy(tmp, c.Pix)
	radius := 2
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			var r, g, b, n float64
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					i := c.idx(x+dx, y+dy)
					if i < 0 {
						continue
					}
					p := tmp[i]
					r += float64(byte(p >> 16))
					g += float64(byte(p >> 8))
					b += float64(byte(p))
					n++
				}
			}
			if n == 0 {
				continue
			}
			i := c.idx(x, y)
			base := tmp[i]
			br := byte(math.Min(255, float64(byte(base>>16))+r/n*strength))
			bg := byte(math.Min(255, float64(byte(base>>8))+g/n*strength))
			bb := byte(math.Min(255, float64(byte(base))+b/n*strength))
			c.Pix[i] = uint32(br)<<16 | uint32(bg)<<8 | uint32(bb)
		}
	}
}

// ApplyChromaticShift offsets RGB channels for a shader aberration look.
func (c *Canvas) ApplyChromaticShift(shift int) {
	if shift <= 0 {
		return
	}
	tmp := make([]uint32, len(c.Pix))
	copy(tmp, c.Pix)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			i := c.idx(x, y)
			_, g, b := rgbParts(tmp[i])
			ri := c.idx(x-shift, y)
			r := byte(0)
			if ri >= 0 {
				r, _, _ = rgbParts(tmp[ri])
			}
			c.Pix[i] = rgb(r, g, b)
		}
	}
}

func rgbParts(col uint32) (byte, byte, byte) {
	return byte(col >> 16), byte(col >> 8), byte(col)
}

func rgb(r, g, b byte) uint32 {
	return uint32(r)<<16 | uint32(g)<<8 | uint32(b)
}

// EncodeGIF writes animated GIF from canvas frames.
func EncodeGIF(path string, frames []*Canvas, frameDelay time.Duration) error {
	if len(frames) == 0 {
		return os.ErrInvalid
	}
	outGif := &gif.GIF{LoopCount: 0}
	ms := int(frameDelay.Milliseconds() / 10)
	if ms < 1 {
		ms = 1
	}
	for _, f := range frames {
		paletted := image.NewPaletted(image.Rect(0, 0, f.W, f.H), palette256())
		for y := 0; y < f.H; y++ {
			for x := 0; x < f.W; x++ {
				p := f.Pix[y*f.W+x]
				paletted.SetColorIndex(x, y, nearestPaletteIndex(p))
			}
		}
		outGif.Image = append(outGif.Image, paletted)
		outGif.Delay = append(outGif.Delay, ms)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, outGif)
}

// WritePNG saves the canvas as a PNG image.
func (c *Canvas) WritePNG(path string) error {
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

func palette256() color.Palette {
	p := make(color.Palette, 256)
	for i := 0; i < 256; i++ {
		v := uint8(i)
		p[i] = color.RGBA{R: v, G: v, B: v, A: 255}
	}
	return p
}

func nearestPaletteIndex(col uint32) uint8 {
	r, g, b := rgbParts(col)
	v := byte((int(r) + int(g) + int(b)) / 3)
	return v
}
