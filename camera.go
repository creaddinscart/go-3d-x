package x3d

import "math"

// Camera holds view and projection state and produces the view-projection
// matrix used by the renderer.
type Camera struct {
	Position Vec3
	Target   Vec3
	Up       Vec3

	FOVY   float32 // radians (perspective)
	Aspect float32
	Near   float32
	Far    float32

	Orthographic bool
	OrthoHeight  float32
}

// NewCamera builds a perspective camera looking from eye toward target.
func NewCamera(eye, target Vec3, aspect float32) *Camera {
	return &Camera{
		Position: eye, Target: target, Up: UnitY,
		FOVY: float32(60 * math.Pi / 180), Aspect: aspect,
		Near: 0.05, Far: 1000,
	}
}

// View returns the world-to-view matrix.
func (c *Camera) View() Mat4 { return LookAt(c.Position, c.Target, c.Up) }

// Projection returns the view-to-clip matrix.
func (c *Camera) Projection() Mat4 {
	if c.Orthographic {
		h := c.OrthoHeight / 2
		w := h * c.Aspect
		return Ortho(-w, w, -h, h, c.Near, c.Far)
	}
	return Perspective(c.FOVY, c.Aspect, c.Near, c.Far)
}

// ViewProjection returns the combined view * projection matrix.
func (c *Camera) ViewProjection() Mat4 { return c.Projection().Mul(c.View()) }

// Orbit returns the eye position for a spherical orbit around Target.
func (c *Camera) Orbit(yaw, pitch, dist float32) Vec3 {
	x := dist * float32(math.Cos(float64(pitch))) * float32(math.Sin(float64(yaw)))
	y := dist * float32(math.Sin(float64(pitch)))
	z := dist * float32(math.Cos(float64(pitch))) * float32(math.Cos(float64(yaw)))
	return c.Target.Add(V3(x, y, z))
}

// LookDir returns the normalized forward direction of the camera.
func (c *Camera) LookDir() Vec3 { return c.Target.Sub(c.Position).Normalize() }

// ScreenRay casts a ray from the camera through normalized device coordinates
// ndc in [-1, 1] (x right, y up). Useful for picking.
func (c *Camera) ScreenRay(ndc Vec2) (origin, dir Vec3) {
	origin = c.Position
	vp := c.ViewProjection().Inverse()
	p := vp.MulVec4(V4(ndc.X, ndc.Y, 1, 1)).XYZ()
	q := vp.MulVec4(V4(ndc.X, ndc.Y, -1, 1)).XYZ()
	return origin, p.Sub(q).Normalize()
}
