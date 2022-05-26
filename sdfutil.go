package main

import (
	"math"

	"github.com/soypat/sdf"
	"gonum.org/v1/gonum/spatial/r3"
)

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

func sdfCurvature(s sdf.SDF3, c Vec, tol float64) float64 {
	H := sdfHessian(s, c, tol)
	n := sdfNormal(s, c, tol)
	P := NewMat(nil)
	aux := NewMat(nil)
	// P is projection matrix.
	P.Outer(1, n, n)
	P.Sub(Eye(), P)
	aux.Mul(P, H)
	aux.Mul(aux, P)
	aux.Scale(Norm(n), aux)
	r, _ := aux.Eigs() // symmetric matrix!
	_, k2, k1 := sort3(r[0], r[1], r[2])
	// Eigenvalues of aux computed, discard zero eigenvalue.
	// Remaining two eigenvalues are k1 and k2, whose absolute sum
	// is the positive+negative curvature.
	return math.Abs(k1) + math.Abs(k2)
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
