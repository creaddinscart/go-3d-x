// Command cube renders a textured, lit, spinning cube and demonstrates both
// output paths of the 3d-x library: a PNG image and an ANSI terminal frame.
//
// Usage:
//
//	go run ./examples/cube            // writes cube.png and prints one ANSI frame
//	go run ./examples/cube -out f.png // custom PNG path
package main

import (
	"flag"
	"fmt"
	"os"

	x3d "creaddinscart.studio.go.sl/3d-x"
)

func main() {
	out := flag.String("out", "cube.png", "output PNG path")
	w := flag.Int("w", 320, "render width in pixels")
	h := flag.Int("h", 200, "render height in pixels")
	noANSI := flag.Bool("noansi", false, "skip printing the ANSI terminal frame")
	flag.Parse()

	fb := x3d.NewFramebuffer(*w, *h)
	cam := x3d.NewCamera(x3d.V3(0, 1.4, 4), x3d.Zero3, float32(*w)/float32(*h))

	scene := x3d.NewScene()
	cube := scene.AddMesh("cube", x3d.Cube(1.6),
		x3d.NewMaterial(x3d.Hex("#e8e8e8")).Textured(x3d.CheckerTexture(8, 8, x3d.Hex("#2b6cff"), x3d.White)))
	cube.Rotation = x3d.QuatFromAxisAngle(x3d.V3(0.4, 1, 0.2), x3d.Deg2Rad(35))

	r := x3d.NewRenderer(fb, cam)
	r.Render(scene)

	if err := fb.SavePNG(*out); err != nil {
		fmt.Fprintln(os.Stderr, "png:", err)
	}
	if !*noANSI {
		fmt.Print("\x1b[2J\x1b[H")
		fb.Print()
		fmt.Print("\x1b[0m\n")
	}
	fmt.Println(x3d.Brand())
}
