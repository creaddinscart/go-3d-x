// Command gif renders an animated turntable of a lit torus and a sphere and
// writes an animated GIF plus a still PNG — the visual proof that 3d-x builds
// and renders 3D content and animation.
package main

import (
	"flag"
	"fmt"
	"os"

	x3d "creaddinscart.studio.go.sl/3d-x"
)

func main() {
	out := flag.String("out", "turntable.gif", "output GIF path")
	pngOut := flag.String("png", "turntable.png", "output still PNG path")
	frames := flag.Int("frames", 36, "number of frames")
	w := flag.Int("w", 240, "render width")
	h := flag.Int("h", 150, "render height")
	flag.Parse()

	fb := x3d.NewFramebuffer(*w, *h)
	cam := x3d.NewCamera(x3d.V3(0, 1.2, 4.2), x3d.V3(0, 0, 0), float32(*w)/float32(*h))

	scene := x3d.NewScene()
	tex := x3d.BrickTexture(64, 64, x3d.Hex("#c9b79a"), x3d.Hex("#b23a2e"))
	torus := scene.AddMesh("torus", x3d.Torus(1.0, 0.4, 48, 20),
		x3d.NewMaterial(x3d.White).Textured(tex).Glossy(48, x3d.RGB(0.6, 0.6, 0.6)))
	sphere := scene.AddMesh("sphere", x3d.Icosahedron(0.55, 2),
		x3d.NewMaterial(x3d.Hex("#2b6cff")).Glossy(64, x3d.RGB(0.9, 0.9, 1.0)))
	sphere.Position = x3d.V3(0, 0, 0)
	torus.Position = x3d.V3(0, 0, 0)

	r := x3d.NewRenderer(fb, cam)
	shots := make([]*x3d.Framebuffer, 0, *frames)

	for i := 0; i < *frames; i++ {
		a := float32(i) / float32(*frames) * 2 * 3.14159265
		torus.Rotation = x3d.QuatFromAxisAngle(x3d.V3(0, 1, 0.3), a)
		sphere.Rotation = x3d.QuatFromAxisAngle(x3d.V3(1, 1, 0), a*1.5)
		r.Render(scene)
		cp := x3d.NewFramebuffer(*w, *h)
		copy(cp.Pix, fb.Pix)
		shots = append(shots, cp)
		if i == 0 {
			if err := fb.SavePNG(*pngOut); err != nil {
				fmt.Fprintln(os.Stderr, "png:", err)
			}
		}
	}

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	if err := x3d.EncodeGIF(f, shots, 6); err != nil {
		fmt.Fprintln(os.Stderr, "gif:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d frames) and %s\n", *out, len(shots), *pngOut)
	fmt.Println(x3d.Brand())
}
