//go:build js
// +build js

package main

import (
	"fmt"
	"math"
	"math/rand"
	"play/kdtree"
	"time"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3/must3"
	"github.com/soypat/three"
	"gonum.org/v1/gonum/spatial/r3"
)

func ExampleClosestTriangleIcosahedron() {
	grp := three.NewGroup()
	grp.Add(three.NewAxesHelper(0.5))
	rnd := rand.New(rand.NewSource(time.Now().UnixMilli()))
	var tri1, tri2 three.Object3D
	tris := icosphere(2)

	m := newMesh(tris, 1e-4)

	kdtri := make([]kdTriangle, len(m.triangles))
	for i := range m.triangles {
		kdtri[i] = kdTriangle{sdfTriangle: m.triangles[i]}
	}
	tree := kdtree.New[kdPoint, *kdTriangle](kdMesh{tri: kdtri}, true)

	lookup := Scale(1.1, Unit(Vec{X: rnd.Float64(), Y: rnd.Float64(), Z: rnd.Float64()}))
	nearbySDF, dist2 := tree.Nearest(kdPoint{lookup})
	// tree.NearestSet(kdKeeper{})
	nearby := nearbySDF.Triangle()
	fmt.Println(lookup)
	closePoint := nearby.Closest(lookup)
	minDist := math.Sqrt(dist2)
	nearest := nearby
	for _, t := range tris {
		close := t.Closest(lookup)
		gotDist := Norm(Sub(lookup, close))
		if gotDist < minDist {
			nearest = t
			minDist = gotDist
		}
	}
	grp.Add(three.NewAxesHelper(0.5))
	grp.Add(triangleOutlines([]Triangle{nearby.Add(Scale(0.05, nearbySDF.N))}, lineColor("gold")))
	grp.Add(triangleOutlines([]Triangle{nearest.Add(Scale(0.025, nearest.Normal()))}, lineColor("pink")))
	grp.Add(pointsObj([]Vec{lookup}, pointColor("red")))
	grp.Add(pointsObj([]Vec{closePoint}, pointColor("gold")))
	// grp.Add(linesObj(Nes, lineColor("gold")))
	grp.Add(boxesObj([]Box{nearby.Bounds()}, lineColor("white")))

	// grp.Add(triangleOutlines(tris, lineColor("blue")))
	grp.Add(triangleOutlines(m.Triangles(), lineColor("darkgreen")))
	_, _ = tri1, tri2
}

func ExampleClosestPointOnTriangle() {
	grp := three.NewGroup()
	rnd := rand.New(rand.NewSource(1))
	normie := Triangle{{}, {X: 0.5, Y: 1}, {X: 1}}
	worst := Triangle{{X: 0.30901699437494745, Y: 0.5000000000000001, Z: -0.8090169943749475}, {X: 0.16245984811645314, Y: 0.2628655560595668, Z: -0.9510565162951533}, {X: 0, Y: 0.5257311121191337, Z: -0.85065080835204}}
	tris := []Triangle{normie}
	for _, t := range tris {
		T := canalisTransform(t)
		tT := T.ApplyTriangle(t)
		point := randomVec(1, rnd)
		pointT := T.Transform(point)
		onTri2, _ := closestOnTriangle2(pointT.lower(), tT.lower())
		onTri3 := t.Closest(point)
		fmt.Println(onTri2, onTri3)
		grp.Add(pointsObj([]Vec{point}, pointColor("aliceblue")))
		grp.Add(triangleOutlines([]Triangle{t}, lineColor("azure")))
		grp.Add(pointsObj([]Vec{{X: onTri2.X, Y: onTri2.Y}}, pointColor("fuchsia")))
		grp.Add(pointsObj([]Vec{onTri3}, pointColor("gold")))

	}
	normieT := canalisTransform(normie).ApplyTriangle(normie)
	T := canalisTransform(worst)
	worstT := T.ApplyTriangle(worst)
	grp.Add(triangleOutlines([]Triangle{worst}, lineColor("gold")))
	grp.Add(triangleOutlines([]Triangle{worstT}, lineColor("green")))
	grp.Add(triangleOutlines([]Triangle{normie}, lineColor("cyan")))
	grp.Add(triangleOutlines([]Triangle{normieT}, lineColor("navy")))
	// return grp.
}

func ExampleMesh() {
	grp := three.NewGroup()
	var s sdf.SDF3
	grp.Add(three.NewAxesHelper(1))
	const quality = 150
	sc := must3.Cylinder(2, .5, .1)
	s = sc
	s = sdftransform{
		sdf: sc,
		inv: Warp{XY: .1}.shearAffine().Inv(),
	}
	s = sdf.Transform3D(s, sdf.Rotate3D(r3.Vec{Z: 1}, math.Pi))
	s = csgBasic()
	sbb := s.Bounds()
	bb := Box{Vec(sbb.Min), Vec(sbb.Max)}

	mainBox := CenteredBox(bb.Center(), bb.Scale(Vec{1.1, 1.1, 1.1}).Size())
	grp.Add(boxesObj([]Box{mainBox}, lineColor("green")))
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
	eval := func(v Vec) float64 { return s.Evaluate(r3.Vec(v)) }
	nodes, tetras := mesh.meshTetraBCC(eval)

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
	for iter := 1; iter <= 6; iter++ {
		scaler := float64(iter) / 6.0
		boundary := make(map[int]struct{})
		omesh.foreach(func(i int, on *onode) {
			d := s.Evaluate(r3.Vec(on.c))
			if d > 0 {
				boundary[i] = struct{}{}
				n := Scale(scaler*d, Unit(sdfNormal(s, on.c, 1e-6)))
				on.c = Sub(on.c, n)
			}
			nodes[i] = on.c
		})
		omesh.foreach(func(i int, on *onode) {
			if _, ok := boundary[i]; ok {
				// don't smooth boundary nodes.
				return
			}
			var sum Vec
			for _, conn := range on.connectivity {
				vi := omesh.nodes[conn].c
				sum = Add(sum, vi)
			}
			on.c = Scale(1/float64(len(on.connectivity)), sum)
		})
	}
	grp.Add(triangleMesh(tetraTriangles(omesh.nodePositions(), omesh.tetras), phongMaterial("orange", 0.5)))
}
