package x3d

import "math"

type drawItem struct {
	world Mat4
	ent   *Entity
}

// clipVertex carries homogeneous clip-space position plus the attributes we
// interpolate through the pipeline.
type clipVertex struct {
	clip   Vec4
	world  Vec3
	normal Vec3
	uv     Vec2
}

// screenVertex is a projected vertex ready for rasterization.
type screenVertex struct {
	x, y, z float32
	invW    float32
	world   Vec3
	normal  Vec3
	uv      Vec2
	col     Color // precomputed for flat/gouraud shading
}

// Renderer draws scenes into a framebuffer with a fixed-function-style
// software pipeline: transform -> clip -> perspective-correct rasterize ->
// per-fragment lighting.
type Renderer struct {
	FB           *Framebuffer
	Camera       *Camera
	CullBackface bool
	lights       []Light
}

// NewRenderer builds a renderer targeting fb with cam.
func NewRenderer(fb *Framebuffer, cam *Camera) *Renderer {
	return &Renderer{FB: fb, Camera: cam, CullBackface: true}
}

// Render draws the whole scene into the renderer's framebuffer.
func (r *Renderer) Render(s *Scene) {
	r.lights = s.Lights
	r.FB.Clear(s.Background, 1)
	vp := r.Camera.ViewProjection()
	for _, it := range s.visibleEntities() {
		mvp := vp.Mul(it.world)
		r.drawMesh(it.ent, mvp, it.world)
	}
}

func (r *Renderer) drawMesh(e *Entity, mvp, world Mat4) {
	m := e.Mesh
	vc := make([]clipVertex, len(m.Vertices))
	for i, v := range m.Vertices {
		vc[i] = clipVertex{
			clip:   mvp.MulVec4(v.Position.ToVec4(1)),
			world:  world.TransformPoint(v.Position),
			normal: world.TransformVector(v.Normal).Normalize(),
			uv:     v.UV,
		}
	}
	for t := 0; t+2 < len(m.Indices); t += 3 {
		tri := [3]clipVertex{vc[m.Indices[t]], vc[m.Indices[t+1]], vc[m.Indices[t+2]]}
		poly := clipTriangle(tri)
		if len(poly) < 3 {
			continue
		}
		for i := 1; i+1 < len(poly); i++ {
			r.rasterize(poly[0], poly[i], poly[i+1], e.Material)
		}
	}
}

// frustum planes in clip space: keep points where a*x+b*y+c*z+d*w >= 0.
var frustumPlanes = [6]Vec4{
	{1, 0, 0, 1}, {-1, 0, 0, 1},
	{0, 1, 0, 1}, {0, -1, 0, 1},
	{0, 0, 1, 1}, {0, 0, -1, 1},
}

func planeDist(p, pl Vec4) float32 {
	return pl.X*p.X + pl.Y*p.Y + pl.Z*p.Z + pl.W*p.W
}

func lerpCV(a, b clipVertex, t float32) clipVertex {
	return clipVertex{
		clip:   a.clip.Mul(1 - t).Add(b.clip.Mul(t)),
		world:  a.world.Lerp(b.world, t),
		normal: a.normal.Lerp(b.normal, t).Normalize(),
		uv:     V2(Lerp(a.uv.X, b.uv.X, t), Lerp(a.uv.Y, b.uv.Y, t)),
	}
}

// clipTriangle clips a triangle against all frustum planes, returning a convex
// polygon (0..6 vertices) in clip space.
func clipTriangle(tri [3]clipVertex) []clipVertex {
	poly := tri[:]
	for _, pl := range frustumPlanes {
		if len(poly) == 0 {
			return nil
		}
		var out []clipVertex
		for i := 0; i < len(poly); i++ {
			cur := poly[i]
			nxt := poly[(i+1)%len(poly)]
			dc := planeDist(cur.clip, pl)
			dn := planeDist(nxt.clip, pl)
			if dc >= 0 {
				out = append(out, cur)
			}
			if (dc >= 0) != (dn >= 0) {
				t := dc / (dc - dn)
				out = append(out, lerpCV(cur, nxt, t))
			}
		}
		poly = out
	}
	return poly
}

func (r *Renderer) project(cv clipVertex, mat Material) screenVertex {
	w := cv.clip.W
	if w == 0 {
		w = 1e-6
	}
	ndc := Vec3{cv.clip.X / w, cv.clip.Y / w, cv.clip.Z / w}
	sv := screenVertex{
		x:      (ndc.X*0.5 + 0.5) * float32(r.FB.W),
		y:      (1 - (ndc.Y*0.5 + 0.5)) * float32(r.FB.H),
		z:      ndc.Z,
		invW:   1 / w,
		world:  cv.world,
		normal: cv.normal,
		uv:     cv.uv,
	}
	if mat.ShadeMode == ShadeFlat || mat.ShadeMode == ShadeGouraud {
		sv.col = r.shade(mat, cv.world, cv.normal, cv.uv)
	}
	return sv
}

func (r *Renderer) rasterize(a, b, c clipVertex, mat Material) {
	sa, sb, sc := r.project(a, mat), r.project(b, mat), r.project(c, mat)

	area := edge(sa, sb) + edge(sb, sc) + edge(sc, sa)
	if area == 0 {
		return
	}
	if r.CullBackface && !mat.DoubleSided && area > 0 {
		// screen-space Y is flipped, so positive signed area is back-facing
		return
	}

	// flat shading uses one normal for the whole face
	if mat.ShadeMode == ShadeFlat {
		faceN := sb.world.Sub(sa.world).Cross(sc.world.Sub(sa.world)).Normalize()
		if faceN.Dot(sa.normal) < 0 {
			faceN = faceN.Neg()
		}
		col := r.shade(mat, sa.world, faceN, sa.uv)
		sa.col, sb.col, sc.col = col, col, col
	}

	minX := int(math.Floor(float64(Min3(sa.x, sb.x, sc.x))))
	maxX := int(math.Ceil(float64(Max3(sa.x, sb.x, sc.x))))
	minY := int(math.Floor(float64(Min3(sa.y, sb.y, sc.y))))
	maxY := int(math.Ceil(float64(Max3(sa.y, sb.y, sc.y))))
	minX = int(Max(float32(minX), 0))
	minY = int(Max(float32(minY), 0))
	maxX = int(Min(float32(maxX), float32(r.FB.W-1)))
	maxY = int(Min(float32(maxY), float32(r.FB.H-1)))

	invArea := 1 / area
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			cx := float32(px) + 0.5
			cy := float32(py) + 0.5
			w0 := baryWeight(sb, sc, cx, cy)
			w1 := baryWeight(sc, sa, cx, cy)
			w2 := area - w0 - w1
			l0, l1, l2 := w0*invArea, w1*invArea, w2*invArea
			if l0 < 0 || l1 < 0 || l2 < 0 {
				continue
			}

			iw := l0*sa.invW + l1*sb.invW + l2*sc.invW
			k := 1 / iw
			z := (l0*sa.z*sa.invW + l1*sb.z*sb.invW + l2*sc.z*sc.invW) * k
			if z >= r.FB.Depth[py*r.FB.W+px] {
				continue
			}
			r.FB.Depth[py*r.FB.W+px] = z
			r.FB.Pix[py*r.FB.W+px] = r.shadeInterp(mat, l0, l1, l2, sa, sb, sc, k)
		}
	}
}

// shadeInterp computes the final pixel color for the current shade mode using
// perspective-correct interpolation of the relevant attributes.
func (r *Renderer) shadeInterp(mat Material, l0, l1, l2 float32, sa, sb, sc screenVertex, k float32) Color {
	if mat.ShadeMode == ShadePhong {
		wn := V3(
			(l0*sa.invW)*sa.normal.X+(l1*sb.invW)*sb.normal.X+(l2*sc.invW)*sc.normal.X,
			(l0*sa.invW)*sa.normal.Y+(l1*sb.invW)*sb.normal.Y+(l2*sc.invW)*sc.normal.Y,
			(l0*sa.invW)*sa.normal.Z+(l1*sb.invW)*sb.normal.Z+(l2*sc.invW)*sc.normal.Z,
		).Mul(k).Normalize()
		wp := V3(
			(l0*sa.invW)*sa.world.X+(l1*sb.invW)*sb.world.X+(l2*sc.invW)*sc.world.X,
			(l0*sa.invW)*sa.world.Y+(l1*sb.invW)*sb.world.Y+(l2*sc.invW)*sc.world.Y,
			(l0*sa.invW)*sa.world.Z+(l1*sb.invW)*sb.world.Z+(l2*sc.invW)*sc.world.Z,
		).Mul(k)
		uv := V2(
			(l0*sa.invW)*sa.uv.X+(l1*sb.invW)*sb.uv.X+(l2*sc.invW)*sc.uv.X,
			(l0*sa.invW)*sa.uv.Y+(l1*sb.invW)*sb.uv.Y+(l2*sc.invW)*sc.uv.Y,
		).Mul(k)
		return r.shade(mat, wp, wn, uv)
	}
	// flat / gouraud: perspective-correct blend of precomputed vertex colors
	cR := l0 * sa.invW
	cG := l1 * sb.invW
	cB := l2 * sc.invW
	s := cR + cG + cB
	if s == 0 {
		return Color{0, 0, 0, mat.Opacity}
	}
	return Color{
		(cR*sa.col.R + cG*sb.col.R + cB*sc.col.R) / s,
		(cR*sa.col.G + cG*sb.col.G + cB*sc.col.G) / s,
		(cR*sa.col.B + cG*sb.col.B + cB*sc.col.B) / s,
		mat.Opacity,
	}
}

// shade evaluates the lighting model for one fragment.
func (r *Renderer) shade(mat Material, worldPos, normal Vec3, uv Vec2) Color {
	n := normal.Normalize()
	view := r.Camera.Position.Sub(worldPos).Normalize()

	// base albedo, optionally modulated by the diffuse texture
	albedo := mat.Diffuse
	ambientTerm := mat.Ambient
	if mat.Texture != nil {
		tex := mat.Texture.Sample(uv.X, uv.Y)
		albedo = albedo.Modulate(tex)
		ambientTerm = ambientTerm.Modulate(tex)
	}

	out := mat.Emissive
	out.R += ambientTerm.R
	out.G += ambientTerm.G
	out.B += ambientTerm.B

	for _, L := range r.lights {
		var lightDir Vec3
		atten := float32(1)
		switch L.Kind {
		case LightAmbient:
			a := L.Color.Mul(L.Intensity)
			out.R += ambientTerm.R * a.R
			out.G += ambientTerm.G * a.G
			out.B += ambientTerm.B * a.B
			continue
		case LightDirectional:
			lightDir = L.Direction.Neg()
		case LightPoint:
			toL := L.Position.Sub(worldPos)
			d := toL.Length()
			if d == 0 {
				continue
			}
			lightDir = toL.Div(d)
			if L.Range > 0 {
				atten = Clamp01(1 - d/L.Range)
			}
		case LightSpot:
			toL := worldPos.Sub(L.Position)
			d := toL.Length()
			if d == 0 {
				continue
			}
			toL = toL.Div(d)
			if toL.Dot(L.Direction) < L.SpotCos {
				continue
			}
			lightDir = toL.Neg()
			atten = 1
			if L.Range > 0 {
				atten = Clamp01(1 - d/L.Range)
			}
		}

		ndl := n.Dot(lightDir)
		if ndl < 0 {
			ndl = 0
		}
		lc := L.Color.Mul(L.Intensity * atten)
		// diffuse
		out.R += albedo.R * lc.R * ndl
		out.G += albedo.G * lc.G * ndl
		out.B += albedo.B * lc.B * ndl
		// specular (Blinn-Phong)
		if ndl > 0 {
			hv := lightDir.Add(view).Normalize()
			spe := n.Dot(hv)
			if spe > 0 {
				spe = float32(math.Pow(float64(spe), float64(mat.Shininess)))
				out.R += mat.Specular.R * lc.R * spe
				out.G += mat.Specular.G * lc.G * spe
				out.B += mat.Specular.B * lc.B * spe
			}
		}
	}

	out.R = Clamp01(out.R)
	out.G = Clamp01(out.G)
	out.B = Clamp01(out.B)
	out.A = mat.Opacity
	return out
}

func edge(a, b screenVertex) float32 { return a.x*b.y - b.x*a.y }

func baryWeight(v0, v1 screenVertex, px, py float32) float32 {
	return (v1.x-v0.x)*(py-v0.y) - (v1.y-v0.y)*(px-v0.x)
}

// Min3 returns the smallest of three floats.
func Min3(a, b, c float32) float32 { return Min(Min(a, b), c) }

// Max3 returns the largest of three floats.
func Max3(a, b, c float32) float32 { return Max(Max(a, b), c) }
