// Copyright ©2022 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "math"

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

// Degenerate returns true if triangle's vertices are collinear to within
// a certain tolerance.
func (t Triangle) Degenerate(tol float64) bool {
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
