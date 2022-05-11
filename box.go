// Copyright ©2022 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

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
// Vertex order is unspecified.
func (a Box) Vertices() []Vec {
	return []Vec{
		a.Min,
		{X: a.Min.X, Y: a.Min.Y, Z: a.Max.Z},
		{X: a.Min.X, Y: a.Max.Y, Z: a.Min.Z},
		{X: a.Min.X, Y: a.Max.Y, Z: a.Max.Z},
		{X: a.Max.X, Y: a.Min.Y, Z: a.Min.Z},
		{X: a.Max.X, Y: a.Min.Y, Z: a.Max.Z},
		{X: a.Max.X, Y: a.Max.Y, Z: a.Min.Z},
		a.Max,
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
