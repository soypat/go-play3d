//go:build !js

package main

import (
	"log"
	"os"
	"unsafe"

	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/render"
	"github.com/soypat/sdf3ui/uirender"
)

//go:generate go run .

func main() {
	const quality = 20
	sp, _ := form3.Sphere(1)
	fp, _ := os.Create("shape.tri")
	render.CreateSTL("sphere.stl", render.NewOctreeRenderer(sp, quality))
	err := uirender.EncodeRenderer(fp, render.NewOctreeRenderer(sp, quality))
	if err != nil {
		log.Fatal(err)
	}
	model := icosphere(3)
	// model, _ := render.RenderAll(render.NewOctreeRenderer(sp, quality))
	fp.Close()
	// m := NewGonumSDFMesh(convertToRenderTriangles(model))
	m := NewSDFMesh(convertToRenderTriangles(model))
	render.CreateSTL("sdf.stl", render.NewOctreeRenderer(m, 20))
}

func convertToRenderTriangles(t []Triangle) []render.Triangle3 {
	return *(*[]render.Triangle3)(unsafe.Pointer(&t))
	// c := make([]render.Triangle3, len(t))
	// for i := range t {
	// 	c[i] = render.Triangle3(t[i])
	// }
	// return c
}
