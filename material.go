package x3d

// ShadeMode selects the per-pixel interpolation strategy.
type ShadeMode int

const (
	ShadeFlat    ShadeMode = iota // one normal per triangle
	ShadeGouraud                  // interpolated normals, lighting per vertex approximated per pixel
	ShadePhong                    // interpolated normals, full per-pixel Phong
)

// Material describes how a surface reflects light.
type Material struct {
	Ambient     Color
	Diffuse     Color
	Specular    Color
	Shininess   float32
	Emissive    Color
	Texture     *Texture
	ShadeMode   ShadeMode
	DoubleSided bool
	Opacity     float32
}

// NewMaterial returns a matte material with the given base color.
func NewMaterial(base Color) Material {
	return Material{
		Ambient:   base.Mul(0.35),
		Diffuse:   base,
		Specular:  RGB(0.2, 0.2, 0.2),
		Shininess: 16,
		Emissive:  Black,
		ShadeMode: ShadePhong,
		Opacity:   1,
	}
}

// Textured returns a copy using the given texture as the diffuse map.
func (m Material) Textured(t *Texture) Material {
	m.Texture = t
	return m
}

// Glossy returns a copy with a stronger specular response.
func (m Material) Glossy(shininess float32, spec Color) Material {
	m.Shininess = shininess
	m.Specular = spec
	return m
}

// Flat returns a copy using flat shading.
func (m Material) Flat() Material { m.ShadeMode = ShadeFlat; return m }
