// Command orbit renders a Phong-shaded sphere floating above a textured plane
// lit by a moving point light, writing a still PNG. It showcases smooth
// normals, per-pixel specular highlights and distance-attenuated point lights.
package main

import (
	"flag"
	"fmt"
	"os"

	x3d "creaddinscart.studio.go.sl/3d-x"
)

func main() {
	out := flag.String("out", "orbit.png", "output PNG path")
	w := flag.Int("w", 320, "render width")
	h := flag.Int("h", 200, "render height")
	flag.Parse()

	fb := x3d.NewFramebuffer(*w, *h)
	cam := x3d.NewCamera(x3d.V3(3.5, 2.2, 3.5), x3d.V3(0, 0.3, 0), float32(*w)/float32(*h))

	scene := x3d.NewScene()
	scene.Lights = []x3d.Light{
		x3d.NewAmbient(x3d.Hex("#20242e"), 1),
		x3d.NewDirectional(x3d.Hex("#fff2e0"), 0.55, x3d.V3(-0.4, -1, -0.5)),
		x3d.NewPoint(x3d.Hex("#66b3ff"), 1.1, x3d.V3(2.2, 2.6, 1.4), 14),
	}
	ground := x3d.Plane(8, 8, 1)
	scene.AddMesh("ground", ground,
		x3d.NewMaterial(x3d.Hex("#3a3f4a")).Textured(x3d.CheckerTexture(8, 8, x3d.Hex("#2b2f38"), x3d.Hex("#454b58"))))
	ball := scene.AddMesh("ball", x3d.Sphere(1, 32, 40),
		x3d.NewMaterial(x3d.Hex("#d94f30")).Glossy(96, x3d.RGB(1, 1, 1)))
	ball.Position = x3d.V3(0, 1.05, 0)

	r := x3d.NewRenderer(fb, cam)
	r.Render(scene)

	if err := fb.SavePNG(*out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", *out)
	fmt.Println(x3d.Brand())
}
