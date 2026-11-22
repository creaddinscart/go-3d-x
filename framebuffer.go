package x3d

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
)

// Framebuffer is a color + depth buffer that a Renderer draws into. It can be
// exported as an image or streamed to a terminal as ANSI art.
type Framebuffer struct {
	W, H  int
	Pix   []Color
	Depth []float32
}

// NewFramebuffer allocates a cleared framebuffer.
func NewFramebuffer(w, h int) *Framebuffer {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	fb := &Framebuffer{W: w, H: h, Pix: make([]Color, w*h), Depth: make([]float32, w*h)}
	fb.Clear(Black, 1)
	return fb
}

// Clear fills color and depth buffers.
func (fb *Framebuffer) Clear(c Color, depth float32) {
	for i := range fb.Pix {
		fb.Pix[i] = c
		fb.Depth[i] = depth
	}
}

// SetPixel writes a color unconditionally.
func (fb *Framebuffer) SetPixel(x, y int, c Color) {
	if x >= 0 && x < fb.W && y >= 0 && y < fb.H {
		fb.Pix[y*fb.W+x] = c
	}
}

// BlendOrSet writes the pixel honoring alpha when the destination is empty.
func (fb *Framebuffer) BlendOrSet(x, y int, c Color) {
	if x < 0 || x >= fb.W || y < 0 || y >= fb.H {
		return
	}
	dst := fb.Pix[y*fb.W+x]
	if c.A >= 1 {
		fb.Pix[y*fb.W+x] = c
		return
	}
	a := c.A
	fb.Pix[y*fb.W+x] = Color{
		Lerp(dst.R, c.R, a), Lerp(dst.G, c.G, a),
		Lerp(dst.B, c.B, a), Clamp01(dst.A + a),
	}
}

// TryDepth writes the pixel if z passes the depth test (smaller is nearer).
func (fb *Framebuffer) TryDepth(x, y int, z float32, c Color) bool {
	if x < 0 || x >= fb.W || y < 0 || y >= fb.H {
		return false
	}
	idx := y*fb.W + x
	if z < fb.Depth[idx] {
		fb.Depth[idx] = z
		fb.Pix[idx] = c
		return true
	}
	return false
}

// Pixel reads a color.
func (fb *Framebuffer) Pixel(x, y int) Color { return fb.Pix[y*fb.W+x] }

// Fill paints the entire color buffer with c (depth is preserved).
func (fb *Framebuffer) Fill(c Color) {
	for i := range fb.Pix {
		fb.Pix[i] = c
	}
}

// ToImage converts the framebuffer to an image.NRGBA.
func (fb *Framebuffer) ToImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, fb.W, fb.H))
	for i := 0; i < fb.W*fb.H; i++ {
		img.SetNRGBA(i%fb.W, i/fb.W, fb.Pix[i].ToColorNRGBA())
	}
	return img
}

// SavePNG writes the framebuffer to a PNG file.
func (fb *Framebuffer) SavePNG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, fb.ToImage())
}

// WritePNG encodes the framebuffer as PNG to any writer.
func (fb *Framebuffer) WritePNG(w io.Writer) error { return png.Encode(w, fb.ToImage()) }

func quantize8(v float32) int {
	n := int(v*255 + 0.5)
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return n
}

// ToANSI renders the framebuffer as 24-bit truecolor half-block art. Each
// terminal cell packs two vertical pixels (background = top, foreground =
// bottom) using the "▀" glyph, so the terminal shows fb.H/2 rows of fb.W
// columns.
func (fb *Framebuffer) ToANSI() string {
	var b strings.Builder
	b.Grow(fb.W * fb.H * 4)
	for y := 0; y < fb.H; y += 2 {
		for x := 0; x < fb.W; x++ {
			top := fb.Pix[y*fb.W+x]
			bot := top
			if y+1 < fb.H {
				bot = fb.Pix[(y+1)*fb.W+x]
			}
			b.WriteString("\x1b[48;2;")
			b.WriteString(strconv.Itoa(quantize8(top.R)))
			b.WriteByte(';')
			b.WriteString(strconv.Itoa(quantize8(top.G)))
			b.WriteByte(';')
			b.WriteString(strconv.Itoa(quantize8(top.B)))
			b.WriteString("m\x1b[38;2;")
			b.WriteString(strconv.Itoa(quantize8(bot.R)))
			b.WriteByte(';')
			b.WriteString(strconv.Itoa(quantize8(bot.G)))
			b.WriteByte(';')
			b.WriteString(strconv.Itoa(quantize8(bot.B)))
			b.WriteString("m▀")
		}
		b.WriteString("\x1b[0m")
		if y+2 < fb.H {
			b.WriteString("\x1b[2K\r\n")
		}
	}
	return b.String()
}

// xterm256 maps a channel triple to the nearest xterm 256-color index.
func xterm256(c Color) int {
	r, g, b := c.R, c.G, c.B
	// grayscale cube at 232..255 for near-gray colors
	mx, mn := Max(Max(r, g), b), Min(Min(r, g), b)
	if mx-mn < 0.1 {
		v := int((r*24 + 0.5))
		if v < 0 {
			v = 0
		}
		if v > 23 {
			v = 23
		}
		return 232 + v
	}
	q := func(v float32) int {
		levels := []float32{0, 0.2, 0.4, 0.55, 0.7, 0.9}
		best, bi := float32(10), 0
		for i, l := range levels {
			d := Abs(v - l)
			if d < best {
				best, bi = d, i
			}
		}
		return bi
	}
	return 16 + 36*q(r) + 6*q(g) + q(b)
}

// ToANSI256 renders the framebuffer using the 256-color palette (fallback for
// terminals without truecolor support).
func (fb *Framebuffer) ToANSI256() string {
	var b strings.Builder
	for y := 0; y < fb.H; y += 2 {
		for x := 0; x < fb.W; x++ {
			top := fb.Pix[y*fb.W+x]
			bot := top
			if y+1 < fb.H {
				bot = fb.Pix[(y+1)*fb.W+x]
			}
			fmt.Fprintf(&b, "\x1b[48;5;%dm\x1b[38;5;%dm▀\x1b[0m", xterm256(top), xterm256(bot))
		}
		if y+2 < fb.H {
			b.WriteString("\x1b[2K\r\n")
		}
	}
	return b.String()
}

// Print writes the ANSI frame to stdout, first moving the cursor home so the
// frame overwrites in place (no flicker from clearing).
func (fb *Framebuffer) Print() {
	out := bufio.NewWriter(os.Stdout)
	out.WriteString("\x1b[H")
	out.WriteString(fb.ToANSI())
	out.Flush()
}
