//go:build js

package main

import (
	"fmt"
	"math"
	"math/rand"
	"play/kdtree"
	"time"

	"github.com/soypat/three"
	"gonum.org/v1/gonum/num/quat"
)

func makeObjects() three.Object3D {
	grp := three.NewGroup()
	rnd := rand.New(rand.NewSource(time.Now().UnixMilli()))
	normie := Triangle{{}, {X: 0.5, Y: 1}, {X: 1}}
	worst := Triangle{{X: 0.30901699437494745, Y: 0.5000000000000001, Z: -0.8090169943749475}, {X: 0.16245984811645314, Y: 0.2628655560595668, Z: -0.9510565162951533}, {X: 0, Y: 0.5257311121191337, Z: -0.85065080835204}}
	// m4 := three.NewMatrix4()
	normieT := jonesTransform(normie).ApplyTriangle(normie)
	T := jonesTransform(worst)
	worstT := T.ApplyTriangle(worst)
	grp.Add(triangleOutlines([]Triangle{worst}, lineColor("gold")))
	grp.Add(triangleOutlines([]Triangle{worstT}, lineColor("green")))
	grp.Add(triangleOutlines([]Triangle{normie}, lineColor("cyan")))
	grp.Add(triangleOutlines([]Triangle{normieT}, lineColor("navy")))
	grp.Add(three.NewAxesHelper(0.5))
	fmt.Println("worstT:", worstT)
	return grp
	var tri1, tri2 three.Object3D
	tris := icosphere(2)

	// tris := genShape()

	m := newMesh(tris, 1e-4)
	// Vertex pseudo norms calculation
	Nvs := make([][2]Vec, len(m.vertices))
	for i, v := range m.vertices {
		Nvs[i] = [2]Vec{
			v.V,
			Add(v.V, v.N),
		}
	}
	// Edge pseudo norm calc
	Nes := make([][2]Vec, len(m.edgeNorm))
	i := 0
	for verts, norm := range m.edgeNorm {
		v1 := m.vertices[verts[0]]
		v2 := m.vertices[verts[1]]
		midpoint := Scale(0.5, Add(v1.V, v2.V)) // edge midpoint
		Nes[i] = [2]Vec{
			midpoint,
			Add(midpoint, norm),
		}
		i++
	}
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
	return grp
}

func slerpQ(x float64, a, b quat.Number) quat.Number {
	cost := dotQ(a, b)
	theta := math.Acos(cost)
	sint := math.Sqrt(1 - cost*cost)
	sinxt := math.Sin(x*theta) / sint
	sinxtm1 := math.Sin((1-x)*theta) / sint
	result := quat.Scale(sinxtm1, a)
	return quat.Add(result, quat.Scale(sinxt, b))
}

func dotQ(a, b quat.Number) float64 {
	return a.Real*b.Real + a.Imag*b.Imag + a.Jmag*b.Jmag + a.Kmag*b.Kmag
}
