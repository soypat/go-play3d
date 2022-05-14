package main

// edgeIdx represents an edge of the icosahedron
type edgeIdx [2]int

// subdivideEdge takes the vertices list and indices first and second to
// the vertices defining the edge that will be subdivided. lookup is a
// table of all newly generated vertices from previous calls to subdivideEdge
// so as to not duplicate vertices.
func subdivideEdge(lookup map[edgeIdx]int, vertices []Vec, first, second int) (int, []Vec) {
	key := edgeIdx{first, second}
	if first > second {
		// Swap to ensure edgeIdx always has lower index first.
		key[0], key[1] = key[1], key[0]
	}
	vertIdx, vertExists := lookup[key]
	if !vertExists {
		// If edge not already subdivided add
		// new dividing vertex to lookup table.
		edge0 := vertices[first]
		edge1 := vertices[second]
		point := Unit(Add(edge0, edge1)) // vertex at a normalized position.
		vertices = append(vertices, point)
		vertIdx = len(vertices) - 1
		lookup[key] = vertIdx
	}
	return vertIdx, vertices
}

func subdivide(vertices []Vec, triangles [][3]int) ([]Vec, [][3]int) {
	// We generate a lookup table of all newly generated vertices so as to not
	// duplicate new vertices. edgeIdx has lower index first.
	lookup := make(map[edgeIdx]int)
	var result [][3]int
	for _, triangle := range triangles {
		var mid [3]int
		for edge := 0; edge < 3; edge++ {
			mid[edge], vertices = subdivideEdge(lookup, vertices, triangle[edge], triangle[(edge+1)%3])
		}
		newTriangles := [][3]int{
			{triangle[0], mid[0], mid[2]},
			{triangle[1], mid[1], mid[0]},
			{triangle[2], mid[2], mid[1]},
			{mid[0], mid[1], mid[2]},
		}
		result = append(result, newTriangles...)
	}
	return vertices, result
}

// Generates an icosphere by dividing icosahedron
// faces subdivisions times
func icosphere(subdivisions int) []Triangle {
	// Attempted to be taken from here. May not be fully correct.
	// https://schneide.blog/2016/07/15/generating-an-icosphere-in-c/
	// declare icosahedron vertices
	vertices, triangles := icosahedron(1.0)
	for i := 0; i < subdivisions; i++ {
		vertices, triangles = subdivide(vertices, triangles)
	}
	var faces []Triangle
	for _, t := range triangles {
		var face Triangle
		for i := 0; i < 3; i++ {
			face[i] = vertices[t[i]]
		}
		faces = append(faces, face)
	}
	return faces
}

func icosahedron(radius float64) (vertices []Vec, triangles [][3]int) {
	const (
		x = .525731112119133606
		z = .850650808352039932
		n = 0.
	)
	X, Z, N := x*radius, z*radius, n*radius
	return []Vec{
			{-X, N, Z}, {X, N, Z}, {-X, N, -Z}, {X, N, -Z},
			{N, Z, X}, {N, Z, -X}, {N, -Z, X}, {N, -Z, -X},
			{Z, X, N}, {-Z, X, N}, {Z, -X, N}, {-Z, -X, N},
		}, [][3]int{{0, 1, 4}, {0, 4, 9}, {9, 4, 5}, {4, 8, 5},
			{4, 1, 8}, {8, 1, 10}, {8, 10, 3}, {5, 8, 3},
			{5, 3, 2}, {2, 3, 7}, {7, 3, 10}, {7, 10, 6},
			{7, 6, 11}, {11, 6, 0}, {0, 6, 1}, {6, 10, 1},
			{9, 11, 0}, {9, 2, 11}, {9, 5, 2}, {7, 11, 2}}
}
