// Copyright ©2019 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kdtree

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
)

type Point interface {
	Dims() int
	Component(i Dim) float64
}

// Interface is the set of methods required for construction of efficiently
// searchable k-d trees. A k-d tree may be constructed without using the
// Interface type, but it is likely to have reduced search performance.
type Interface[P Point, T Comparable[P]] interface {
	// Index returns the ith element of the list of points.
	Index(i int) T

	// Len returns the length of the list.
	Len() int

	// Pivot partitions the list based on the dimension specified.
	Pivot(Dim) int

	// Slice returns a slice of the list using zero-based half
	// open indexing equivalent to built-in slice indexing.
	Slice(start, end int) Interface[P, T]
}

// Bounder returns a bounding volume containing the list of points. Bounds may return nil.
type Bounder[P Point] interface {
	Bounds() (min, max P)
}

type bounder[P Point, T Comparable[P]] interface {
	Interface[P, T]
	Bounder[P]
}

// Dim is an index into a point's coordinates.
type Dim int

// Comparable is the element interface for values stored in a k-d tree.
type Comparable[P Point] interface {
	// Compare returns the signed distance of a from the plane passing through
	// b and perpendicular to the dimension d.
	//
	// Given c = a.Compare(b, d):
	//  c = a_d - b_d
	//
	// Compare(Comparable[P], Dim) float64 // Probably remove this in favor of comparePoint
	ComparePoint(P, Dim) float64
	Point() P

	// Distance returns the squared Euclidean distance between the receiver and
	// the parameter.
	Distance(P) float64
}

// Extender is a Comparable that can increase a bounding volume to include the
// point represented by the Comparable.
type Extender[P Point] interface {
	Comparable[P]

	// Extend returns a bounding box that has been extended to include the
	// receiver. Extend may return nil.
	Extend(*Bounding[P]) *Bounding[P]
}

// Bounding represents a volume bounding box.
type Bounding[P Point] struct {
	Min, Max P
}

// Contains returns whether c is within the volume of the Bounding. A nil Bounding
// returns true.
func (b *Bounding[P]) Contains(c P) bool {
	if b == nil {
		return true
	}
	for d := Dim(0); d < Dim(c.Dims()); d++ {
		comp := c.Component(d)
		if comp < b.Min.Component(d) || comp > b.Max.Component(d) {
			return false
		}
	}
	return true
}

// Node holds a single point value in a k-d tree.
type Node[P Point, T Comparable[P]] struct {
	Point       T
	Plane       Dim
	Left, Right *Node[P, T]
	*Bounding[P]
}

func (n *Node[P, T]) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v %d", n.Point.Point(), n.Plane)
}

// Tree implements a k-d tree creation and nearest neighbor search.
type Tree[P Point, T Comparable[P]] struct {
	Root  *Node[P, T]
	Count int
}

// New returns a k-d tree constructed from the values in p. If p is a Bounder and
// bounding is true, bounds are determined for each node.
// The ordering of elements in p may be altered after New returns.
func New[P Point, T Comparable[P]](p Interface[P, T], bounding bool) *Tree[P, T] {
	if p, ok := p.(bounder[P, T]); ok && bounding {
		return &Tree[P, T]{
			Root:  buildBounded(p, 0, bounding),
			Count: p.Len(),
		}
	}
	return &Tree[P, T]{
		Root:  build(p, 0),
		Count: p.Len(),
	}
}

func build[P Point, T Comparable[P]](p Interface[P, T], plane Dim) *Node[P, T] {
	if p.Len() == 0 {
		return nil
	}

	piv := p.Pivot(plane)
	d := p.Index(piv)
	np := (plane + 1) % Dim(d.Point().Dims())

	return &Node[P, T]{
		Point:    d,
		Plane:    plane,
		Left:     build(p.Slice(0, piv), np),
		Right:    build(p.Slice(piv+1, p.Len()), np),
		Bounding: nil,
	}
}

func buildBounded[P Point, T Comparable[P]](p bounder[P, T], plane Dim, bounding bool) *Node[P, T] {
	if p.Len() == 0 {
		return nil
	}

	piv := p.Pivot(plane)
	d := p.Index(piv)
	np := (plane + 1) % Dim(d.Point().Dims())

	min, max := p.Bounds()
	return &Node[P, T]{
		Point:    d,
		Plane:    plane,
		Left:     buildBounded(p.Slice(0, piv).(bounder[P, T]), np, bounding),
		Right:    buildBounded(p.Slice(piv+1, p.Len()).(bounder[P, T]), np, bounding),
		Bounding: &Bounding[P]{Min: min, Max: max},
	}
}

// Insert adds a point to the tree, updating the bounding volumes if bounding is
// true, and the tree is empty or the tree already has bounding volumes stored,
// and c is an Extender. No rebalancing of the tree is performed.
func (t *Tree[P, T]) Insert(c T, bounding bool) {
	t.Count++
	if t.Root != nil {
		bounding = t.Root.Bounding != nil
	}
	// Generics error:
	// if c, ok := c.(Extender[P]); ok && bounding {
	// 	t.Root = t.Root.insertBounded(c, 0, bounding)
	// 	return
	// } else if !ok && t.Root != nil {
	// 	// If we are not rebounding, mark the tree as non-bounded.
	// 	t.Root.Bounding = nil
	// }
	t.Root = t.Root.insert(c, 0)
}

func (n *Node[P, T]) insert(c T, d Dim) *Node[P, T] {
	if n == nil {
		return &Node[P, T]{
			Point:    c,
			Plane:    d,
			Bounding: nil,
		}
	}

	d = (n.Plane + 1) % Dim(c.Point().Dims())
	if c.ComparePoint(n.Point.Point(), n.Plane) <= 0 {
		n.Left = n.Left.insert(c, d)
	} else {
		n.Right = n.Right.insert(c, d)
	}

	return n
}

// func (n *Node[P, T]) insertBounded(c Extender[P], d Dim, bounding bool) *Node[P, T] {
// 	if n == nil {
// 		var b *Bounding[P]
// 		if bounding {
// 			b = c.Extend(b)
// 		}
// 		return &Node[P, T]{
// 			Point:    c,
// 			Plane:    d,
// 			Bounding: b,
// 		}
// 	}

// 	if bounding {
// 		n.Bounding = c.Extend(n.Bounding)
// 	}
// 	d = (n.Plane + 1) % Dim(c.Dims())
// 	if c.Compare(n.Point, n.Plane) <= 0 {
// 		n.Left = n.Left.insertBounded(c, d, bounding)
// 	} else {
// 		n.Right = n.Right.insertBounded(c, d, bounding)
// 	}

// 	return n
// }

// Len returns the number of elements in the tree.
func (t *Tree[P, T]) Len() int { return t.Count }

// Contains returns whether a Comparable is in the bounds of the tree. If no bounding has
// been constructed Contains returns true.
func (t *Tree[P, T]) Contains(c P) bool {
	if t.Root.Bounding == nil {
		return true
	}
	return t.Root.Contains(c)
}

var inf = math.Inf(1)

// Nearest returns the nearest value to the query and the distance between them.
func (t *Tree[P, T]) Nearest(q P) (nearest T, dist2 float64) {
	if t.Root == nil {
		return nearest, inf
	}
	n, dist := t.Root.search(q, inf)
	if n == nil {
		return nearest, inf
	}
	return n.Point, dist
}

func (n *Node[P, T]) search(q P, dist float64) (*Node[P, T], float64) {
	if n == nil {
		return nil, inf
	}
	c := -n.Point.ComparePoint(q, n.Plane)
	// c := q.ComparePoint(n.Point.Point(), n.Plane)
	dist = math.Min(dist, n.Point.Distance(q))
	// dist = math.Min(dist, q.Distance(n.Point.Point()))
	// dist = math.Min(dist, q.Distance(n.Point.Point()))

	bn := n
	if c <= 0 {
		ln, ld := n.Left.search(q, dist)
		if ld < dist {
			bn, dist = ln, ld
		}
		if c*c < dist {
			rn, rd := n.Right.search(q, dist)
			if rd < dist {
				bn, dist = rn, rd
			}
		}
		return bn, dist
	}
	rn, rd := n.Right.search(q, dist)
	if rd < dist {
		bn, dist = rn, rd
	}
	if c*c < dist {
		ln, ld := n.Left.search(q, dist)
		if ld < dist {
			bn, dist = ln, ld
		}
	}
	return bn, dist
}

// ComparableDist holds a Comparable and a distance to a specific query. A nil Comparable
// is used to mark the end of the heap, so clients should not store nil values except for
// this purpose.
type ComparableDist[P Point, T Comparable[P]] struct {
	Comparable T
	Dist       float64
}

// Heap is a max heap sorted on Dist.
type Heap[P Point, T Comparable[P]] []ComparableDist[P, T]

func (h *Heap[P, T]) Max() ComparableDist[P, T] { return (*h)[0] }
func (h *Heap[P, T]) Len() int                  { return len(*h) }
func (h *Heap[P, T]) Less(i, j int) bool {
	return (*h)[i].Dist > (*h)[j].Dist
	// TODO return (*h)[i] == nil || (*h)[i].Dist > (*h)[j].Dist
}
func (h *Heap[P, T]) Swap(i, j int)        { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }
func (h *Heap[P, T]) Push(x interface{})   { (*h) = append(*h, x.(ComparableDist[P, T])) }
func (h *Heap[P, T]) Pop() (i interface{}) { i, *h = (*h)[len(*h)-1], (*h)[:len(*h)-1]; return i }

// NKeeper is a Keeper that retains the n best ComparableDists that have been passed to Keep.
type NKeeper[P Point, T Comparable[P]] struct {
	Heap[P, T]
}

// NewNKeeper returns an NKeeper with the max value of the heap set to infinite distance. The
// returned NKeeper is able to retain at most n values.
func NewNKeeper[P Point, T Comparable[P]](n int) *NKeeper[P, T] {
	k := NKeeper[P, T]{make(Heap[P, T], 1, n)}
	k.Heap[0].Dist = inf
	return &k
}

// Keep adds c to the heap if its distance is less than the maximum value of the heap. If adding
// c would increase the size of the heap beyond the initial maximum length, the maximum value of
// the heap is dropped.
func (k *NKeeper[P, T]) Keep(c ComparableDist[P, T]) {
	if c.Dist <= k.Heap[0].Dist { // Favour later finds to displace sentinel.
		if len(k.Heap) == cap(k.Heap) {
			heap.Pop(k)
		}
		heap.Push(k, c)
	}
}

// DistKeeper is a Keeper that retains the ComparableDists within the specified distance of the
// query that it is called to Keep.
type DistKeeper[P Point, T Comparable[P]] struct {
	Heap[P, T]
}

// NewDistKeeper returns an DistKeeper with the maximum value of the heap set to d.
func NewDistKeeper[P Point, T Comparable[P]](d float64) *DistKeeper[P, T] {
	return &DistKeeper[P, T]{Heap[P, T]{{Dist: d}}}
}

// Keep adds c to the heap if its distance is less than or equal to the max value of the heap.
func (k *DistKeeper[P, T]) Keep(c ComparableDist[P, T]) {
	if c.Dist <= k.Heap[0].Dist {
		heap.Push(k, c)
	}
}

// Keeper implements a conditional max heap sorted on the Dist field of the ComparableDist type.
// kd search is guided by the distance stored in the max value of the heap.
type Keeper[P Point, T Comparable[P]] interface {
	Keep(ComparableDist[P, T]) // Keep conditionally pushes the provided ComparableDist onto the heap.
	Max() ComparableDist[P, T] // Max returns the maximum element of the Keeper.
	heap.Interface
}

// NearestSet finds the nearest values to the query accepted by the provided Keeper, k.
// k must be able to return a ComparableDist specifying the maximum acceptable distance
// when Max() is called, and retains the results of the search in min sorted order after
// the call to NearestSet returns.
// If a sentinel ComparableDist with a nil Comparable is used by the Keeper to mark the
// maximum distance, NearestSet will remove it before returning.
func (t *Tree[P, T]) NearestSet(k Keeper[P, T], q T) {
	if t.Root == nil {
		return
	}
	t.Root.searchSet(q, k)

	// Check whether we have retained a sentinel
	// and flag removal if we have.
	removeSentinel := k.Len() != 0 // TODO && k.Max().Comparable == nil

	sort.Sort(sort.Reverse(k))

	// This abuses the interface to drop the max.
	// It is reasonable to do this because we know
	// that the maximum value will now be at element
	// zero, which is removed by the Pop method.
	if removeSentinel {
		k.Pop()
	}
}

func (n *Node[P, T]) searchSet(q T, k Keeper[P, T]) {
	if n == nil {
		return
	}

	c := q.ComparePoint(n.Point.Point(), n.Plane)
	k.Keep(ComparableDist[P, T]{Comparable: n.Point, Dist: q.Distance(n.Point.Point())}) // TODO
	if c <= 0 {
		n.Left.searchSet(q, k)
		if c*c <= k.Max().Dist {
			n.Right.searchSet(q, k)
		}
		return
	}
	n.Right.searchSet(q, k)
	if c*c <= k.Max().Dist {
		n.Left.searchSet(q, k)
	}
}

// Operation is a function that operates on a Comparable. The bounding volume and tree depth
// of the point is also provided. If done is returned true, the Operation is indicating that no
// further work needs to be done and so the Do function should traverse no further.
type Operation[P Point, T Comparable[P]] func(T, *Bounding[P], int) (done bool)

// Do performs fn on all values stored in the tree. A boolean is returned indicating whether the
// Do traversal was interrupted by an Operation returning true. If fn alters stored values' sort
// relationships, future tree operation behaviors are undefined.
func (t *Tree[P, T]) Do(fn Operation[P, T]) bool {
	if t.Root == nil {
		return false
	}
	return t.Root.do(fn, 0)
}

func (n *Node[P, T]) do(fn Operation[P, T], depth int) (done bool) {
	if n.Left != nil {
		done = n.Left.do(fn, depth+1)
		if done {
			return
		}
	}
	done = fn(n.Point, n.Bounding, depth)
	if done {
		return
	}
	if n.Right != nil {
		done = n.Right.do(fn, depth+1)
	}
	return
}

// DoBounded performs fn on all values stored in the tree that are within the specified bound.
// If b is nil, the result is the same as a Do. A boolean is returned indicating whether the
// DoBounded traversal was interrupted by an Operation returning true. If fn alters stored
// values' sort relationships future tree operation behaviors are undefined.
func (t *Tree[P, T]) DoBounded(b *Bounding[P], fn Operation[P, T]) bool {
	if t.Root == nil {
		return false
	}
	if b == nil {
		return t.Root.do(fn, 0)
	}
	return t.Root.doBounded(fn, b, 0)
}

func (n *Node[P, T]) doBounded(fn Operation[P, T], b *Bounding[P], depth int) (done bool) {
	// TODO which is better? See operation below too.
	// n.Point.Point().Component(n.Plane) < min
	// n.Point.ComparePoint(b.Min, n.Plane) > 0
	if n.Left != nil && n.Point.ComparePoint(b.Min, n.Plane) > 0 { // TODO b.Min.Compare(n.Point, n.Plane) < 0
		done = n.Left.doBounded(fn, b, depth+1)
		if done {
			return
		}
	}
	// TODO: this looks OK for now:
	if b.Contains(n.Point.Point()) {
		done = fn(n.Point, n.Bounding, depth)
		if done {
			return
		}
	}
	// TODO
	if n.Right != nil && n.Point.ComparePoint(b.Max, n.Plane) < 0 { // before: 0 < b.Max.Compare(n.Point, n.Plane) {
		done = n.Right.doBounded(fn, b, depth+1)
	}
	return
}
