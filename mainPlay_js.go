//go:build js

package main

import (
	"fmt"
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3/must3"
	"github.com/soypat/three"
	"gonum.org/v1/gonum/spatial/r3"
)

func addObjects(grp three.Group) {
	grp.Add(three.NewAxesHelper(1))
	const quality = 20
	s := must3.Cylinder(2, .5, .1)

	grp.Add(sdf3Obj(s, quality, "red", .5))
	sbb := s.Bounds()
	bb := Box{Vec(sbb.Min), Vec(sbb.Max)}

	// grp.Add(boxesObj(bb.Octree(), lineColor("green")))
	boxes := boxDivide(CenteredBox(bb.Center(), bb.Scale(Vec{1.1, 1.1, 1.1}).Size()), 10)
	// Projection matrix.
	eye := Eye()
	P := NewMat(nil)
	aux := NewMat(nil)
	norms := make([][2]Vec, len(boxes))
	for i, box := range boxes {
		const tol = 1e-3
		c := box.Center()
		H := sdfHessian(s, c, tol)
		n := sdfNormal(s, c, tol)

		P.Outer(1, n, n)
		P.Sub(eye, P)
		aux.Mul(P, H)
		aux.Mul(aux, P)
		aux.Scale(Norm(n), aux)
		r, _ := aux.Eigs() // symmetric matrix!
		_, k2, k1 := sort3(r[0], r[1], r[2])
		curvature := math.Abs(k1) + math.Abs(k2)
		norms[i] = [2]Vec{c, Add(c, Scale(curvature*1e5, n))}
		fmt.Println(curvature)
		// Eigenvalues of aux computed, discard zero eigenvalue.
		// Remaining two eigenvalues are k1 and k2, whose absolute sum
		// is the positive+negative curvature.
	}
	grp.Add(boxesObj(boxes, lineColor("blue")))
	grp.Add(linesObj(norms, lineColor("gold")))
}

func sdfHessian(sdf sdf.SDF3, p Vec, h float64) *Mat {
	h2 := h * h * 4
	dx := Vec{X: h}
	dy := Vec{Y: h}
	dz := Vec{Z: h}
	eval := func(p Vec) float64 { return sdf.Evaluate(r3.Vec(p)) }
	fp := eval(p)
	diff2 := func(p, d1, d2 Vec, f func(p Vec) float64) float64 {
		return (f(Add(p, Add(d1, d2))) - f(Add(p, d2)) - f(Add(p, d1)) + fp) / h2
	}
	fxx := diff2(p, dx, dx, eval)
	fyy := diff2(p, dy, dy, eval)
	fzz := diff2(p, dz, dz, eval)
	fxy := diff2(p, dx, dy, eval)
	fxz := diff2(p, dx, dz, eval)
	fyz := diff2(p, dy, dz, eval)
	return NewMat([]float64{
		fxx, fxy, fxz,
		fxy, fyy, fyz,
		fxz, fyz, fzz,
	})
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

// sdfNormal returns the normal of an SDF3 at a point (doesn't need to be on the surface).
// Computed by sampling it 6 times inside a box of side 2*h centered on p.
func sdfNormal(s sdf.SDF3, p Vec, h float64) Vec {
	Evaluate := func(p Vec) float64 { return s.Evaluate(r3.Vec(p)) }
	return Vec{
		X: Evaluate(Add(p, Vec{X: h})) - Evaluate(Add(p, Vec{X: -h})),
		Y: Evaluate(Add(p, Vec{Y: h})) - Evaluate(Add(p, Vec{Y: -h})),
		Z: Evaluate(Add(p, Vec{Z: h})) - Evaluate(Add(p, Vec{Z: -h})),
	}
}
