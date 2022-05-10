package main

import (
	"math"

	"github.com/soypat/three"
)

func makeObjects() three.Object3D {
	var tri1, tri2 three.Object3D
	grp := three.NewGroup()
	goldie := Triangle{Vec{0, 0, 0}, Vec{1, 0, 0}, Vec{0, 1, 0}} // goldie is our base working triangle
	goldDisplacement := Vec{Z: 1}
	goldie = goldie.Add(goldDisplacement)
	Trot := Rotate3d(Vec{X: 1, Y: 1, Z: 1}, 1)
	goldie = Trot.ApplyTriangle(goldie)
	Tform := jonesTransform(goldie)
	transformed := Tform.ApplyTriangle(goldie)
	const plen = 10
	points := PointCloud(plen, 2)
	transformedPoints := make([]Vec, plen)
	for i := range points {
		transformedPoints[i] = Tform.ApplyPosition(points[i])
		transformedPoints[i].X = 0
	}

	tri1 = triangleOutlines([]Triangle{goldie}, lineColor("gold"))
	tri2 = triangleOutlines([]Triangle{transformed}, lineColor("fuchsia"))
	// pts := points(PointCloud(100, 2), pointColor("red"))
	grp.Add(three.NewAxesHelper(1))
	grp.Add(tri1)
	grp.Add(tri2)
	grp.Add(pointsObj(points, pointColor("white")))
	grp.Add(pointsObj(transformedPoints, pointColor("red")))
	_, _ = tri1, tri2
	return grp
}

// Returns a transformation for a triangle so that:
//  - the triangle's first edge (t[0],t[1]) is on the Z axis
//  - the triangle's first vertex is at the origin
//  - the triangle's last vertex is in the YZ plane.
func jonesTransform(t Triangle) Transform {
	// Mark W. Jones "3D Distance from a Point to a Triangle"
	// Department of Computer Science, University of Wales Swansea
	p1p2, _, _ := t.sides()
	Tform := rotateToVec(p1p2, Vec{Z: 1})
	Tdis := Translate(Scale(-1, t[0]))
	Tform = Tform.Mul(Tdis)
	t = Tform.ApplyTriangle(t)
	// rotate third point so that it is on yz plane
	alpha := math.Acos(Cos(Vec{Y: 1}, t[2]))
	Trot := Rotate3d(Vec{Z: 1}, -alpha)
	Tform = Trot.Mul(Tform)
	return Tform
}
