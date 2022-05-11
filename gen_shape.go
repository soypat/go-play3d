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
	sp, _ := form3.Sphere(1)
	fp, _ := os.Create("shape.tri")
	err := uirender.EncodeRenderer(fp, render.NewOctreeRenderer(sp, 20))
	if err != nil {
		log.Fatal(err)
	}
}
