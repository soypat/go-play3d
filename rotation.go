// Copyright ©2021 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/num/quat"
)

// TODO: possibly useful additions to the current rotation API:
//  - create rotations from Euler angles (NewRotationFromEuler?)
//  - create rotations from rotation matrices (NewRotationFromMatrix?)
//  - return the equivalent Euler angles from a Rotation
//
// Euler angles have issues (see [1] for a discussion).
// We should think carefully before adding them in.
// [1]: http://www.euclideanspace.com/maths/geometry/rotations/conversions/quaternionToEuler/

// Rotation describes a rotation in space.
type Rotation quat.Number

// NewRotation creates a rotation by alpha, around axis.
func NewRotation(alpha float64, axis Vec) Rotation {
	if alpha == 0 {
		return Rotation{Real: 1}
	}
	q := raise(axis)
	sin, cos := math.Sincos(0.5 * alpha)
	q = quat.Scale(sin/quat.Abs(q), q)
	q.Real += cos
	if len := quat.Abs(q); len != 1 {
		q = quat.Scale(1/len, q)
	}

	return Rotation(q)
}

// Rotate returns p rotated according to the parameters used to construct
// the receiver.
func (r Rotation) Rotate(p Vec) Vec {
	if r.isIdentity() {
		return p
	}
	qq := quat.Number(r)
	pp := quat.Mul(quat.Mul(qq, raise(p)), quat.Conj(qq))
	return Vec{X: pp.Imag, Y: pp.Jmag, Z: pp.Kmag}
}

func (r Rotation) isIdentity() bool {
	return r == Rotation{Real: 1}
}

func raise(p Vec) quat.Number {
	return quat.Number{Imag: p.X, Jmag: p.Y, Kmag: p.Z}
}

func rotateBetween(u, v Vec) Rotation {
	const tol = 1e-8
	kct := Dot(u, v)
	k := math.Sqrt(Norm2(u) * Norm2(v))
	if math.Abs(kct/k+1) < tol {
		//180 degree rotation
		return Rotation(raise(orthogonal(u)))
	}
	q := raise(Cross(u, v))
	q.Real = k + kct
	return Rotation(quat.Scale(1/quat.Abs(q), q))

}

// The orthogonal function returns any vector orthogonal to the given vector.
// This implementation uses the cross product with the most orthogonal basis vector.
func orthogonal(v Vec) Vec {
	fmt.Println("orthogonal called")
	x := math.Abs(v.X)
	y := math.Abs(v.Y)
	z := math.Abs(v.Z)
	other := Vec{Z: 1}
	if x < y {
		if x < z {
			other = Vec{X: 1}
		}
		other = Vec{Z: 1}
	}
	if y < z {
		other = Vec{Y: 1}
	}
	return Cross(v, other)
}

func vecApproxEqual(a, b Vec, tol float64) bool {
	return math.Abs(a.X-b.X) < tol &&
		math.Abs(a.Y-b.Y) < tol &&
		math.Abs(a.Z-b.Z) < tol
}

func rotateOntoVecHalfway(u, v Vec) Rotation {
	// Vector half way solution.
	u = Unit(u)
	v = Unit(v)
	if vecApproxEqual(u, v, 1e-8) {
		return NewRotation(0, Vec{})
	}
	if vecApproxEqual(Scale(-1, u), v, 1e-8) {
		return Rotation(raise(orthogonal(u)))
	}
	half := Unit(Add(u, v))
	raised := raise(Cross(u, half))
	raised.Real = Dot(u, half)
	return Rotation(raised)
}
