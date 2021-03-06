//go:build js

package main

import (
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/render"
	"github.com/soypat/three"
)

func phongMaterial(color string, opacity float64) three.MeshPhongMaterial {
	transparent := opacity < 1
	return three.NewMeshPhongMaterial(&three.MaterialParameters{
		Color:       three.NewColor(color),
		Transparent: transparent,
		Opacity:     opacity,
		Specular:    three.NewColor("grey"),
		Shading:     three.SmoothShading,
	})
}

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
		Size:  size / 1000,
		// Size:  10,
	}
}

func triangleNormalsObj(length float64, t []Triangle, material three.MaterialParameters) three.Object3D {
	norms := make([][2]Vec, len(t))
	for i := range t {
		c := t[i].Centroid()
		norms[i][0] = c
		norms[i][1] = Add(c, Scale(length, Unit(t[i].Normal())))
	}
	return linesObj(norms, material)
}

func triangleOutlineGeometry(t []Triangle) three.BufferGeometry {
	// We have 3 edges to draw per triangle
	// and need 2 points to define an edge
	// and each point is defined by 3 numbers (X,Y,Z data)
	e32 := make([]float32, 2*3*3*len(t))
	for it, triangle := range t {
		eOffset := it * 18
		for i := range triangle {
			edgeI := eOffset + i*6
			v1, v2 := triangle[i], triangle[(i+1)%3]
			e32[edgeI] = float32(v1.X)
			e32[edgeI+1] = float32(v1.Y)
			e32[edgeI+2] = float32(v1.Z)
			e32[edgeI+3] = float32(v2.X)
			e32[edgeI+4] = float32(v2.Y)
			e32[edgeI+5] = float32(v2.Z)
		}
	}
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(e32, 3))
	return geom
}

func triangleGeometry(t []Triangle, color func(i int) [3]float32) three.BufferGeometry {
	if color == nil {
		color = func(int) [3]float32 { return [3]float32{.5, .5, .5} }
	}
	// Triangle vertices, 3 per triangle.
	v32 := make([]float32, 3*3*len(t))
	// Triangle normals. One per vertex.
	n32 := make([]float32, 3*3*len(t))
	// Triangle colors. One per vertex, should be same for all triangle.
	c32 := make([]float32, 3*3*len(t))
	for it, triangle := range t {
		offset := 3 * 3 * it
		c := color(it)
		n := Unit(triangle.Normal())
		for i := range triangle {
			edgeI := offset + i*3
			v := triangle[i]
			v32[edgeI] = float32(v.X)
			v32[edgeI+1] = float32(v.Y)
			v32[edgeI+2] = float32(v.Z)
			n32[edgeI] = float32(n.X)
			n32[edgeI+1] = float32(n.Y)
			n32[edgeI+2] = float32(n.Z)
			c32[edgeI] = c[0]
			c32[edgeI+1] = c[1]
			c32[edgeI+2] = c[2]
		}
	}
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(v32, 3))
	geom.SetAttribute("normal", three.NewBufferAttribute(n32, 3))
	geom.SetAttribute("color", three.NewBufferAttribute(c32, 3))
	return geom
}

func triangleOutlines(t []Triangle, material three.MaterialParameters) three.Object3D {
	return three.NewLineSegments(triangleOutlineGeometry(t), three.NewLineBasicMaterial(&material))
}

func triangleMesh(t []Triangle, material three.Material) three.Object3D {
	return three.NewMesh(triangleGeometry(t, nil), material)
}

func pointsObj(p []Vec, material three.MaterialParameters) three.Object3D {
	p32 := make([]float32, 3*len(p))
	for it, vert := range p {
		p32[it*3] = float32(vert.X)
		p32[it*3+1] = float32(vert.Y)
		p32[it*3+2] = float32(vert.Z)
	}
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(p32, 3))
	lines := three.NewPoints(geom, three.NewPointsMaterial(material))
	return lines
}

func linesObj(edges [][2]Vec, material three.MaterialParameters) three.Object3D {
	// and need 2 points to define a line
	// and each point is defined by 3 numbers (X,Y,Z data)
	e32 := make([]float32, 2*3*len(edges))
	for it, edge := range edges {
		eOffset := it * 6
		e32[eOffset] = float32(edge[0].X)
		e32[eOffset+1] = float32(edge[0].Y)
		e32[eOffset+2] = float32(edge[0].Z)
		e32[eOffset+3] = float32(edge[1].X)
		e32[eOffset+4] = float32(edge[1].Y)
		e32[eOffset+5] = float32(edge[1].Z)
	}
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(e32, 3))
	lines := three.NewLineSegments(geom, three.NewLineBasicMaterial(&material))
	return lines
}

func boxesObj(boxes []Box, material three.MaterialParameters) three.Object3D {
	// We need 4+4+4 lines to define a box
	// and need 2 points to define a line
	// and each point is defined by 3 numbers (X,Y,Z data)
	e32 := make([]float32, 12*6*len(boxes))
	edges := [12][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}
	for it, box := range boxes {
		vertices := box.Vertices()
		boxOffset := it * 12 * 6
		for i, verts := range edges {
			eOffset := boxOffset + 6*i
			e32[eOffset] = float32(vertices[verts[0]].X)
			e32[eOffset+1] = float32(vertices[verts[0]].Y)
			e32[eOffset+2] = float32(vertices[verts[0]].Z)
			e32[eOffset+3] = float32(vertices[verts[1]].X)
			e32[eOffset+4] = float32(vertices[verts[1]].Y)
			e32[eOffset+5] = float32(vertices[verts[1]].Z)
		}
	}
	geom := three.NewBufferGeometry()
	geom.SetAttribute("position", three.NewBufferAttribute(e32, 3))
	lines := three.NewLineSegments(geom, three.NewLineBasicMaterial(&material))
	return lines
}

func sdf3Obj(obj sdf.SDF3, cells int, color string, opacity float64) three.Object3D {
	t, err := render.RenderAll(render.NewOctreeRenderer(obj, cells))
	if err != nil {
		panic(err)
	}
	tris := make([]Triangle, len(t))
	for i := range tris {
		tris[i] = Triangle{Vec(t[i][0]), Vec(t[i][1]), Vec(t[i][2])}
	}
	return triangleMesh(tris, phongMaterial(color, opacity))
}

func tetraTriangles(nodes []Vec, tetras [][4]int) []Triangle {
	var triangles []Triangle
	for _, tetra := range tetras {
		nd := [4]Vec{nodes[tetra[0]], nodes[tetra[1]], nodes[tetra[2]], nodes[tetra[3]]}
		triangles = append(triangles,
			Triangle{nd[0], nd[1], nd[2]},
			Triangle{nd[1], nd[3], nd[2]},
			Triangle{nd[0], nd[3], nd[1]},
			Triangle{nd[0], nd[2], nd[3]},
		)
	}
	return triangles
}
