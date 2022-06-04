package main

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3/must3"
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

type sdftransform struct {
	sdf sdf.SDF3
	inv Transformer
}

func (t sdftransform) Evaluate(v r3.Vec) float64 {
	vv := Vec(v)
	return t.sdf.Evaluate(r3.Vec(t.inv.Transform(vv)))
}

func (t sdftransform) Bounds() r3.Box {
	var bbt Box
	bbr3 := t.sdf.Bounds()
	bb := Box{Min: Vec(bbr3.Min), Max: Vec(bbr3.Max)}
	bb = bb.Scale(Vec{1.05, 1.05, 1.05})
	if a, ok := t.inv.(Affine); ok {
		bbt = a.ApplyBox(bb)
	} else {
		bbt = bb.TransformBox(t.inv)
	}
	return r3.Box{
		Min: r3.Vec(bbt.Min),
		Max: r3.Vec(bbt.Max),
	}
}

func csgBasic() (s sdf.SDF3) {
	const max = 0.03
	s = must3.Box(r3.Vec(Elem(1)), 0)
	c := must3.Cylinder(2, .3, 0)
	cx := sdf.Transform3D(c, sdf.RotateY(math.Pi/2))
	cz := sdf.Transform3D(c, sdf.RotateX(math.Pi/2))
	var d sdf.SDF3Diff
	d = sdf.Difference3D(s, cx)
	d.SetMax(sdf.MaxPoly(2, max))
	d = sdf.Difference3D(d, cz)
	d.SetMax(sdf.MaxPoly(2, max))
	d = sdf.Difference3D(d, c)
	d.SetMax(sdf.MaxPoly(2, max))
	s = sdf.Intersect3D(d, must3.Sphere(.7))
	return s
}
