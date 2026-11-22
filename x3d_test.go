package x3d

import (
	"bytes"
	"math"
	"testing"
)

func approx(a, b float32, eps float32) bool { return Abs(a-b) < eps }

func TestMat4Inverse(t *testing.T) {
	m := Translate(V3(1, 2, 3)).Mul(RotateY(Deg2Rad(37))).Mul(Scale(V3(2, 3, 4)))
	id := m.Mul(m.Inverse())
	for i := 0; i < 16; i++ {
		want := float32(0)
		if i%5 == 0 {
			want = 1
		}
		if !approx(id[i], want, 1e-3) {
			t.Fatalf("inverse product at %d = %v, want %v", i, id[i], want)
		}
	}
}

func TestQuaternionRoundTrip(t *testing.T) {
	q := QuatFromAxisAngle(V3(1, 2, 3), Deg2Rad(50))
	m := q.Mat4()
	p := m.TransformPoint(UnitX)
	want := V3(1, 0, 0)
	// rotating unit X by 50deg about (1,2,3) keeps unit length
	if !approx(p.Length(), want.Length(), 1e-3) {
		t.Fatalf("rotated length = %v, want %v", p.Length(), want.Length())
	}
}

func TestColorHex(t *testing.T) {
	c := Hex("#ff8000")
	if !approx(c.R, 1, 1e-3) || !approx(c.G, 128.0/255, 1e-2) || !approx(c.B, 0, 1e-3) {
		t.Fatalf("Hex parse wrong: %+v", c)
	}
}

func TestTextureSample(t *testing.T) {
	ck := CheckerTexture(2, 2, Black, White)
	a := ck.Sample(0.1, 0.1)
	b := ck.Sample(0.6, 0.1)
	if a.R == b.R {
		t.Fatalf("checker did not alternate: a=%v b=%v", a, b)
	}
}

func TestEasing(t *testing.T) {
	if !approx(EaseOutQuad(0.5), 0.75, 1e-5) {
		t.Fatalf("EaseOutQuad(0.5) = %v, want 0.75", EaseOutQuad(0.5))
	}
	if !approx(EaseLinear(0.3), 0.3, 1e-6) {
		t.Fatal("EaseLinear broken")
	}
}

func TestOBJRoundTrip(t *testing.T) {
	cube := Cube(1)
	var buf bytes.Buffer
	if err := cube.WriteOBJ(&buf); err != nil {
		t.Fatal(err)
	}
	back, err := LoadOBJ(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if back.TriangleCount() != cube.TriangleCount() {
		t.Fatalf("triangles %d != %d", back.TriangleCount(), cube.TriangleCount())
	}
}

func TestRasterizerDrawsSilhouette(t *testing.T) {
	fb := NewFramebuffer(120, 80)
	cam := NewCamera(V3(0, 0, 4), Zero3, 120.0/80.0)
	sc := NewScene()
	sc.AddMesh("cube", Cube(2), NewMaterial(Red))
	NewRenderer(fb, cam).Render(sc)

	nonBg := 0
	for _, p := range fb.Pix {
		if p != sc.Background {
			nonBg++
		}
	}
	if float32(nonBg) < float32(len(fb.Pix))*0.1 {
		t.Fatalf("rasterizer drew too few pixels: %d/%d", nonBg, len(fb.Pix))
	}
}

func TestAnimationAppliesPosition(t *testing.T) {
	sc := NewScene()
	n := sc.AddMesh("m", Cube(1), NewMaterial(White))
	n.Position = Zero3
	clip := NewClip("slide")
	clip.AddPosition("m", []Keyframe{
		{Time: 0, Value: V3(-2, 0, 0)},
		{Time: 1, Value: V3(2, 0, 0), Ease: EaseLinear},
	})
	an := NewAnimator(clip, sc)
	an.Advance(0.5)
	if !approx(n.Position.X, 0, 1e-3) {
		t.Fatalf("animated X = %v, want 0", n.Position.X)
	}
}

func TestFramebufferANSI(t *testing.T) {
	fb := NewFramebuffer(4, 4)
	fb.Fill(RGB(1, 0, 0))
	s := fb.ToANSI()
	if !bytes.Contains([]byte(s), []byte("▀")) {
		t.Fatal("ANSI output missing half-block glyph")
	}
	if !bytes.Contains([]byte(s), []byte("48;2;")) {
		t.Fatal("ANSI output missing truecolor background code")
	}
}

func BenchmarkRenderCube(b *testing.B) {
	fb := NewFramebuffer(200, 120)
	cam := NewCamera(V3(0, 0, 4), Zero3, 200.0/120.0)
	sc := NewScene()
	sc.AddMesh("sphere", Sphere(1.5, 24, 32), NewMaterial(White))
	r := NewRenderer(fb, cam)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Render(sc)
	}
	_ = math.Pi
}
