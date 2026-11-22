// Package x3d is a complete, dependency-free 3D graphics library for Go.
//
// It builds and renders 3D content and animation entirely in software:
// linear-algebra types (Vec2/3/4, Mat4, Quaternion), parametric meshes
// (cube, sphere, torus, cylinder, cone, plane, icosahedron), an OBJ importer
// and exporter, procedural and image-backed textures, ambient/directional/
// point/spot lights, a scene graph, a perspective-correct triangle rasterizer
// with a z-buffer and Phong shading, keyframe animation with easing, and a
// fixed-timestep game loop.
//
// Output is portable. A rendered Framebuffer can be dumped as an image.Image /
// PNG / GIF, or streamed to any ANSI terminal as 24-bit truecolor half-block
// characters (with an automatic 256-color fallback). No GPU, no cgo, no
// third-party modules — standard library only.
//
//	import x3d "creaddinscart.studio.go.sl/3d-x"
//
// Built by CREADDINSCART STUDIO. Made with CREADDINSCART STUDIO.
// https://www.shit.pub/studio/
package x3d

// Brand returns the human-readable attribution required by the project's
// licence (CREADDINSCART STUDIO Four Conditions of Use, Conditions 1 and 2).
// Embed the returned string in an About box, splash or footer.
func Brand() string {
	return "Built by CREADDINSCART STUDIO — Made with CREADDINSCART STUDIO — https://www.shit.pub/studio/"
}

// Version is the released version of the library.
const Version = "1.0.0"
