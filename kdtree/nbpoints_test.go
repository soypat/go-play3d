// Copyright ©2019 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kdtree

var (
	_ Interface[nbPoint, nbPoint] = nbPoints{}
	_ Comparable[nbPoint]         = nbPoint{}
)

// nbRandoms is the maximum number of random values to sample for calculation of median of
// random elements.
var nbRandoms = 100

// nbPoint represents a point in a k-d space that satisfies the Comparable interface.
type nbPoint Vec

func (p nbPoint) Len() int                { return len(p) }
func (p nbPoint) Component(i Dim) float64 { return p[i] }
func (p nbPoint) Point() nbPoint          { return p }

func (p nbPoint) ComparePoint(c nbPoint, d Dim) float64 { return p[d] - c[d] }
func (p nbPoint) Dims() int                             { return len(p) }
func (p nbPoint) Distance(q nbPoint) float64 {
	var sum float64
	for dim, c := range p {
		d := c - q[dim]
		sum += d * d
	}
	return sum
}

// nbPoints is a collection of point values that satisfies the Interface.
type nbPoints []nbPoint

func (p nbPoints) Index(i int) nbPoint                              { return p[i] }
func (p nbPoints) Len() int                                         { return len(p) }
func (p nbPoints) Pivot(d Dim) int                                  { return nbPlane{nbPoints: p, Dim: d}.Pivot() }
func (p nbPoints) Slice(start, end int) Interface[nbPoint, nbPoint] { return p[start:end] }

// nbPlane is a wrapping type that allows a Points type be pivoted on a dimension.
type nbPlane struct {
	Dim
	nbPoints
}

func (p nbPlane) Less(i, j int) bool              { return p.nbPoints[i][p.Dim] < p.nbPoints[j][p.Dim] }
func (p nbPlane) Pivot() int                      { return Partition(p, MedianOfRandoms(p, nbRandoms)) }
func (p nbPlane) Slice(start, end int) SortSlicer { p.nbPoints = p.nbPoints[start:end]; return p }
func (p nbPlane) Swap(i, j int) {
	p.nbPoints[i], p.nbPoints[j] = p.nbPoints[j], p.nbPoints[i]
}
