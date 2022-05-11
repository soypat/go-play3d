package main

import (
	"math"
)

type sdfTriangle struct {
	Vertices [3]int // Indices to vertices
	N        Vec    // Face normal
	Ne       [3]Vec // Edge Normals
}

type sdfVertex struct {
	V Vec
	// N is the weighted pseudo normal where the weights
	// are the opening angle formed by edges for the triangle.
	N Vec // Vertex Normal
}

type mesh struct {
	vertices  []sdfVertex
	triangles []sdfTriangle
	// access toedge pseudo normals using vertex index.
	edgeNorm map[[2]int]Vec
}

func newMesh(triangles []Triangle, tol float64) mesh {
	bb := Box{Elem(math.MaxFloat64), Elem(-math.MaxFloat64)}
	for i := range triangles {
		for _, edge := range triangles[i] {
			bb.Min = minElem(bb.Min, edge)
			bb.Max = maxElem(bb.Max, edge)
		}
	}
	m := mesh{
		triangles: make([]sdfTriangle, len(triangles)),
		edgeNorm:  make(map[[2]int]Vec),
	}
	// center := bb.Center()
	size := bb.Size()
	maxDim := math.Max(size.X, math.Max(size.Y, size.Z))
	div := int(maxDim/tol + 1e-12)
	if div <= 0 || div > math.MaxInt32 {
		panic("bad cell divisions")
	}
	//vertex index cache
	cache := make(map[Veci]int)
	ri := 1 / tol
	for i, tri := range triangles {
		norm := tri.Normal()
		sdfT := sdfTriangle{N: Scale(2*math.Pi, norm)}
		for j, vert := range triangles[i] {
			// Scale vert to be integer in resolution-space.
			vi := R3ToI(Scale(ri, vert))
			vertexIdx, ok := cache[vi]
			if !ok {
				// Initialize the vertex if not in cache.
				vertexIdx = len(m.vertices)
				cache[vi] = vertexIdx
				m.vertices = append(m.vertices, sdfVertex{
					V: vert,
				})
			}
			// Calculate vertex pseudo normal
			s1, s2 := Sub(vert, tri[(j+1)%3]), Sub(vert, tri[(j+2)%3])
			alpha := math.Acos(Cos(s1, s2))
			m.vertices[vertexIdx].N = Add(m.vertices[vertexIdx].N, Scale(alpha, norm))
			sdfT.Vertices[j] = vertexIdx
		}
		m.triangles[i] = sdfT
		// Calculate edge pseudo normals.
		for j := range sdfT.Vertices {
			edge := [2]int{sdfT.Vertices[j], sdfT.Vertices[(j+1)%3]}
			m.edgeNorm[edge] = Add(m.edgeNorm[edge], Scale(math.Pi, norm))
		}
	}
	return m
}

// cache3 implements a 3 dimensional distance cache to avoid repeated evaluations.
// Experimentally about 2/3 of lookups get a hit, and the overall speedup
// is about 2x a non-cached evaluation.
type cache3[T any] struct {
	cache         map[Veci]T  // cache of distances
	evaluator     func(Vec) T // Spatialized data
	origin        Vec         // origin of the overall bounding cube
	invResolution float64
	resolution    float64 // size of smallest octree cube
}

// Evaluate checks cache
func (dc *cache3[T]) Evaluate(v Vec) (Veci, T) {
	vi := R3ToI(Scale(dc.invResolution, v))
	// do we have it in the cache?
	dist, found := dc.cache[vi]
	if found {
		// succesful cache hit.
		return vi, dist
	}
	// evaluate the function
	dist = dc.evaluator(v)
	// write it to the cache
	dc.cache[vi] = dist
	return vi, dist
}

func newCache3[T any](origin Vec, resolution float64, evaluator func(Vec) T) *cache3[T] {
	// if n >= 64 {
	// 	panic("size of n must be less than size of word for hdiag generation")
	// }
	// TODO heuristic for initial cache size. Maybe k * (1 << n)^3
	// Avoiding any resizing of the map seems to be worth 2-5% of speedup.
	dc := cache3[T]{
		origin:     origin,
		resolution: resolution,
		evaluator:  evaluator,
		cache:      make(map[Veci]T),
	}
	return &dc
}
