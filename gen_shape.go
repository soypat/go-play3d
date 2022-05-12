//go:build !js

package main

import (
	"log"
	"os"

	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/render"
	"github.com/soypat/sdf3ui/uirender"
)

//go:generate go run gen_shape.go

func main() {
	const quality = 20
	sp, _ := form3.Sphere(1)
	fp, _ := os.Create("shape.tri")
	render.CreateSTL("sphere.stl", render.NewOctreeRenderer(sp, quality))
	err := uirender.EncodeRenderer(fp, render.NewOctreeRenderer(sp, quality))
	if err != nil {
		log.Fatal(err)
	}
	model, _ := render.RenderAll(render.NewOctreeRenderer(sp, quality))
	fp.Close()
	m := NewSDFMesh(model)
	render.CreateSTL("sdf.stl", render.NewOctreeRenderer(m, 20))
}
