//go:build js

package main

import (
	"math/rand"
	"play/kdtree"
	"time"

	"github.com/soypat/three"
)

func makeObjects() three.Object3D {
	rng := rand.New(rand.NewSource(time.Now().UnixMilli()))
	var tri1, tri2 three.Object3D
	grp := three.NewGroup()
	tris := genShape()

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

	lookup := Scale(1.1, Unit(Vec{X: rng.Float64(), Y: rng.Float64(), Z: rng.Float64()}))
	nearbySDF, _ := tree.Nearest(kdPoint{lookup})
	nearby := nearbySDF.Triangle()
	grp.Add(triangleOutlines(icosphere(5), lineColor("green")))
	grp.Add(three.NewAxesHelper(0.5))
	return grp
	grp.Add(triangleOutlines([]Triangle{nearby.Add(Scale(0.05, nearbySDF.N))}, lineColor("gold")))
	grp.Add(pointsObj([]Vec{lookup}, pointColor("red")))
	// grp.Add(linesObj(Nes, lineColor("gold")))
	grp.Add(boxesObj([]Box{nearby.Bounds()}, lineColor("white")))

	// grp.Add(triangleOutlines(tris, lineColor("blue")))
	grp.Add(triangleOutlines(m.Triangles(), lineColor("darkgreen")))
	_, _ = tri1, tri2
	return grp
}
