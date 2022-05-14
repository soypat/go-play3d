package main

import (
	"math"
	kdt "play/kdtree"
	"unsafe"

	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/kdtree"
	"gonum.org/v1/gonum/spatial/r3"
)

type gonumSDFMesh struct {
	tree kdtree.Tree
}

func (s gonumSDFMesh) Evaluate(q r3.Vec) float64 {
	p := Vec(q)
	triface, dist2 := s.tree.Nearest(kdgTriangle{
		T: &kdTriangle{
			sdfTriangle: sdfTriangle{C: Vec(q)},
		},
	})
	tri := triface.(kdgTriangle)
	return tri.T.CopySign(p, math.Sqrt(dist2))
}

func (s gonumSDFMesh) Bounds() r3.Box {
	min := s.tree.Root.Bounding.Min.(kdgTriangle)
	max := s.tree.Root.Bounding.Max.(kdgTriangle)
	return r3.Box{
		Min: r3.Vec(min.T.C),
		Max: r3.Vec(max.T.C),
	}
}

func NewGonumSDFMesh(model []render.Triangle3) gonumSDFMesh {
	triangles := *(*[]Triangle)(unsafe.Pointer(&model))
	m := newMesh(triangles, 1e-4)
	kdtri := make([]kdTriangle, len(m.triangles))
	for i := range m.triangles {
		kdtri[i] = kdTriangle{sdfTriangle: m.triangles[i]}
	}
	tree := kdtree.New(kdgTree{M: kdMesh{tri: kdtri}}, true)
	return gonumSDFMesh{
		tree: *tree,
	}
}

var (
	_ kdtree.Comparable = kdgTriangle{}
	_ kdtree.Interface  = (*kdgTree)(nil)
)

type kdgTriangle struct {
	T *kdTriangle
}

type kdgTree struct {
	M kdMesh
}

// func (t kdgTriangle) Bounds() *kdtree.Bounding {
// 	box:=t.T.Triangle().Bounds()
// 	return &kdtree.Bounding{
// 		Min: &kdgTriangle{T:&kdTriangle{sdfTriangle: }} ,
// 	}
// }

func (t kdgTriangle) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	q := c.(kdgTriangle)
	return t.T.ComparePoint(q.T.Point(), kdt.Dim(d))
}

// Dims returns the number of dimensions described in the Comparable.
func (t kdgTriangle) Dims() int { return 3 }

// Distance returns the squared Euclidean distance between the receiver and
// the parameter.
func (t kdgTriangle) Distance(c kdtree.Comparable) float64 {
	q := c.(kdgTriangle)
	if t.T.T == nil {
		if q.T.T == nil {
			panic("nothing initialized")
		}
		return q.T.Distance(kdPoint{Vec: t.T.C})
		// panic("receiver not initialized")
	}

	return t.T.Distance(q.T.Point())
}

// Index returns the ith element of the list of points.
func (tr kdgTree) Index(i int) kdtree.Comparable { return kdgTriangle{T: tr.M.Index(i)} }

// Len returns the length of the list.
func (tr kdgTree) Len() int { return tr.M.Len() }

// Pivot partitions the list based on the dimension specified.
func (tr kdgTree) Pivot(d kdtree.Dim) int { return tr.M.Pivot(kdt.Dim(d)) }

// Slice returns a slice of the list using zero-based half
// open indexing equivalent to built-in slice indexing.
func (tr kdgTree) Slice(start, end int) kdtree.Interface {
	sli := tr.M.Slice(start, end)
	return &kdgTree{M: sli.(kdMesh)}
}

func (tr kdgTree) Bounds() *kdtree.Bounding {
	min, max := tr.M.Bounds()
	return &kdtree.Bounding{
		Min: kdgTriangle{&kdTriangle{sdfTriangle: sdfTriangle{
			C: min.Vec,
		}}},
		Max: kdgTriangle{&kdTriangle{sdfTriangle: sdfTriangle{
			C: max.Vec,
		}}},
	}
}

func (tr kdgTriangle) Bounds() *kdtree.Bounding {
	min, max := tr.T.Bounds()
	return &kdtree.Bounding{
		Min: kdgTriangle{&kdTriangle{sdfTriangle: sdfTriangle{
			C: min.Vec,
		}}},
		Max: kdgTriangle{&kdTriangle{sdfTriangle: sdfTriangle{
			C: max.Vec,
		}}},
	}
}
