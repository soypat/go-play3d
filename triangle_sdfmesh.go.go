package main

import (
	"math"
	"play/kdtree"
	"unsafe"

	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

func NewSDFMesh(model []render.Triangle3) sdfMesh {
	triangles := *(*[]Triangle)(unsafe.Pointer(&model))
	m := newMesh(triangles, 1e-4)
	kdtri := make([]kdTriangle, len(m.triangles))
	for i := range m.triangles {
		kdtri[i] = kdTriangle{sdfTriangle: m.triangles[i]}
	}
	tree := kdtree.New[kdPoint, *kdTriangle](kdMesh{tri: kdtri}, true)
	return sdfMesh{
		tree: *tree,
	}
}

type sdfMesh struct {
	tree kdtree.Tree[kdPoint, *kdTriangle]
}

func (s sdfMesh) Evaluate(q r3.Vec) float64 {
	p := Vec(q)
	tri, dist2 := s.tree.Nearest(kdPoint{p})
	return tri.CopySign(p, math.Sqrt(dist2))
}

func (s sdfMesh) Bounds() r3.Box {
	return r3.Box{
		Min: r3.Vec(s.tree.Root.Bounding.Min.Vec),
		Max: r3.Vec(s.tree.Root.Bounding.Max.Vec),
	}
}
