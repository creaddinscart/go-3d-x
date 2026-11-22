package x3d

import "math"

// LightKind selects a lighting model.
type LightKind int

const (
	LightAmbient LightKind = iota
	LightDirectional
	LightPoint
	LightSpot
)

// Light is a single scene light source. All positions/directions are in world
// space. Intensity scales the color.
type Light struct {
	Kind      LightKind
	Color     Color
	Intensity float32
	Direction Vec3    // directional/spot: direction the light travels
	Position  Vec3    // point/spot: emitter position
	Range     float32 // point/spot: falloff radius (0 = no falloff)
	SpotCos   float32 // spot: cosine of the outer cone angle
}

// NewAmbient builds an ambient light.
func NewAmbient(c Color, intensity float32) Light {
	return Light{Kind: LightAmbient, Color: c, Intensity: intensity}
}

// NewDirectional builds a directional light pointing along dir (normalized).
func NewDirectional(c Color, intensity float32, dir Vec3) Light {
	return Light{Kind: LightDirectional, Color: c, Intensity: intensity, Direction: dir.Normalize()}
}

// NewPoint builds a point light at pos with optional range.
func NewPoint(c Color, intensity float32, pos Vec3, rng float32) Light {
	return Light{Kind: LightPoint, Color: c, Intensity: intensity, Position: pos, Range: rng}
}

// NewSpot builds a spot light at pos aimed along dir with a cone half-angle in
// radians.
func NewSpot(c Color, intensity float32, pos, dir Vec3, angleRad, rng float32) Light {
	return Light{
		Kind: LightSpot, Color: c, Intensity: intensity,
		Position: pos, Direction: dir.Normalize(),
		Range: rng, SpotCos: float32(math.Cos(float64(angleRad))),
	}
}

// DefaultLights returns a pleasant three-light studio rig.
func DefaultLights() []Light {
	return []Light{
		NewAmbient(RGB(0.18, 0.18, 0.22), 1),
		NewDirectional(RGB(1, 0.98, 0.92), 0.9, V3(-0.5, -1, -0.35)),
		NewPoint(RGB(0.4, 0.6, 1.0), 0.6, V3(3, 3, 3), 12),
	}
}
