package main

import (
	"math"

	"github.com/soypat/three"
	"gonum.org/v1/gonum/spatial/r2"
)

func makeObjects() three.Object3D {
	var tri1, tri2 three.Object3D
	grp := three.NewGroup()
	goldie := Triangle{Vec{0, 0, 0}, Vec{2, 0, 0}, Vec{0.5, 1, 0}} // goldie is our base working triangle
	goldDisplacement := Vec{Z: 1}
	goldie = goldie.Add(goldDisplacement)
	Trot := Rotate3d(Vec{X: 1, Y: 1, Z: 1}, 0)
	goldie = Trot.ApplyTriangle(goldie)
	Tform := jonesTransform(goldie)
	transformed := Tform.ApplyTriangle(goldie)
	const plen = 80
	points := PointCloud(plen, goldie.Circumradius()*2, goldie.Centroid())
	transformedPoints := make([]Vec, plen)
	transformedPointDist := make([][2]Vec, plen)
	pointDist := make([][2]Vec, plen)
	for i := range points {
		transformedPoints[i] = Tform.ApplyPosition(points[i])
		pxy := r2.Vec{X: transformedPoints[i].X, Y: transformedPoints[i].Y}
		transformedPoints[i].Z = 0
		tri2d := transformed.lower()
		pt := closestOnTriangle2(pxy, tri2d)
		transformedPointDist[i] = [2]Vec{transformedPoints[i], Vec{X: pt.X, Y: pt.Y}}
		pointDist[i] = [2]Vec{points[i], goldie.Closest(points[i])}
	}

	tri1 = triangleOutlines([]Triangle{goldie}, lineColor("gold"))
	tri2 = triangleOutlines([]Triangle{transformed}, lineColor("fuchsia"))
	// pts := points(PointCloud(100, 2), pointColor("red"))

	grp.Add(three.NewAxesHelper(1))
	grp.Add(tri1)
	grp.Add(tri2)
	grp.Add(pointsObj(points, pointColor("white")))
	grp.Add(pointsObj(transformedPoints, pointColor("red")))
	grp.Add(linesObj(transformedPointDist, lineColor("white")))
	grp.Add(pointsObj([]Vec{goldie.Circumcenter()}, pointColor("gold")))
	grp.Add(linesObj(pointDist, lineColor("beige")))
	_, _ = tri1, tri2
	return grp
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

func closestOnTriangle2(p r2.Vec, tri [3]r2.Vec) (pointOnTriangle r2.Vec) {
	if inTriangle(p, tri) {
		return p
	}
	minDist := math.MaxFloat64
	for j := range tri {
		edge := [2]r2.Vec{{X: tri[j].X, Y: tri[j].Y}, {X: tri[(j+1)%3].X, Y: tri[(j+1)%3].Y}}
		distance, _ := distToLine(p, edge)
		d2 := r2.Norm2(distance)
		if d2 < minDist {
			minDist = d2
			pointOnTriangle = r2.Sub(p, distance)
		}
	}
	return pointOnTriangle
}

// distToLine returns distance vector from point to line.
// The boolean return parameter is set to true if the point
// is closest to a vertex of the line.
func distToLine(p r2.Vec, ln [2]r2.Vec) (r2.Vec, bool) {

	lineDir := r2.Sub(ln[1], ln[0])
	perpendicular := r2.Vec{-lineDir.Y, lineDir.X}

	perpend2 := r2.Add(ln[1], perpendicular)
	e2 := edgeEquation(p, [2]r2.Vec{ln[1], perpend2})
	if e2 > 0 {
		return r2.Sub(p, ln[1]), true
	}
	perpend1 := r2.Add(ln[0], perpendicular)
	e1 := edgeEquation(p, [2]r2.Vec{ln[0], perpend1})
	if e1 < 0 {
		return r2.Sub(p, ln[0]), true
	}

	e3 := distToLineInfinite(p, ln) //edgeEquation(p, line)
	return r2.Scale(-e3, r2.Unit(perpendicular)), false
}

// line passes through two points P1 = (x1, y1) and P2 = (x2, y2)
// then the distance of (x0, y0)
func distToLineInfinite(p r2.Vec, line [2]r2.Vec) float64 {
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
