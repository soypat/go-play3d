//go:build js

package main

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3/must3"
	"github.com/soypat/three"
	"gonum.org/v1/gonum/spatial/r3"
)

func addObjects(grp three.Group) {
	var s sdf.SDF3
	grp.Add(three.NewAxesHelper(1))
	const quality = 150
	sc := must3.Cylinder(2, .5, .1)
	s = sc
	s = sdftransform{
		sdf: sc,
		t:   warpit(),
	}
	s = sdf.Transform3D(s, sdf.Rotate3d(r3.Vec{Z: 1}, math.Pi))
	grp.Add(sdf3Obj(s, quality, "red", .5))
	sbb := s.Bounds()
	bb := Box{Vec(sbb.Min), Vec(sbb.Max)}

	mainBox := CenteredBox(bb.Center(), bb.Scale(Vec{1.1, 1.1, 1.1}).Size())
	grp.Add(boxesObj([]Box{mainBox}, lineColor("green")))
	return
	boxes := boxDivide(mainBox, 10)
	norms := make([][2]Vec, len(boxes))
	for i, box := range boxes {
		const tol = 1e-3
		c := box.Center()
		curvature := sdfCurvature(s, c, tol)
		norms[i] = [2]Vec{c, Add(c, Scale(curvature*1e5, sdfNormal(s, c, tol)))}
	}
	mesh := maketmesh(mainBox, .1)
	// grp.Add(boxesObj(boxes, lineColor("blue")))
	// grp.Add(linesObj(norms, lineColor("gold")))
	// grp.Add(boxesObj(mesh.boxes(), lineColor("green")))
	nodes, tetras := mesh.meshTetraBCC()
	eval := func(v Vec) float64 { return s.Evaluate(r3.Vec(v)) }
	newtetras := make([][4]int, 0, len(tetras))

	for _, tetra := range tetras {
		nd := Tetra{nodes[tetra[0]], nodes[tetra[1]], nodes[tetra[2]], nodes[tetra[3]]}
		// aspect := nd.longestEdge()
		// evals := [4]float64{eval(nd[0]), eval(nd[1]), eval(nd[2]), eval(nd[3])}
		if eval(nd[0]) < 0 || eval(nd[1]) < 0 || eval(nd[2]) < 0 || eval(nd[3]) < 0 {
			newtetras = append(newtetras, tetra)
		}
	}
	omesh := newOptimesh(nodes, newtetras)
	omesh.foreach(func(i int, on *onode) {
		// Lapla
	})
	// grp.Add(pointsObj(nodes, pointColor("cyan")))
	grp.Add(triangleMesh(tetraTriangles(nodes, newtetras), phongMaterial("orange", 0.5)))
}

type sdftransform struct {
	sdf sdf.SDF3
	t   Transformer
}

func (t sdftransform) Evaluate(v r3.Vec) float64 {
	return t.sdf.Evaluate(r3.Vec(t.t.Transform(Vec(v))))
}

func (t sdftransform) Bounds() r3.Box {
	var bbt Box
	bbr3 := t.sdf.Bounds()
	bb := Box{Min: Vec(bbr3.Min), Max: Vec(bbr3.Max)}
	bb = bb.Scale(Vec{1.05, 1.05, 1.05})
	if a, ok := t.t.(Affine); ok {
		bbt = a.ApplyBox(bb)
	} else {
		bbt = bb.TransformBox(t.t)
	}
	return r3.Box{
		Min: r3.Vec(bbt.Min),
		Max: r3.Vec(bbt.Max),
	}
}
