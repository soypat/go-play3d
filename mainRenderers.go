package main

import (
	"fmt"

	"github.com/soypat/three"
)

func lineColor(s string) three.MaterialParameters {
	return three.MaterialParameters{Color: three.NewColor("red")}
}

func triangleOutlines(t []Triangle, material three.MaterialParameters) three.Object3D {
	// We have 3 edges to draw per triangle
	// and need 2 points to define an edge
	// and each point is defined by 3 numbers (X,Y,Z data)
	edges := make([]float32, 2*3*3*len(t))
	for it, triangle := range t {
		triOffset := it * 12
		for i := range triangle {
			edgeI := triOffset + i*6
			v1, v2 := triangle[i], triangle[(i+1)%3]
			edges[edgeI] = float32(v1.X)
			edges[edgeI+1] = float32(v1.Y)
			edges[edgeI+2] = float32(v1.Z)
			edges[edgeI+3] = float32(v2.X)
			edges[edgeI+4] = float32(v2.Y)
			edges[edgeI+5] = float32(v2.Z)
			fmt.Println("triangle", it, "edge", i, v1, v2)
		}
	}
	fmt.Println(edges)
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(edges, 3))
	lines := three.NewLineSegments(geom, three.NewLineBasicMaterial(&material))
	return lines
}

func points(t []Vec, material three.MaterialParameters) three.Object3D {
	edges := make([]float32, 3*len(t))
	for it, v := range t {
		edges[it] = float32(v.X)
		edges[it+1] = float32(v.Y)
		edges[it+2] = float32(v.Z)
	}
	fmt.Println(edges)
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(edges, 3))
	lines := three.NewPoints(geom, three.NewPointsMaterial(material))
	return lines
}
