//go:build js

package main

import (
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3/must3"
	"github.com/soypat/sdf/render"
	"github.com/soypat/three"
)

func addObjects(grp three.Group) {
	grp.Add(three.NewAxesHelper(1))
	grp.Add(drawTransparent(must3.Cylinder(2, .5, .1), 80, "red"))
}

func drawTransparent(obj sdf.SDF3, cells int, color string) three.Object3D {
	t, err := render.RenderAll(render.NewOctreeRenderer(obj, cells))
	if err != nil {
		panic(err)
	}
	tris := make([]Triangle, len(t))
	for i := range tris {
		tris[i] = Triangle{Vec(t[i][0]), Vec(t[i][1]), Vec(t[i][2])}
	}
	grp := three.NewGroup()
	grp.Add(triangleMesh(tris, phongMaterial(color, .4), nil))
	// grp.Add(triangleOutlines(tris, lineColor(color)))
	return grp
}
