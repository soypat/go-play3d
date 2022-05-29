package main

import "math"

type Tetra [4]Vec

// edges returns vectors for each of the edges of the tetrahedron.
func (t Tetra) edges() [6]Vec {
	return [6]Vec{
		Sub(t[1], t[0]), Sub(t[2], t[1]),
		Sub(t[0], t[2]), Sub(t[2], t[3]),
		Sub(t[3], t[0]), Sub(t[3], t[1]),
	}
}

func (t Tetra) aspect() float64 {
	e := t.longestEdge()
	alts := t.altitudes()
	amin, _, _ := sort3(alts[0], alts[1], alts[2])
	if alts[3] < amin {
		amin = alts[3]
	}
	return e / amin
}

func (t Tetra) longestEdge() float64 {
	edges := t.edges()
	_, _, e1 := sort3(Norm2(edges[0]), Norm2(edges[1]), Norm2(edges[2]))
	_, _, e2 := sort3(Norm2(edges[3]), Norm2(edges[4]), Norm2(edges[5]))
	if e1 > e2 {
		return math.Sqrt(e1)
	}
	return math.Sqrt(e2)
}

func (t Tetra) altitudes() (alt [4]float64) {
	for i := range t {
		j := (i + 1) % 4
		k := (i + 2) % 4
		l := (i + 3) % 4
		e1 := Sub(t[k], t[j])
		e2 := Sub(t[l], t[j])
		p := newPlane(t[l], Cross(e1, e2))
		alt[i] = p.distanceTo(t[i])
	}
	return alt
}

func (t Tetra) volume() float64 {
	const third = 1.0 / 3.0
	base := Triangle{t[0], t[1], t[2]}
	area := base.Area()
	height := base.plane().distanceTo(t[3])
	return third * area * height
}

type plane struct {
	// P is a point on the plane
	P Vec
	// n is the unit vector normal to the plane.
	n Vec
}

func newPlane(p, n Vec) plane {
	return plane{P: p, n: Unit(n)}
}

func (p plane) distanceTo(q Vec) float64 {
	v := Sub(q, p.P)
	return math.Abs(Dot(v, p.n))
}
