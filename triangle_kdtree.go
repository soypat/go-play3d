package main

import (
	"play/kdtree"
)

var (
	_ kdtree.Interface[kdPoint, kdTriangle] = kdMesh{}
	_ kdtree.Comparable[kdPoint]            = kdTriangle{}
)

type kdMesh struct {
	tri []kdTriangle
}

func (m kdMesh) Index(i int) kdTriangle { return m.tri[i] }
func (m kdMesh) Len() int               { return len(m.tri) }
func (m kdMesh) Pivot(d kdtree.Dim) int {
	p := kdPlane{dim: int(d), triangles: m.tri}
	return kdtree.Partition(p, kdtree.MedianOfMedians(p))
}

func (m kdMesh) Slice(start, end int) kdtree.Interface[kdPoint, kdTriangle] {
	return kdMesh{
		tri: m.tri[start:end],
	}
}

type kdTriangle struct {
	sdfTriangle
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
func (t kdTriangle) ComparePoint(p kdPoint, d kdtree.Dim) float64 {
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

func (t kdTriangle) Dims() int { return 3 }

func (t kdTriangle) Distance(p kdPoint) float64 {
	pxy := t.T.ApplyPosition(p.Vec)
	txy := t.T.ApplyTriangle(t.Triangle())
	// get point on triangle closest to point
	ptxy := closestOnTriangle2(pxy.lower(), txy.lower())
	inv := t.T.Inverse()
	// Transform point on triangle back to 3D
	pxy = inv.ApplyPosition(Vec{ptxy.X, ptxy.Y, 0})
	return Norm2(Sub(p.Vec, pxy))
}

func (t kdTriangle) Point() kdPoint { return kdPoint{t.C} }

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
