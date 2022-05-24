//go:build js
// +build js

package main

import (
	"fmt"
	"math"
	"math/rand"
	"play/kdtree"
	"time"

	"github.com/soypat/three"
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
	return
	normieT := canalisTransform(normie).ApplyTriangle(normie)
	T := canalisTransform(worst)
	worstT := T.ApplyTriangle(worst)
	grp.Add(triangleOutlines([]Triangle{worst}, lineColor("gold")))
	grp.Add(triangleOutlines([]Triangle{worstT}, lineColor("green")))
	grp.Add(triangleOutlines([]Triangle{normie}, lineColor("cyan")))
	grp.Add(triangleOutlines([]Triangle{normieT}, lineColor("navy")))
	// return grp.
}
