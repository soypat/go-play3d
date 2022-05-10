package main

import (
	"github.com/soypat/three"
)

func makeObjects() three.Object3D {
	grp := three.NewGroup()
	basicTriangle := Triangle{Vec{0, 0, 1}, Vec{1, 0, 0}, Vec{0, 1, 0}}
	p1p2, _, _ := basicTriangle.sides()
	Tform := rotateToVec(p1p2, Vec{Z: 1})
	// Tform.SetTranslate(Scale(-1, basicTriangle[0]))
	// Tform = MirrorXY()
	triangle2 := Tform.ApplyTriangle(basicTriangle)
	tri := triangleOutlines([]Triangle{basicTriangle}, lineColor("red"))

	grp.Add(three.NewAxesHelper(1))
	grp.Add(tri)
	_ = triangle2
	return grp
}
