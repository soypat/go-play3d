package main

import (
	"math"
	"play/kdtree"
)

var (
	_ kdtree.Interface[kdPoint, *kdTriangle] = kdMesh{}
	_ kdtree.Comparable[kdPoint]             = &kdTriangle{}
)

type kdMesh struct {
	tri []kdTriangle
}

func (m kdMesh) Index(i int) *kdTriangle { return &m.tri[i] }
func (m kdMesh) Len() int                { return len(m.tri) }
func (m kdMesh) Pivot(d kdtree.Dim) int {
	p := kdPlane{dim: int(d), triangles: m.tri}
	return kdtree.Partition(p, kdtree.MedianOfMedians(p))
}
func (m kdMesh) Bounds() (min, max kdPoint) {
	for _, t := range m.tri {
		lo, hi := t.Bounds()
		min = kdPoint{minElem(min.Vec, lo.Vec)}
		max = kdPoint{maxElem(max.Vec, hi.Vec)}
	}
	return min, max
}
func (m kdMesh) Slice(start, end int) kdtree.Interface[kdPoint, *kdTriangle] {
	return kdMesh{
		tri: m.tri[start:end],
	}
}

type triangleFeature int

const (
	featureV0 triangleFeature = iota
	featureV1
	featureV2
	featureE0
	featureE1
	featureE2
	featureFace
)

type kdTriangle struct {
	sdfTriangle
	lastDist    triangleFeature
	lastClosest Vec
}

type kdPoint struct {
	Vec
}

func (p kdPoint) Dims() int { return 3 }
func (p kdPoint) Component(d kdtree.Dim) float64 {
	switch d {
	case 0:
		return p.X
	case 1:
		return p.Y
	case 2:
		return p.Z
	}
	panic("unreachable")
	// return *(*float64)(unsafe.Add(unsafe.Pointer(&p), ))
}

// Given c = a.Compare(b, d):
//  c = a_d - b_d
func (t *kdTriangle) ComparePoint(p kdPoint, d kdtree.Dim) float64 {
	switch d {
	case 0:
		return t.C.X - p.X
	case 1:
		return t.C.Y - p.Y
	case 2:
		return t.C.Z - p.Z
	}
	panic("unreachable")
}

func (t *kdTriangle) Dims() int { return 3 }
func (t *kdTriangle) Bounds() (min, max kdPoint) {
	b := t.Triangle().Bounds()
	return kdPoint{b.Min}, kdPoint{b.Max}
}

func (t *kdTriangle) Distance(p kdPoint) float64 {
	pxy := t.T.ApplyPosition(p.Vec)
	txy := t.T.ApplyTriangle(t.Triangle())
	// get point on triangle closest to point
	ptxy, feat := closestOnTriangle2(pxy.lower(), txy.lower())
	t.lastDist = feat

	inv := t.T.Inverse()
	// Transform point on triangle back to 3D
	t.lastClosest = inv.ApplyPosition(Vec{ptxy.X, ptxy.Y, 0})
	return Norm2(Sub(p.Vec, t.lastClosest))
}

// Sign returns -1 or 1 depending on whether point is inside or outside
// mesh defined by the triangle and surrounding triangle normals.
// Must call Distance on the argument previous to calling Sign.
func (t *kdTriangle) CopySign(p Vec, dist float64) (signed float64) {
	if t.lastDist <= featureV2 {
		// Distance last called nearest to triangle vertex.
		vertex := t.m.vertices[t.sdfTriangle.Vertices[t.lastDist]]
		signed = Dot(vertex.N, Sub(p, vertex.V))
	} else if t.lastDist <= featureE2 {
		vertex1 := t.lastDist - 3
		edge := [2]int{t.sdfTriangle.Vertices[vertex1], t.sdfTriangle.Vertices[(vertex1+1)%3]}
		if edge[0] > edge[1] {
			edge[0], edge[1] = edge[1], edge[0]
		}
		norm := t.m.edgeNorm[edge]
		signed = Dot(norm, Sub(p, t.lastClosest))
	} else {
		signed = Dot(t.N, Sub(p, t.lastClosest))
	}
	return math.Copysign(dist, signed)
}

func (t *kdTriangle) Point() kdPoint { return kdPoint{t.C} }

type kdPlane struct {
	dim       int
	triangles []kdTriangle
}

func (p kdPlane) Less(i, j int) bool {
	return p.triangles[i].ComparePoint(p.triangles[j].Point(), kdtree.Dim(p.dim)) < 0
}
func (p kdPlane) Swap(i, j int) {
	p.triangles[i], p.triangles[j] = p.triangles[j], p.triangles[i]
}
func (p kdPlane) Len() int {
	return len(p.triangles)
}
func (p kdPlane) Slice(start, end int) kdtree.SortSlicer {
	p.triangles = p.triangles[start:end]
	return p
}
