package x3d

import "math"

// Vertex is a single rasterizable vertex.
type Vertex struct {
	Position Vec3
	Normal   Vec3
	UV       Vec2
}

// Mesh is an indexed triangle list.
type Mesh struct {
	Vertices []Vertex
	Indices  []uint32
}

// NewMesh returns an empty mesh.
func NewMesh() *Mesh { return &Mesh{} }

// AddTriangle appends three vertices and their indices.
func (m *Mesh) AddTriangle(a, b, c Vertex) {
	i := uint32(len(m.Vertices))
	m.Vertices = append(m.Vertices, a, b, c)
	m.Indices = append(m.Indices, i, i+1, i+2)
}

// AddQuad appends a quad as two triangles (a,b,c,d in order).
func (m *Mesh) AddQuad(a, b, c, d Vertex) {
	m.AddTriangle(a, b, c)
	m.AddTriangle(a, c, d)
}

// TriangleCount returns the number of triangles in the mesh.
func (m *Mesh) TriangleCount() int { return len(m.Indices) / 3 }

// Bounds returns the axis-aligned bounding box of the vertex positions.
func (m *Mesh) Bounds() (min, max Vec3) {
	if len(m.Vertices) == 0 {
		return
	}
	min, max = m.Vertices[0].Position, m.Vertices[0].Position
	for _, v := range m.Vertices {
		p := v.Position
		min.X, max.X = Min(p.X, min.X), Max(p.X, max.X)
		min.Y, max.Y = Min(p.Y, min.Y), Max(p.Y, max.Y)
		min.Z, max.Z = Min(p.Z, min.Z), Max(p.Z, max.Z)
	}
	return
}

// ComputeNormals recomputes smooth (averaged) vertex normals from faces.
func (m *Mesh) ComputeNormals() {
	for i := range m.Vertices {
		m.Vertices[i].Normal = Zero3
	}
	for i := 0; i+2 < len(m.Indices); i += 3 {
		a := &m.Vertices[m.Indices[i]]
		b := &m.Vertices[m.Indices[i+1]]
		c := &m.Vertices[m.Indices[i+2]]
		n := b.Position.Sub(a.Position).Cross(c.Position.Sub(a.Position))
		a.Normal = a.Normal.Add(n)
		b.Normal = b.Normal.Add(n)
		c.Normal = c.Normal.Add(n)
	}
	for i := range m.Vertices {
		m.Vertices[i].Normal = m.Vertices[i].Normal.Normalize()
	}
}

// Transform applies a matrix to positions and normals (normals use the
// inverse-transpose approximation via the rotation part).
func (m *Mesh) Transform(mat Mat4) *Mesh {
	out := &Mesh{Vertices: make([]Vertex, len(m.Vertices)), Indices: append([]uint32(nil), m.Indices...)}
	nrm := mat
	for i, v := range m.Vertices {
		nv := v
		nv.Position = nrm.TransformPoint(v.Position)
		nv.Normal = nrm.TransformVector(v.Normal).Normalize()
		out.Vertices[i] = nv
	}
	return out
}

// Cube builds a unit-centered cube of the given size with sharp face normals.
func Cube(size float32) *Mesh {
	h := size / 2
	faces := []struct {
		n  Vec3
		v0 [3]float32
		v1 [3]float32
		v2 [3]float32
	}{
		{Vec3{0, 0, 1}, [3]float32{-h, -h, h}, [3]float32{h, -h, h}, [3]float32{h, h, h}},      // +Z
		{Vec3{0, 0, -1}, [3]float32{h, -h, -h}, [3]float32{-h, -h, -h}, [3]float32{-h, h, -h}}, // -Z
		{Vec3{1, 0, 0}, [3]float32{h, -h, h}, [3]float32{h, -h, -h}, [3]float32{h, h, -h}},     // +X
		{Vec3{-1, 0, 0}, [3]float32{-h, -h, -h}, [3]float32{-h, -h, h}, [3]float32{-h, h, h}},  // -X
		{Vec3{0, 1, 0}, [3]float32{-h, h, h}, [3]float32{h, h, h}, [3]float32{h, h, -h}},       // +Y
		{Vec3{0, -1, 0}, [3]float32{-h, -h, -h}, [3]float32{h, -h, -h}, [3]float32{h, -h, h}},  // -Y
	}
	m := NewMesh()
	uvs := [4]Vec2{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	for _, f := range faces {
		a := V3(f.v0[0], f.v0[1], f.v0[2])
		b := V3(f.v1[0], f.v1[1], f.v1[2])
		c := V3(f.v2[0], f.v2[1], f.v2[2])
		d := a.Add(c).Sub(b)
		m.AddQuad(
			Vertex{a, f.n, uvs[0]},
			Vertex{b, f.n, uvs[1]},
			Vertex{c, f.n, uvs[2]},
			Vertex{d, f.n, uvs[3]},
		)
	}
	return m
}

// Sphere builds a UV sphere with the given radius and tessellation.
func Sphere(radius float32, stacks, slices int) *Mesh {
	if stacks < 3 {
		stacks = 3
	}
	if slices < 3 {
		slices = 3
	}
	m := NewMesh()
	for i := 0; i < stacks; i++ {
		phi0 := float32(math.Pi) * (float32(i)/float32(stacks) - 0.5)
		phi1 := float32(math.Pi) * (float32(i+1)/float32(stacks) - 0.5)
		for j := 0; j < slices; j++ {
			t0 := 2 * float32(math.Pi) * float32(j) / float32(slices)
			t1 := 2 * float32(math.Pi) * float32(j+1) / float32(slices)
			p := func(phi, theta float32, u, v float32) Vertex {
				n := V3(float32(math.Cos(float64(phi)))*float32(math.Cos(float64(theta))),
					float32(math.Sin(float64(phi))),
					float32(math.Cos(float64(phi)))*float32(math.Sin(float64(theta))))
				return Vertex{n.Mul(radius), n, V2(u, v)}
			}
			u0, u1 := float32(j)/float32(slices), float32(j+1)/float32(slices)
			v0, v1 := float32(i)/float32(stacks), float32(i+1)/float32(stacks)
			m.AddQuad(p(phi0, t0, u0, v0), p(phi0, t1, u1, v0), p(phi1, t1, u1, v1), p(phi1, t0, u0, v1))
		}
	}
	return m
}

// Plane builds an XZ plane of the given width and depth with subdivisions.
func Plane(width, depth float32, subdiv int) *Mesh {
	if subdiv < 1 {
		subdiv = 1
	}
	m := NewMesh()
	n := V3(0, 1, 0)
	for i := 0; i < subdiv; i++ {
		for j := 0; j < subdiv; j++ {
			x0 := (float32(j)/float32(subdiv) - 0.5) * width
			x1 := (float32(j+1)/float32(subdiv) - 0.5) * width
			z0 := (float32(i)/float32(subdiv) - 0.5) * depth
			z1 := (float32(i+1)/float32(subdiv) - 0.5) * depth
			u0, u1 := float32(j)/float32(subdiv), float32(j+1)/float32(subdiv)
			v0, v1 := float32(i)/float32(subdiv), float32(i+1)/float32(subdiv)
			m.AddQuad(
				Vertex{V3(x0, 0, z0), n, V2(u0, v0)},
				Vertex{V3(x1, 0, z0), n, V2(u1, v0)},
				Vertex{V3(x1, 0, z1), n, V2(u1, v1)},
				Vertex{V3(x0, 0, z1), n, V2(u0, v1)},
			)
		}
	}
	return m
}

// Torus builds a torus with the given radii and tessellation.
func Torus(R, r float32, rings, sides int) *Mesh {
	if rings < 3 {
		rings = 3
	}
	if sides < 3 {
		sides = 3
	}
	m := NewMesh()
	at := func(i, j int) Vertex {
		u := 2 * float32(math.Pi) * float32(i) / float32(rings)
		v := 2 * float32(math.Pi) * float32(j) / float32(sides)
		cu, su := float32(math.Cos(float64(u))), float32(math.Sin(float64(u)))
		cv, sv := float32(math.Cos(float64(v))), float32(math.Sin(float64(v)))
		pos := V3((R+r*cv)*cu, r*sv, (R+r*cv)*su)
		nrm := V3(cv*cu, sv, cv*su)
		return Vertex{pos, nrm, V2(float32(i)/float32(rings), float32(j)/float32(sides))}
	}
	for i := 0; i < rings; i++ {
		for j := 0; j < sides; j++ {
			m.AddQuad(at(i, j), at(i+1, j), at(i+1, j+1), at(i, j+1))
		}
	}
	return m
}

// Cylinder builds a capped cylinder along the Y axis.
func Cylinder(radius, height float32, segments int) *Mesh {
	if segments < 3 {
		segments = 3
	}
	m := NewMesh()
	hh := height / 2
	for i := 0; i < segments; i++ {
		a0 := 2 * float32(math.Pi) * float32(i) / float32(segments)
		a1 := 2 * float32(math.Pi) * float32(i+1) / float32(segments)
		c0, s0 := float32(math.Cos(float64(a0))), float32(math.Sin(float64(a0)))
		c1, s1 := float32(math.Cos(float64(a1))), float32(math.Sin(float64(a1)))
		n0, n1 := V3(c0, 0, s0), V3(c1, 0, s1)
		u0, u1 := float32(i)/float32(segments), float32(i+1)/float32(segments)
		m.AddQuad(
			Vertex{V3(c0*radius, -hh, s0*radius), n0, V2(u0, 0)},
			Vertex{V3(c1*radius, -hh, s1*radius), n1, V2(u1, 0)},
			Vertex{V3(c1*radius, hh, s1*radius), n1, V2(u1, 1)},
			Vertex{V3(c0*radius, hh, s0*radius), n0, V2(u0, 1)},
		)
		// caps
		m.AddTriangle(Vertex{Zero3, V3(0, 1, 0), V2(0.5, 0.5)},
			Vertex{V3(c0*radius, hh, s0*radius), V3(0, 1, 0), V2(c0*0.5+0.5, s0*0.5+0.5)},
			Vertex{V3(c1*radius, hh, s1*radius), V3(0, 1, 0), V2(c1*0.5+0.5, s1*0.5+0.5)})
		m.AddTriangle(Vertex{Zero3, V3(0, -1, 0), V2(0.5, 0.5)},
			Vertex{V3(c1*radius, -hh, s1*radius), V3(0, -1, 0), V2(c1*0.5+0.5, s1*0.5+0.5)},
			Vertex{V3(c0*radius, -hh, s0*radius), V3(0, -1, 0), V2(c0*0.5+0.5, s0*0.5+0.5)})
	}
	return m
}

// Cone builds a capped cone along the Y axis with its base at -height/2.
func Cone(radius, height float32, segments int) *Mesh {
	if segments < 3 {
		segments = 3
	}
	m := NewMesh()
	hh := height / 2
	apex := V3(0, hh, 0)
	for i := 0; i < segments; i++ {
		a0 := 2 * float32(math.Pi) * float32(i) / float32(segments)
		a1 := 2 * float32(math.Pi) * float32(i+1) / float32(segments)
		c0, s0 := float32(math.Cos(float64(a0))), float32(math.Sin(float64(a0)))
		c1, s1 := float32(math.Cos(float64(a1))), float32(math.Sin(float64(a1)))
		b0, b1 := V3(c0*radius, -hh, s0*radius), V3(c1*radius, -hh, s1*radius)
		nSide := a0 // approximate per-vertex normals
		_ = nSide
		na := V3(c0, radius/height, s0).Normalize()
		nb := V3(c1, radius/height, s1).Normalize()
		m.AddTriangle(Vertex{b0, na, V2(float32(i)/float32(segments), 0)},
			Vertex{b1, nb, V2(float32(i+1)/float32(segments), 0)},
			Vertex{apex, nb, V2(0.5, 1)})
		m.AddTriangle(Vertex{Zero3.SetY(-hh), V3(0, -1, 0), V2(0.5, 0.5)},
			Vertex{b1, V3(0, -1, 0), V2(c1*0.5+0.5, s1*0.5+0.5)},
			Vertex{b0, V3(0, -1, 0), V2(c0*0.5+0.5, s0*0.5+0.5)})
	}
	return m
}

// Tetrahedron builds a regular tetrahedron inscribed in a sphere of the given
// radius, with flat face normals.
func Tetrahedron(radius float32) *Mesh {
	verts := []Vec3{
		V3(1, 1, 1), V3(-1, -1, 1), V3(-1, 1, -1), V3(1, -1, -1),
	}
	for i := range verts {
		verts[i] = verts[i].Normalize().Mul(radius)
	}
	faces := [][3]int{{0, 1, 2}, {0, 3, 1}, {0, 2, 3}, {1, 3, 2}}
	m := NewMesh()
	for _, f := range faces {
		a, b, c := verts[f[0]], verts[f[1]], verts[f[2]]
		n := b.Sub(a).Cross(c.Sub(a)).Normalize()
		m.AddTriangle(Vertex{a, n, V2(0, 0)}, Vertex{b, n, V2(1, 0)}, Vertex{c, n, V2(0.5, 1)})
	}
	return m
}

// Icosahedron builds a geodesic sphere from an icosahedron subdivided
// `subdiv` times, projected onto a sphere of the given radius.
func Icosahedron(radius float32, subdiv int) *Mesh {
	t := (1 + float32(math.Sqrt(5))) / 2
	base := []Vec3{
		V3(-1, t, 0), V3(1, t, 0), V3(-1, -t, 0), V3(1, -t, 0),
		V3(0, -1, t), V3(0, 1, t), V3(0, -1, -t), V3(0, 1, -t),
		V3(t, 0, -1), V3(t, 0, 1), V3(-t, 0, -1), V3(-t, 0, 1),
	}
	for i := range base {
		base[i] = base[i].Normalize()
	}
	faces := [][3]int{
		{0, 11, 5}, {0, 5, 1}, {0, 1, 7}, {0, 7, 10}, {0, 10, 11},
		{1, 5, 9}, {5, 11, 4}, {11, 10, 2}, {10, 7, 6}, {7, 1, 8},
		{3, 9, 4}, {3, 4, 2}, {3, 2, 6}, {3, 6, 8}, {3, 8, 9},
		{4, 9, 5}, {2, 4, 11}, {6, 2, 10}, {8, 6, 7}, {9, 8, 1},
	}
	// build triangle soup then optionally subdivide
	m := NewMesh()
	addFace := func(a, b, c Vec3) {
		a, b, c = a.Normalize(), b.Normalize(), c.Normalize()
		n := b.Sub(a).Cross(c.Sub(a)).Normalize()
		m.AddTriangle(Vertex{a.Mul(radius), n, V2(0, 0)},
			Vertex{b.Mul(radius), n, V2(1, 0)}, Vertex{c.Mul(radius), n, V2(0.5, 1)})
	}
	var cur [][3]Vec3
	for _, f := range faces {
		cur = append(cur, [3]Vec3{base[f[0]], base[f[1]], base[f[2]]})
	}
	for s := 0; s < subdiv; s++ {
		var next [][3]Vec3
		for _, tri := range cur {
			ab := tri[0].Add(tri[1]).Normalize()
			bc := tri[1].Add(tri[2]).Normalize()
			ca := tri[2].Add(tri[0]).Normalize()
			next = append(next,
				[3]Vec3{tri[0], ab, ca},
				[3]Vec3{tri[1], bc, ab},
				[3]Vec3{tri[2], ca, bc},
				[3]Vec3{ab, bc, ca})
		}
		cur = next
	}
	for _, tri := range cur {
		addFace(tri[0], tri[1], tri[2])
	}
	return m
}

// SetY returns a copy of the vector with Y replaced.
func (a Vec3) SetY(y float32) Vec3 { return Vec3{a.X, y, a.Z} }
