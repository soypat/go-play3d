// Copyright ©2022 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "math"

// CenteredBox creates a Box with a given center and size.
func CenteredBox(center, size Vec) Box {
	half := Scale(0.5, size)
	return Box{Min: Sub(center, half), Max: Add(center, half)}
}

// Size returns the size of the Box.
func (a Box) Size() Vec {
	return Sub(a.Max, a.Min)
}

// Center returns the center of the Box.
func (a Box) Center() Vec {
	return Add(a.Min, Scale(0.5, a.Size()))
}

// Vertices returns a slice of the 8 vertices
// corresponding to each of the Box's corners.
//
// Vertex ordering between 0-3 and 4-7 outline
// 2D boxes in the XY plane where vertices 0-3 are in Z=Box.Min.Z.
// One can then construct the edges with indices for the
// return value of Vertices:
//  edges := [12][2]int{
//   {0, 1}, {1, 2}, {2, 3}, {3, 0},
//   {4, 5}, {5, 6}, {6, 7}, {7, 4},
//   {0, 4}, {1, 5}, {2, 6}, {3, 7},
//  }
func (a Box) Vertices() []Vec {
	return []Vec{
		a.Min,                                // 0
		{X: a.Max.X, Y: a.Min.Y, Z: a.Min.Z}, // 1
		{X: a.Max.X, Y: a.Max.Y, Z: a.Min.Z}, // 2
		{X: a.Min.X, Y: a.Max.Y, Z: a.Min.Z}, // 3
		{X: a.Min.X, Y: a.Min.Y, Z: a.Max.Z}, // 4
		{X: a.Max.X, Y: a.Min.Y, Z: a.Max.Z}, // 5
		a.Max,                                // 6
		{X: a.Min.X, Y: a.Max.Y, Z: a.Max.Z}, // 7
	}
}

// Union returns a box enclosing both the receiver and argument Boxes.
func (a Box) Union(b Box) Box {
	return Box{
		Min: minElem(a.Min, b.Min),
		Max: maxElem(a.Max, b.Max),
	}
}

// Add adds v to the bounding box components.
// It is the equivalent of translating the Box by v.
func (a Box) Add(v Vec) Box {
	return Box{Add(a.Min, v), Add(a.Max, v)}
}

// Scale returns a new Box scaled by a size vector around its center.
// The scaling is done element wise, which is to say
// the Box's X size is scaled by v.X.
func (a Box) Scale(v Vec) Box {
	// TODO(soypat): Probably a better way to do this.
	return CenteredBox(a.Center(), mulElem(v, a.Size()))
}

// Contains checks if the Box contains the given vector within its bounds.
func (a Box) Contains(v Vec) bool {
	return a.Min.X <= v.X && a.Min.Y <= v.Y && a.Min.Z <= v.Z &&
		v.X <= a.Max.X && v.Y <= a.Max.Y && v.Z <= a.Max.Z
}

func (b Box) Octree() []Box {
	ctr := b.Center()
	min := b.Min
	max := b.Max
	return []Box{
		0: {Min: b.Min, Max: ctr},
		1: {Min: Vec{ctr.X, min.Y, min.Z}, Max: Vec{max.X, ctr.Y, ctr.Z}},
		2: {Min: Vec{min.X, ctr.Y, min.Z}, Max: Vec{ctr.X, max.Y, ctr.Z}},
		3: {Min: Vec{ctr.X, ctr.Y, min.Z}, Max: Vec{max.X, max.Y, ctr.Z}},
		4: {Min: Vec{min.X, min.Y, ctr.Z}, Max: Vec{ctr.X, ctr.Y, max.Z}},
		5: {Min: Vec{ctr.X, min.Y, ctr.Z}, Max: Vec{max.X, ctr.Y, max.Z}},
		6: {Min: Vec{min.X, ctr.Y, ctr.Z}, Max: Vec{ctr.X, max.Y, max.Z}},
		7: {Min: ctr, Max: max},
	}
}

func boxDivide(b Box, maxCell int) (boxes []Box) {
	sz := b.Size()
	maxDim := math.Max(sz.Z, math.Max(sz.X, sz.Y))
	resolution := maxDim / float64(maxCell)
	div := [3]int{
		int(math.Ceil(sz.X / resolution)),
		int(math.Ceil(sz.Y / resolution)),
		int(math.Ceil(sz.Z / resolution)),
	}
	for i := 0; i < div[0]; i++ {
		x := (float64(i)+0.5)*resolution + b.Min.X
		for j := 0; j < div[1]; j++ {
			y := (float64(j)+0.5)*resolution + b.Min.Y
			for k := 0; k < div[2]; k++ {
				z := (float64(k)+0.5)*resolution + b.Min.Z
				bb := CenteredBox(Vec{x, y, z}, Vec{resolution, resolution, resolution})
				boxes = append(boxes, bb)
			}
		}
	}
	return boxes
}
