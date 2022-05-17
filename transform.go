// Copyright ©2022 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"
)

// Transform represents a 3D spatial transformation.
// The zero value of Transform is the identity transform.
type Transform struct {
	// in order to make the zero value of Transform represent the identity
	// transform we store it with the identity matrix subtracted.
	// These diagonal elements are subtracted such that
	//  d00 = x00-1, d11 = x11-1, d22 = x22-1, d33 = x33-1
	// where x00, x11, x22, x33 are the matrix diagonal elements.
	// We can then check for identity in if blocks like so:
	//  if T == (Transform{})
	d00, x01, x02, x03 float64
	x10, d11, x12, x13 float64
	x20, x21, d22, x23 float64
	x30, x31, x32, d33 float64
}

// Transform applies the Transform to the argument vector
// and returns the result.
func (t Transform) Transform(v Vec) Vec {
	// https://github.com/mrdoob/three.js/blob/dev/src/math/Vector3.js#L262
	w := 1 / (t.x30*v.X + t.x31*v.Y + t.x32*v.Z + t.d33 + 1)
	return Vec{
		X: ((t.d00+1)*v.X + t.x01*v.Y + t.x02*v.Z + t.x03) * w,
		Y: (t.x10*v.X + (t.d11+1)*v.Y + t.x12*v.Z + t.x13) * w,
		Z: (t.x20*v.X + t.x21*v.Y + (t.d22+1)*v.Z + t.x23) * w,
	}
}

// zeroTransform is the Transform that returns zeroTransform when multiplied by any Transform.
var zeroTransform = Transform{d00: -1, d11: -1, d22: -1, d33: -1}

// NewTransform returns a new Transform type and populates its elements
// with values passed in row-major form. If val is nil then NewTransform
// returns a Transform filled with zeros.
func NewTransform(a []float64) Transform {
	if a == nil {
		return zeroTransform
	}
	if len(a) != 16 {
		panic("Transform is initialized with 16 values")
	}
	return Transform{
		d00: a[0], x01: a[1], x02: a[2], x03: a[3],
		x10: a[4], d11: a[5], x12: a[6], x13: a[7],
		x20: a[8], x21: a[9], d22: a[10], x23: a[11],
		x30: a[12], x31: a[13], x32: a[14], d33: a[15],
	}
}

// ComposeTransform creates a new transform for a given translation to
// positon, scaling vector scale and quaternion rotation.
// The identity Transform is constructed with
//  ComposeTransform(Vec{}, Vec{1,1,1}, Rotation{})
func ComposeTransform(position, scale Vec, q Rotation) Transform {
	x2 := q.Imag + q.Imag
	y2 := q.Jmag + q.Jmag
	z2 := q.Kmag + q.Kmag
	xx := q.Imag * x2
	yy := q.Jmag * y2
	zz := q.Kmag * z2
	xy := q.Imag * y2
	xz := q.Imag * z2
	yz := q.Jmag * z2
	wx := q.Real * x2
	wy := q.Real * y2
	wz := q.Real * z2

	var t Transform
	t.d00 = (1-(yy+zz))*scale.X - 1
	t.x10 = (xy + wz) * scale.X
	t.x20 = (xz - wy) * scale.X

	t.x01 = (xy - wz) * scale.Y
	t.d11 = (1-(xx+zz))*scale.Y - 1
	t.x21 = (yz + wx) * scale.Y

	t.x02 = (xz + wy) * scale.Z
	t.x12 = (yz - wx) * scale.Z
	t.d22 = (1-(xx+yy))*scale.Z - 1

	t.x03 = position.X
	t.x13 = position.Y
	t.x23 = position.Z
	return t
}

// Translate adds Vec to the positional Transform.
func (t Transform) Translate(v Vec) Transform {
	t.x03 += v.X
	t.x13 += v.Y
	t.x23 += v.Z
	return t
}

// Scale returns the transform with scaling added around
// the argumnt origin.
func (t Transform) Scale(origin, factor Vec) Transform {
	if origin == (Vec{}) {
		return t.scale(factor)
	}
	t = t.Translate(Scale(-1, origin))
	t = t.scale(factor)
	return t.Translate(origin)
}

func (t Transform) scale(factor Vec) Transform {
	t.d00 = (t.d00+1)*factor.X - 1
	t.x10 *= factor.X
	t.x20 *= factor.X
	t.x30 *= factor.X

	t.x01 *= factor.Y
	t.d11 = (t.d11+1)*factor.Y - 1
	t.x21 *= factor.Y
	t.x31 *= factor.Y

	t.x02 *= factor.Z
	t.x12 *= factor.Z
	t.d22 = (t.d22+1)*factor.Z - 1
	t.x32 *= factor.Z
	return t
}

// Mul multiplies the Transforms a and b and returns the result.
// This is the equivalent of combining two transforms in one.
func (t Transform) Mul(b Transform) Transform {
	if t == (Transform{}) {
		return b
	}
	if b == (Transform{}) {
		return t
	}
	x00 := t.d00 + 1
	x11 := t.d11 + 1
	x22 := t.d22 + 1
	x33 := t.d33 + 1
	y00 := b.d00 + 1
	y11 := b.d11 + 1
	y22 := b.d22 + 1
	y33 := b.d33 + 1
	var m Transform
	m.d00 = x00*y00 + t.x01*b.x10 + t.x02*b.x20 + t.x03*b.x30 - 1
	m.x10 = t.x10*y00 + x11*b.x10 + t.x12*b.x20 + t.x13*b.x30
	m.x20 = t.x20*y00 + t.x21*b.x10 + x22*b.x20 + t.x23*b.x30
	m.x30 = t.x30*y00 + t.x31*b.x10 + t.x32*b.x20 + x33*b.x30
	m.x01 = x00*b.x01 + t.x01*y11 + t.x02*b.x21 + t.x03*b.x31
	m.d11 = t.x10*b.x01 + x11*y11 + t.x12*b.x21 + t.x13*b.x31 - 1
	m.x21 = t.x20*b.x01 + t.x21*y11 + x22*b.x21 + t.x23*b.x31
	m.x31 = t.x30*b.x01 + t.x31*y11 + t.x32*b.x21 + x33*b.x31
	m.x02 = x00*b.x02 + t.x01*b.x12 + t.x02*y22 + t.x03*b.x32
	m.x12 = t.x10*b.x02 + x11*b.x12 + t.x12*y22 + t.x13*b.x32
	m.d22 = t.x20*b.x02 + t.x21*b.x12 + x22*y22 + t.x23*b.x32 - 1
	m.x32 = t.x30*b.x02 + t.x31*b.x12 + t.x32*y22 + x33*b.x32
	m.x03 = x00*b.x03 + t.x01*b.x13 + t.x02*b.x23 + t.x03*y33
	m.x13 = t.x10*b.x03 + x11*b.x13 + t.x12*b.x23 + t.x13*y33
	m.x23 = t.x20*b.x03 + t.x21*b.x13 + x22*b.x23 + t.x23*y33
	m.d33 = t.x30*b.x03 + t.x31*b.x13 + t.x32*b.x23 + x33*y33 - 1
	return m
}

// Det returns the determinant of the Transform.
func (t Transform) Det() float64 {
	x00 := t.d00 + 1
	x11 := t.d11 + 1
	x22 := t.d22 + 1
	x33 := t.d33 + 1
	return x00*x11*x22*x33 - x00*x11*t.x23*t.x32 +
		x00*t.x12*t.x23*t.x31 - x00*t.x12*t.x21*x33 +
		x00*t.x13*t.x21*t.x32 - x00*t.x13*x22*t.x31 -
		t.x01*t.x12*t.x23*t.x30 + t.x01*t.x12*t.x20*x33 -
		t.x01*t.x13*t.x20*t.x32 + t.x01*t.x13*x22*t.x30 -
		t.x01*t.x10*x22*x33 + t.x01*t.x10*t.x23*t.x32 +
		t.x02*t.x13*t.x20*t.x31 - t.x02*t.x13*t.x21*t.x30 +
		t.x02*t.x10*t.x21*x33 - t.x02*t.x10*t.x23*t.x31 +
		t.x02*x11*t.x23*t.x30 - t.x02*x11*t.x20*x33 -
		t.x03*t.x10*t.x21*t.x32 + t.x03*t.x10*x22*t.x31 -
		t.x03*x11*x22*t.x30 + t.x03*x11*t.x20*t.x32 -
		t.x03*t.x12*t.x20*t.x31 + t.x03*t.x12*t.x21*t.x30
}

// Inv returns the inverse of the transform such that
// t.Inv() * t is the identity Transform.
// If matrix is singular then Inv() returns the zero transform.
func (t Transform) Inv() Transform {
	if t == (Transform{}) {
		return t
	}
	det := t.Det()
	if math.Abs(det) < 1e-16 {
		return zeroTransform
	}
	// Do something if singular?
	d := 1 / det
	x00 := t.d00 + 1
	x11 := t.d11 + 1
	x22 := t.d22 + 1
	x33 := t.d33 + 1
	var m Transform
	m.d00 = (t.x12*t.x23*t.x31-t.x13*x22*t.x31+t.x13*t.x21*t.x32-x11*t.x23*t.x32-t.x12*t.x21*x33+x11*x22*x33)*d - 1
	m.x01 = (t.x03*x22*t.x31 - t.x02*t.x23*t.x31 - t.x03*t.x21*t.x32 + t.x01*t.x23*t.x32 + t.x02*t.x21*x33 - t.x01*x22*x33) * d
	m.x02 = (t.x02*t.x13*t.x31 - t.x03*t.x12*t.x31 + t.x03*x11*t.x32 - t.x01*t.x13*t.x32 - t.x02*x11*x33 + t.x01*t.x12*x33) * d
	m.x03 = (t.x03*t.x12*t.x21 - t.x02*t.x13*t.x21 - t.x03*x11*x22 + t.x01*t.x13*x22 + t.x02*x11*t.x23 - t.x01*t.x12*t.x23) * d
	m.x10 = (t.x13*x22*t.x30 - t.x12*t.x23*t.x30 - t.x13*t.x20*t.x32 + t.x10*t.x23*t.x32 + t.x12*t.x20*x33 - t.x10*x22*x33) * d
	m.d11 = (t.x02*t.x23*t.x30-t.x03*x22*t.x30+t.x03*t.x20*t.x32-x00*t.x23*t.x32-t.x02*t.x20*x33+x00*x22*x33)*d - 1
	m.x12 = (t.x03*t.x12*t.x30 - t.x02*t.x13*t.x30 - t.x03*t.x10*t.x32 + x00*t.x13*t.x32 + t.x02*t.x10*x33 - x00*t.x12*x33) * d
	m.x13 = (t.x02*t.x13*t.x20 - t.x03*t.x12*t.x20 + t.x03*t.x10*x22 - x00*t.x13*x22 - t.x02*t.x10*t.x23 + x00*t.x12*t.x23) * d
	m.x20 = (x11*t.x23*t.x30 - t.x13*t.x21*t.x30 + t.x13*t.x20*t.x31 - t.x10*t.x23*t.x31 - x11*t.x20*x33 + t.x10*t.x21*x33) * d
	m.x21 = (t.x03*t.x21*t.x30 - t.x01*t.x23*t.x30 - t.x03*t.x20*t.x31 + x00*t.x23*t.x31 + t.x01*t.x20*x33 - x00*t.x21*x33) * d
	m.d22 = (t.x01*t.x13*t.x30-t.x03*x11*t.x30+t.x03*t.x10*t.x31-x00*t.x13*t.x31-t.x01*t.x10*x33+x00*x11*x33)*d - 1
	m.x23 = (t.x03*x11*t.x20 - t.x01*t.x13*t.x20 - t.x03*t.x10*t.x21 + x00*t.x13*t.x21 + t.x01*t.x10*t.x23 - x00*x11*t.x23) * d
	m.x30 = (t.x12*t.x21*t.x30 - x11*x22*t.x30 - t.x12*t.x20*t.x31 + t.x10*x22*t.x31 + x11*t.x20*t.x32 - t.x10*t.x21*t.x32) * d
	m.x31 = (t.x01*x22*t.x30 - t.x02*t.x21*t.x30 + t.x02*t.x20*t.x31 - x00*x22*t.x31 - t.x01*t.x20*t.x32 + x00*t.x21*t.x32) * d
	m.x32 = (t.x02*x11*t.x30 - t.x01*t.x12*t.x30 - t.x02*t.x10*t.x31 + x00*t.x12*t.x31 + t.x01*t.x10*t.x32 - x00*x11*t.x32) * d
	m.d33 = (t.x01*t.x12*t.x20-t.x02*x11*t.x20+t.x02*t.x10*t.x21-x00*t.x12*t.x21-t.x01*t.x10*x22+x00*x11*x22)*d - 1
	return m
}

func (t Transform) transpose() Transform {
	return Transform{
		d00: t.d00, x01: t.x10, x02: t.x20, x03: t.x30,
		x10: t.x01, d11: t.d11, x12: t.x21, x13: t.x31,
		x20: t.x02, x21: t.x12, d22: t.d22, x23: t.x32,
		x30: t.x03, x31: t.x13, x32: t.x23, d33: t.d33,
	}
}

// equals tests the equality of the Transforms to within a tolerance.
func (t Transform) equals(b Transform, tolerance float64) bool {
	return math.Abs(t.d00-b.d00) < tolerance &&
		math.Abs(t.x01-b.x01) < tolerance &&
		math.Abs(t.x02-b.x02) < tolerance &&
		math.Abs(t.x03-b.x03) < tolerance &&
		math.Abs(t.x10-b.x10) < tolerance &&
		math.Abs(t.d11-b.d11) < tolerance &&
		math.Abs(t.x12-b.x12) < tolerance &&
		math.Abs(t.x13-b.x13) < tolerance &&
		math.Abs(t.x20-b.x20) < tolerance &&
		math.Abs(t.x21-b.x21) < tolerance &&
		math.Abs(t.d22-b.d22) < tolerance &&
		math.Abs(t.x23-b.x23) < tolerance &&
		math.Abs(t.x30-b.x30) < tolerance &&
		math.Abs(t.x31-b.x31) < tolerance &&
		math.Abs(t.x32-b.x32) < tolerance &&
		math.Abs(t.d33-b.d33) < tolerance
}

// SliceCopy returns a copy of the Transform's data
// in row major storage format. It returns 16 elements.
func (t Transform) SliceCopy() []float64 {
	return []float64{
		t.d00 + 1, t.x01, t.x02, t.x03,
		t.x10, t.d11 + 1, t.x12, t.x13,
		t.x20, t.x21, t.d22 + 1, t.x23,
		t.x30, t.x31, t.x32, t.d33 + 1,
	}
}

func (a Transform) ApplyTriangle(b Triangle) Triangle {
	for i := range b {
		b[i] = a.Transform(b[i])
	}
	return b
}

// ApplyBox rotates/translates a 3d bounding box and resizes for axis-alignment.
// func (a Transform) ApplyBox(box Box) Box {

// 	r := Vec{X: a.x00, Y: a.x10, Z: a.x20}
// 	u := Vec{X: a.x01, Y: a.x11, Z: a.x21}
// 	b := Vec{X: a.x02, Y: a.x12, Z: a.x22}
// 	t := Vec{X: a.x03, Y: a.x13, Z: a.x23}

// 	xa := Scale(box.Min.X, r)
// 	xb := Scale(box.Max.X, r)
// 	ya := Scale(box.Min.Y, u)
// 	yb := Scale(box.Max.Y, u)
// 	za := Scale(box.Min.Z, b)
// 	zb := Scale(box.Max.Z, b)
// 	xa, xb = minElem(xa, xb), maxElem(xa, xb)
// 	ya, yb = minElem(ya, yb), maxElem(ya, yb)
// 	za, zb = minElem(za, zb), maxElem(za, zb)
// 	min := Add(Add(xa, ya), Add(za, t))
// 	max := Add(Add(xb, yb), Add(zb, t))

// 	return Box{min, max}
// }

// rotateToVector returns the rotation matrix that transforms a onto the same direction as b.
func rotateToVec(a, b Vec) Transform {
	const epsilon = 1e-12
	// is either vector == 0?
	if EqualWithin(a, Vec{}, epsilon) || EqualWithin(b, Vec{}, epsilon) {
		return Transform{}
	}
	// normalize both vectors
	a = Unit(a)
	b = Unit(b)
	// are the vectors the same?
	if EqualWithin(a, b, epsilon) {
		return Transform{}
	}

	// are the vectors opposite (180 degrees apart)?
	if EqualWithin(Scale(-1, a), b, epsilon) {
		return Transform{
			-1, 0, 0, 0,
			0, -1, 0, 0,
			0, 0, -1, 0,
			0, 0, 0, 1,
		}
	}
	// general case
	// See:	https://math.stackexchange.com/questions/180418/calculate-rotation-matrix-to-align-vector-a-to-vector-b-in-3d
	v := Cross(a, b)
	vx := Skew(v)

	k := 1 / (1 + Dot(a, b))
	vx2 := NewMat(nil)
	vx2.Mul(vx, vx)
	vx2.Scale(k, vx2)

	// Calculate sum of matrices.
	vx.Add(vx, Eye())
	vx.Add(vx, vx2)
	return Transform{
		vx.At(0, 0), vx.At(0, 1), vx.At(0, 2), 0,
		vx.At(1, 0), vx.At(1, 1), vx.At(1, 2), 0,
		vx.At(2, 0), vx.At(2, 1), vx.At(2, 2), 0,
		0, 0, 0, 1,
	}
}

/* Does same thing as rotateToVector
func RotateToVector(a, b Vec) Transform {
	const epsilon = 1e-12
	// is either vector == 0?
	if EqualWithin(a, Vec{}, epsilon) || EqualWithin(b, Vec{}, epsilon) {
		return Identity3d()
	}
	// normalize both vectors
	a = Unit(a)
	b = Unit(b)
	// are the vectors the same?
	if EqualWithin(a, b, epsilon) {
		return Identity3d()
	}
	// are the vectors opposite (180 degrees apart)?
	if EqualWithin(Scale(-1, a), b, epsilon) {
		return Transform{
			-1, 0, 0, 0,
			0, -1, 0, 0,
			0, 0, -1, 0,
			0, 0, 0, 1}
	}

	// general case
	// See:	https://math.stackexchange.com/questions/180418/calculate-rotation-matrix-to-align-vector-a-to-vector-b-in-3d
	v := Cross(a, b)
	k := 1 / (1 + Dot(a, b))
	vx := M33{0, -v.Z, v.Y, v.Z, 0, -v.X, -v.Y, v.X, 0}
	eye := M33{1, 0, 0, 0, 1, 0, 0, 0, 1}
	r := eye.Add(vx).Add(vx.Mul(vx).MulScalar(k))
	return Transform{
		r.x00, r.x01, r.x02, 0,
		r.x10, r.x11, r.x12, 0,
		r.x20, r.x21, r.x22, 0,
		0, 0, 0, 1,
	}
}

// M33 is a 3x3 matrix.
type M33 struct {
	x00, x01, x02 float64
	x10, x11, x12 float64
	x20, x21, x22 float64
}

// Mul multiplies 3x3 matrices.
func (a M33) Mul(b M33) M33 {
	m := M33{}
	m.x00 = a.x00*b.x00 + a.x01*b.x10 + a.x02*b.x20
	m.x10 = a.x10*b.x00 + a.x11*b.x10 + a.x12*b.x20
	m.x20 = a.x20*b.x00 + a.x21*b.x10 + a.x22*b.x20
	m.x01 = a.x00*b.x01 + a.x01*b.x11 + a.x02*b.x21
	m.x11 = a.x10*b.x01 + a.x11*b.x11 + a.x12*b.x21
	m.x21 = a.x20*b.x01 + a.x21*b.x11 + a.x22*b.x21
	m.x02 = a.x00*b.x02 + a.x01*b.x12 + a.x02*b.x22
	m.x12 = a.x10*b.x02 + a.x11*b.x12 + a.x12*b.x22
	m.x22 = a.x20*b.x02 + a.x21*b.x12 + a.x22*b.x22
	return m
}

// Add two 3x3 matrices.
func (a M33) Add(b M33) M33 {
	return M33{
		x00: a.x00 + b.x00,
		x10: a.x10 + b.x10,
		x20: a.x20 + b.x20,
		x01: a.x01 + b.x01,
		x11: a.x11 + b.x11,
		x21: a.x21 + b.x21,
		x02: a.x02 + b.x02,
		x12: a.x12 + b.x12,
		x22: a.x22 + b.x22,
	}
}

// MulScalar multiplies each 3x3 matrix component by a scalar.
func (a M33) MulScalar(k float64) M33 {
	return M33{
		x00: k * a.x00,
		x10: k * a.x10,
		x20: k * a.x20,
		x01: k * a.x01,
		x11: k * a.x11,
		x21: k * a.x21,
		x02: k * a.x02,
		x12: k * a.x12,
		x22: k * a.x22,
	}
}
*/
