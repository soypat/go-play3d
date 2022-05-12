package main

import (
	"math"

	"gonum.org/v1/gonum/spatial/r2"
)

// line is an infinite 3D line
// defined by two points on the line.
type line [2]Vec

// interp takes a value between 0 and 1 to interpolate a point
// on the line.
//  interp(0) returns l[0]
//  interp(1) returns l[1]
func (l line) interp(t float64) Vec {
	lineDir := Sub(l[1], l[0])
	return Add(l[0], Scale(t, lineDir))
}

// equation returns the starting point of the line
// and the direction of the line. This can then be used
// to represent the line in vector, parametric or symmetric form.
func (l line) equation() (p0 Vec, n Vec) {
	// https://math.stackexchange.com/questions/404440/what-is-the-equation-for-a-3d-line
	return l[0], Sub(l[1], l[0])
}

// distance returns the minimum euclidean distance of point p
// to the line.
func (l line) distance(p Vec) float64 {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	num := Norm(Cross(Sub(p, l[0]), Sub(p, l[1])))
	return num / Norm(Sub(l[1], l[0]))
}

// closest returns the closest point on line l to point p.
func (l line) closest(p Vec) Vec {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	t := -Dot(Sub(l[0], p), Sub(l[1], p)) / Norm2(Sub(l[1], l[0]))
	return l.interp(t)
}

// distToLine returns distance vector from point to line.
// The integer returns 0 if closest to first vertex, 1 if closest
// to second vertex and 2 if closest to the line edge between vertices.
func distToLine(p r2.Vec, ln [2]r2.Vec) (r2.Vec, int) {
	lineDir := r2.Sub(ln[1], ln[0])
	perpendicular := r2.Vec{-lineDir.Y, lineDir.X}
	perpend2 := r2.Add(ln[1], perpendicular)
	e2 := edgeEquation(p, [2]r2.Vec{ln[1], perpend2})
	if e2 > 0 {
		return r2.Sub(p, ln[1]), 0
	}
	perpend1 := r2.Add(ln[0], perpendicular)
	e1 := edgeEquation(p, [2]r2.Vec{ln[0], perpend1})
	if e1 < 0 {
		return r2.Sub(p, ln[0]), 1
	}
	e3 := distToLineInfinite(p, ln) //edgeEquation(p, line)
	return r2.Scale(-e3, r2.Unit(perpendicular)), 2
}

// line passes through two points P1 = (x1, y1) and P2 = (x2, y2)
// then the distance of (x0, y0)
func distToLineInfinite(p r2.Vec, line [2]r2.Vec) float64 {
	// https://en.wikipedia.org/wiki/Distance_from_a_point_to_a_line
	p1 := line[0]
	p2 := line[1]
	num := math.Abs((p2.X-p1.X)*(p1.Y-p.Y) - (p1.X-p.X)*(p2.Y-p1.Y))
	return num / math.Hypot(p2.X-p1.X, p2.Y-p1.Y)
}

// edgeEquation returns a signed distance of a point to
// an infinite line defined by two points
// Edge equation for a line passing through (X,Y)
// with gradient dY/dX
// E ( x; y ) =(x-X)*dY - (y-Y)*dX
func edgeEquation(p r2.Vec, line [2]r2.Vec) float64 {
	dxy := r2.Sub(line[1], line[0])
	return (p.X-line[0].X)*dxy.Y - (p.Y-line[0].Y)*dxy.X
}
