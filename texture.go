package x3d

import (
	"image"
	"image/color"
	"math"
)

// WrapMode controls texture coordinate addressing outside [0, 1].
type WrapMode int

const (
	WrapRepeat WrapMode = iota
	WrapClamp
	WrapMirror
)

// Texture is an RGBA8-backed 2D texture sampled with bilinear filtering.
type Texture struct {
	W, H  int
	Pix   []Color
	Wrap  WrapMode
	Neare bool // nearest instead of bilinear
}

// NewTexture allocates a solid-black texture of the given size.
func NewTexture(w, h int) *Texture {
	return &Texture{W: w, H: h, Pix: make([]Color, w*h), Wrap: WrapRepeat}
}

// Set writes a single texel.
func (t *Texture) Set(x, y int, c Color) {
	if x >= 0 && x < t.W && y >= 0 && y < t.H {
		t.Pix[y*t.W+x] = c
	}
}

// At reads a single texel (clamped).
func (t *Texture) At(x, y int) Color {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= t.W {
		x = t.W - 1
	}
	if y >= t.H {
		y = t.H - 1
	}
	return t.Pix[y*t.W+x]
}

func (t *Texture) wrapCoord(i, size int) int {
	switch t.Wrap {
	case WrapClamp:
		return int(Clamp(float32(i), 0, float32(size-1)))
	case WrapMirror:
		m := i % (2 * size)
		if m < 0 {
			m += 2 * size
		}
		if m >= size {
			m = 2*size - 1 - m
		}
		return m
	default: // repeat
		m := i % size
		if m < 0 {
			m += size
		}
		return m
	}
}

// Sample evaluates the texture at continuous UV coordinates.
func (t *Texture) Sample(u, v float32) Color {
	if t.W == 0 || t.H == 0 {
		return White
	}
	// map uv -> texel center
	fx := u*float32(t.W) - 0.5
	fy := v*float32(t.H) - 0.5
	x0 := int(math.Floor(float64(fx)))
	y0 := int(math.Floor(float64(fy)))
	if t.Neare {
		return t.At(t.wrapCoord(x0, t.W), t.wrapCoord(y0, t.H))
	}
	tx := float32(fx) - float32(x0)
	ty := float32(fy) - float32(y0)
	c00 := t.At(t.wrapCoord(x0, t.W), t.wrapCoord(y0, t.H))
	c10 := t.At(t.wrapCoord(x0+1, t.W), t.wrapCoord(y0, t.H))
	c01 := t.At(t.wrapCoord(x0, t.W), t.wrapCoord(y0+1, t.H))
	c11 := t.At(t.wrapCoord(x0+1, t.W), t.wrapCoord(y0+1, t.H))
	top := c00.Mix(c10, tx)
	bot := c01.Mix(c11, tx)
	return top.Mix(bot, ty)
}

// FromImage builds a texture from any standard library image.
func FromImage(img image.Image) *Texture {
	b := img.Bounds()
	t := NewTexture(b.Dx(), b.Dy())
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			c := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			t.Pix[y*t.W+x] = FromNRGBA(c)
		}
	}
	return t
}

// ToImage converts the texture to an image.NRGBA.
func (t *Texture) ToImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, t.W, t.H))
	for y := 0; y < t.H; y++ {
		for x := 0; x < t.W; x++ {
			img.SetNRGBA(x, y, t.Pix[y*t.W+x].ToColorNRGBA())
		}
	}
	return img
}

// Procedural textures ------------------------------------------------------

// SolidTexture returns a texture filled with one color.
func SolidTexture(c Color) *Texture {
	t := NewTexture(1, 1)
	t.Pix[0] = c
	return t
}

// CheckerTexture builds a checkerboard of two colors.
func CheckerTexture(cols, rows int, a, b Color) *Texture {
	t := NewTexture(cols*8, rows*8)
	for y := 0; y < t.H; y++ {
		for x := 0; x < t.W; x++ {
			cx := x / 8
			cy := y / 8
			if (cx+cy)%2 == 0 {
				t.Set(x, y, a)
			} else {
				t.Set(x, y, b)
			}
		}
	}
	return t
}

// GradientTexture builds a horizontal linear gradient between two colors.
func GradientTexture(w int, a, b Color) *Texture {
	t := NewTexture(w, 1)
	for x := 0; x < w; x++ {
		t.Set(x, 0, a.Mix(b, float32(x)/float32(w-1)))
	}
	return t
}

func hash2(x, y int) float32 {
	h := uint32(x*374761393 + y*668265263)
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float32(h&0xffffff) / float32(0xffffff)
}

func valueNoise(x, y float32) float32 {
	ix, iy := int(math.Floor(float64(x))), int(math.Floor(float64(y)))
	fx := x - float32(ix)
	fy := y - float32(iy)
	sx := fx * fx * (3 - 2*fx)
	sy := fy * fy * (3 - 2*fy)
	a := hash2(ix, iy)
	b := hash2(ix+1, iy)
	c := hash2(ix, iy+1)
	d := hash2(ix+1, iy+1)
	return Lerp(Lerp(a, b, sx), Lerp(c, d, sx), sy)
}

// FBM returns fractal Brownian motion noise in [0, 1].
func FBM(x, y float32, octaves int) float32 {
	sum := float32(0)
	amp := float32(0.5)
	freq := float32(1)
	for i := 0; i < octaves; i++ {
		sum += amp * valueNoise(x*freq, y*freq)
		freq *= 2
		amp *= 0.5
	}
	return Clamp01(sum)
}

// NoiseTexture builds a grayscale value-noise texture.
func NoiseTexture(w, h int, scale float32, octaves int) *Texture {
	t := NewTexture(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := FBM(float32(x)/float32(w)*scale, float32(y)/float32(h)*scale, octaves)
			t.Set(x, y, RGB(n, n, n))
		}
	}
	return t
}

// BrickTexture builds a simple brick wall texture.
func BrickTexture(w, h int, mortar, brick Color) *Texture {
	t := NewTexture(w, h)
	bw, bh, gap := w/4, h/8, max(1, w/64)
	for y := 0; y < h; y++ {
		row := y / bh
		off := (row % 2) * (bw / 2)
		for x := 0; x < w; x++ {
			xx := (x + off) % bw
			if y%bh < gap || xx < gap {
				t.Set(x, y, mortar)
			} else {
				sh := hash2(x/bw, y/bh)
				t.Set(x, y, brick.Mul(0.8+0.3*sh))
			}
		}
	}
	return t
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Fill draws a solid color over the whole texture.
func (t *Texture) Fill(c Color) {
	for i := range t.Pix {
		t.Pix[i] = c
	}
}
