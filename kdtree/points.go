// Copyright ©2019 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kdtree

import "math"

var (
	_ Interface[Vec, Vec] = VecSet(nil)
	_ Comparable[Vec]     = Vec(nil)
	_ Point               = Vec(nil)
)

// Vec represents a point in a k-d space that satisfies the Comparable interface.
type Vec []float64

type VecBounding Bounding[Vec]

func (p Vec) Component(i Dim) float64 { return p[i] }

// Dims returns the number of dimensions described by the receiver.
func (p Vec) Dims() int { return len(p) }
func (p Vec) Len() int  { return len(p) }

// Compare returns the signed distance of p from the plane passing through c and
// perpendicular to the dimension d. The concrete type of c must be Point.
func (p Vec) ComparePoint(q Vec, d Dim) float64 { return p[d] - q[d] }
func (p Vec) Point() Vec                        { return p }

// Distance returns the squared Euclidean distance between c and the receiver. The
// concrete type of c must be Point.
func (p Vec) Distance(q Vec) float64 {
	var sum float64
	for dim, c := range p {
		d := c - q[dim]
		sum += d * d
	}
	return sum
}

// Extend returns a bounding box that has been extended to include the receiver.
func (p Vec) Extend(b *VecBounding) *VecBounding {
	if b == nil {
		b = &VecBounding{append(Vec(nil), p...), append(Vec(nil), p...)}
	}
	min := b.Min
	max := b.Max
	for d, v := range p {
		min[d] = math.Min(min[d], v)
		max[d] = math.Max(max[d], v)
	}
	*b = VecBounding{Min: min, Max: max} // TODO is this correct?
	return b
}

// VecSet is a collection of point values that satisfies the Interface.
type VecSet []Vec

func (p VecSet) Bounds() (min, max Vec) {
	if p.Len() == 0 {
		return nil, nil
	}
	min = append(Vec(nil), p[0]...)
	max = append(Vec(nil), p[0]...)
	for _, e := range p[1:] {
		for d, v := range e {
			min[d] = math.Min(min[d], v)
			max[d] = math.Max(max[d], v)
		}
	}
	return min, max
}
func (p VecSet) Index(i int) Vec                          { return p[i] }
func (p VecSet) Len() int                                 { return len(p) }
func (p VecSet) Pivot(d Dim) int                          { return Plane{VecSet: p, Dim: d}.Pivot() }
func (p VecSet) Slice(start, end int) Interface[Vec, Vec] { return p[start:end] }

// Plane is a wrapping type that allows a Points type be pivoted on a dimension.
// The Pivot method of Plane uses MedianOfRandoms sampling at most 100 elements
// to find a pivot element.
type Plane struct {
	Dim
	VecSet
}

// randoms is the maximum number of random values to sample for calculation of
// median of random elements.
const randoms = 100

func (p Plane) Less(i, j int) bool              { return p.VecSet[i][p.Dim] < p.VecSet[j][p.Dim] }
func (p Plane) Pivot() int                      { return Partition(p, MedianOfRandoms(p, randoms)) }
func (p Plane) Slice(start, end int) SortSlicer { p.VecSet = p.VecSet[start:end]; return p }
func (p Plane) Swap(i, j int)                   { p.VecSet[i], p.VecSet[j] = p.VecSet[j], p.VecSet[i] }
