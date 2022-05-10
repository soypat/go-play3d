package main

import (
	"github.com/soypat/three"
)

func lineColor(s string) three.MaterialParameters {
	// randThickness := rand.Float64()
	return three.MaterialParameters{
		Color:     three.NewColor(s),
		LineWidth: 2,
		// Size:  10,
	}
}
func pointColor(s string) three.MaterialParameters {
	// randThickness := rand.Float64()
	return three.MaterialParameters{
		Color: three.NewColor(s),
		Size:  size / 100,
		// Size:  10,
	}
}

func triangleOutlines(t []Triangle, material three.MaterialParameters) three.Object3D {
	// We have 3 edges to draw per triangle
	// and need 2 points to define an edge
	// and each point is defined by 3 numbers (X,Y,Z data)
	edges := make([]float32, 2*3*3*len(t))
	for it, triangle := range t {
		triOffset := it * 18
		for i := range triangle {
			edgeI := triOffset + i*6
			v1, v2 := triangle[i], triangle[(i+1)%3]
			edges[edgeI] = float32(v1.X)
			edges[edgeI+1] = float32(v1.Y)
			edges[edgeI+2] = float32(v1.Z)
			edges[edgeI+3] = float32(v2.X)
			edges[edgeI+4] = float32(v2.Y)
			edges[edgeI+5] = float32(v2.Z)
		}
	}
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(edges, 3))
	lines := three.NewLineSegments(geom, three.NewLineBasicMaterial(&material))
	return lines
}

func pointsObj(p []Vec, material three.MaterialParameters) three.Object3D {
	points := make([]float32, 3*len(p))
	for it, vert := range p {
		points[it*3] = float32(vert.X)
		points[it*3+1] = float32(vert.Y)
		points[it*3+2] = float32(vert.Z)
	}
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(points, 3))
	lines := three.NewPoints(geom, three.NewPointsMaterial(material))
	return lines
}
