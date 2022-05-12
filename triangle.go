// Copyright ©2022 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"

	"gonum.org/v1/gonum/spatial/r2"
)

// Triangle represents a triangle in 3D space and
// is composed by 3 vectors corresponding to the position
// of each of the vertices. Ordering of these vertices
// decides the "normal" direction.
// Inverting ordering of two vertices inverts the resulting direction.
type Triangle [3]Vec

// Normal returns the vector with direction
// perpendicular to the Triangle's face. The ordering
// of the triangle vertices decides the normal's resulting
// direction. Normal is not guaranteed to return a unit-norm vector.
func (t Triangle) Normal() Vec {
	s1, s2, _ := t.sides()
	return Cross(s1, s2)
}

func (t Triangle) Circumcenter() Vec {
	return Add(t[0], t.toCircumcenter())
}

func (t Triangle) Circumradius() float64 {
	return Norm2(t.toCircumcenter())
}

func (t Triangle) Bounds() Box {
	return Box{
		Min: minElem(t[0], minElem(t[1], t[2])),
		Max: maxElem(t[0], maxElem(t[1], t[2])),
	}
}

func (t Triangle) toCircumcenter() Vec {
	// from https://gamedev.stackexchange.com/questions/60630/how-do-i-find-the-circumcenter-of-a-triangle-in-3d
	// N = ∥c−a∥2[(b−a)×(c−a)]×(b−a)+∥b−a∥2[(c−a)×(b−a)]×(c−a)
	// aC = N / 2∥(b−a)×(c−a)∥2
	ac := Sub(t[2], t[0])
	ab := Sub(t[1], t[0])
	abXac := Cross(ab, ac)
	num := Add(Scale(Norm2(ac), Cross(abXac, ab)), Scale(Norm2(ab), Cross(ac, abXac)))
	return Scale(1/(2*Norm2(abXac)), num)
}

// Degenerate returns true if triangle's vertices are collinear to within
// a certain tolerance.
func (t Triangle) Degenerate(tol float64) bool {
	// a triangle whose circumradius-to-shortest edge ratio is greater than ½ is said to be skinny
	// cr := t.Circumradius()
	// a, _, _ := t.orderedLengths()
	// ratio := cr / a
	// return ratio > 0.5

	// TODO(soypat): Must be a better way to do this.
	// https://stackoverflow.com/questions/33037449/given-three-side-of-a-triangle-how-can-i-define-whether-it-is-a-degenerate-tria
	a, b, c := t.orderedLengths()
	// The sum of the shorter sides of a valid triangle is always
	// larger than the longest of the sides.
	return a+b < c+tol
}

// Area returns the surface area of the triangle.
func (t Triangle) Area() float64 {
	// William M. Kahan (24 March 2000). "Miscalculating Area and Angles of a Needle-like Triangle"
	a, b, c := t.orderedLengths()
	A := (c + (b + a)) * (a - (c - b))
	A *= (a + (c - b)) * (c + (b - a))
	return math.Sqrt(A) / 4
}

func (t Triangle) Closest(p Vec) Vec {
	Tform := jonesTransform(t)
	pxy := Tform.ApplyPosition(p)
	txy := Tform.ApplyTriangle(t)
	// get point on triangle closest to point
	ptxy, _ := closestOnTriangle2(pxy.lower(), txy.lower())
	inv := Tform.Inverse()
	return inv.ApplyPosition(Vec{ptxy.X, ptxy.Y, 0})
}

// Centroid returns the intersection of the three medians of the triangle
// as a point in space.
func (t Triangle) Centroid() Vec {
	return Scale(1./3., Add(Add(t[0], t[1]), t[2]))
}

func (t Triangle) Edges() (edges [3][2]Vec) {
	for i := range t {
		edges[i][0] = t[i]
		edges[i][1] = t[(i+1)%3]
	}
	return edges
}

// sides returns vectors for each of the sides of the Triangle.
func (t Triangle) sides() (Vec, Vec, Vec) {
	return Sub(t[1], t[0]), Sub(t[2], t[1]), Sub(t[0], t[2])
}

func (t Triangle) Add(v Vec) Triangle {
	return Triangle{Add(t[0], v), Add(t[1], v), Add(t[2], v)}
}

// orderedLengths returns the lengths of the sides of the triangle such that
// a ≤ b ≤ c.
func (t Triangle) orderedLengths() (a, b, c float64) {
	s1, s2, s3 := t.sides()
	l1 := Norm(s1)
	l2 := Norm(s2)
	l3 := Norm(s3)
	a = math.Min(math.Min(l1, l2), l3)
	c = math.Max(math.Max(l1, l2), l3)
	// Find which length is neither max nor minimum.
	switch {
	case l1 != a && l1 != c:
		b = l1
	case l2 != a && l2 != c:
		b = l2
	default:
		b = l3
	}
	return a, b, c
}

func (t Triangle) lower() [3]r2.Vec {
	return [3]r2.Vec{
		{X: t[0].X, Y: t[0].Y},
		{X: t[1].X, Y: t[1].Y},
		{X: t[2].X, Y: t[2].Y},
	}
}

// Returns a transformation for a triangle so that:
//  - the triangle's first edge (t_0,t_1) is on the X axis
//  - the triangle's first vertex t_0 is at the origin
//  - the triangle's last vertex t_2 is in the XY plane.
func jonesTransform(t Triangle) Transform {
	// Mark W. Jones "3D Distance from a Point to a Triangle"
	// Department of Computer Science, University of Wales Swansea
	p1p2, _, _ := t.sides()
	Tform := rotateToVec(p1p2, Vec{X: 1})
	Tdis := Translate(Scale(-1, t[0]))
	Tform = Tform.Mul(Tdis)
	t = Tform.ApplyTriangle(t)
	// rotate third point so that it is on yz plane
	t[2].X = 0 // eliminate X component.
	alpha := math.Acos(Cos(Vec{Y: 1}, t[2]))
	Trot := Rotate3d(Vec{X: 1}, -alpha)
	Tform = Trot.Mul(Tform)
	return Tform
}

func closestOnTriangle2(p r2.Vec, tri [3]r2.Vec) (pointOnTriangle r2.Vec, feature triangleFeature) {
	if inTriangle(p, tri) {
		return p, featureFace
	}
	minDist := math.MaxFloat64
	for j := range tri {
		edge := [2]r2.Vec{{X: tri[j].X, Y: tri[j].Y}, {X: tri[(j+1)%3].X, Y: tri[(j+1)%3].Y}}
		distance, gotFeat := distToLine(p, edge)
		d2 := r2.Norm2(distance)
		if d2 < minDist {
			if gotFeat < 2 {
				feature = triangleFeature(j+gotFeat) % 3
			} else {
				feature = featureE0 + triangleFeature(j)%3
			}
			minDist = d2
			pointOnTriangle = r2.Sub(p, distance)
		}
	}
	return pointOnTriangle, feature
}
