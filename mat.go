// Copyright ©2021 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"math"

	"gonum.org/v1/gonum/floats/scalar"
	"gonum.org/v1/gonum/mat"
)

// Mat represents a 3×3 matrix. Useful for rotation matrices and such.
// The zero value is usable as the 3×3 zero matrix.
type Mat struct {
	data *array
}

var _ mat.Matrix = (*Mat)(nil)

// NewMat returns a new 3×3 matrix Mat type and populates its elements
// with values passed as argument in row-major form. If val argument
// is nil then NewMat returns a matrix filled with zeros.
func NewMat(val []float64) *Mat {
	if len(val) == 9 {
		return &Mat{arrayFrom(val)}
	}
	if val == nil {
		return &Mat{new(array)}
	}
	panic(mat.ErrShape)
}

// Dims returns the number of rows and columns of this matrix.
// This method will always return 3×3 for a Mat.
func (m *Mat) Dims() (r, c int) { return 3, 3 }

// T returns the transpose of Mat. Changes in the receiver will be reflected in the returned matrix.
func (m *Mat) T() mat.Matrix { return mat.Transpose{Matrix: m} }

// Scale multiplies the elements of a by f, placing the result in the receiver.
//
// See the mat.Scaler interface for more information.
func (m *Mat) Scale(f float64, a mat.Matrix) {
	r, c := a.Dims()
	if r != 3 || c != 3 {
		panic(mat.ErrShape)
	}
	if m.data == nil {
		m.data = new(array)
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m.Set(i, j, f*a.At(i, j))
		}
	}
}

// MulVec returns the matrix-vector product M⋅v.
func (m *Mat) MulVec(v Vec) Vec {
	if m.data == nil {
		return Vec{}
	}
	return Vec{
		X: v.X*m.At(0, 0) + v.Y*m.At(0, 1) + v.Z*m.At(0, 2),
		Y: v.X*m.At(1, 0) + v.Y*m.At(1, 1) + v.Z*m.At(1, 2),
		Z: v.X*m.At(2, 0) + v.Y*m.At(2, 1) + v.Z*m.At(2, 2),
	}
}

// MulVecTrans returns the matrix-vector product Mᵀ⋅v.
func (m *Mat) MulVecTrans(v Vec) Vec {
	if m.data == nil {
		return Vec{}
	}
	return Vec{
		X: v.X*m.At(0, 0) + v.Y*m.At(1, 0) + v.Z*m.At(2, 0),
		Y: v.X*m.At(0, 1) + v.Y*m.At(1, 1) + v.Z*m.At(2, 1),
		Z: v.X*m.At(0, 2) + v.Y*m.At(1, 2) + v.Z*m.At(2, 2),
	}
}

// CloneFrom makes a copy of a into the receiver m.
// Mat expects a 3×3 input matrix.
func (m *Mat) CloneFrom(a mat.Matrix) {
	r, c := a.Dims()
	if r != 3 || c != 3 {
		panic(mat.ErrShape)
	}
	if m.data == nil {
		m.data = new(array)
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m.Set(i, j, a.At(i, j))
		}
	}
}

// Sub subtracts the matrix b from a, placing the result in the receiver.
// Sub will panic if the two matrices do not have the same shape.
func (m *Mat) Sub(a, b mat.Matrix) {
	if r, c := a.Dims(); r != 3 || c != 3 {
		panic(mat.ErrShape)
	}
	if r, c := b.Dims(); r != 3 || c != 3 {
		panic(mat.ErrShape)
	}
	if m.data == nil {
		m.data = new(array)
	}

	m.Set(0, 0, a.At(0, 0)-b.At(0, 0))
	m.Set(0, 1, a.At(0, 1)-b.At(0, 1))
	m.Set(0, 2, a.At(0, 2)-b.At(0, 2))
	m.Set(1, 0, a.At(1, 0)-b.At(1, 0))
	m.Set(1, 1, a.At(1, 1)-b.At(1, 1))
	m.Set(1, 2, a.At(1, 2)-b.At(1, 2))
	m.Set(2, 0, a.At(2, 0)-b.At(2, 0))
	m.Set(2, 1, a.At(2, 1)-b.At(2, 1))
	m.Set(2, 2, a.At(2, 2)-b.At(2, 2))
}

// Add adds a and b element-wise, placing the result in the receiver. Add will panic if the two matrices do not have the same shape.
func (m *Mat) Add(a, b mat.Matrix) {
	if r, c := a.Dims(); r != 3 || c != 3 {
		panic(mat.ErrShape)
	}
	if r, c := b.Dims(); r != 3 || c != 3 {
		panic(mat.ErrShape)
	}
	if m.data == nil {
		m.data = new(array)
	}

	m.Set(0, 0, a.At(0, 0)+b.At(0, 0))
	m.Set(0, 1, a.At(0, 1)+b.At(0, 1))
	m.Set(0, 2, a.At(0, 2)+b.At(0, 2))
	m.Set(1, 0, a.At(1, 0)+b.At(1, 0))
	m.Set(1, 1, a.At(1, 1)+b.At(1, 1))
	m.Set(1, 2, a.At(1, 2)+b.At(1, 2))
	m.Set(2, 0, a.At(2, 0)+b.At(2, 0))
	m.Set(2, 1, a.At(2, 1)+b.At(2, 1))
	m.Set(2, 2, a.At(2, 2)+b.At(2, 2))
}

// VecRow returns the elements in the ith row of the receiver.
func (m *Mat) VecRow(i int) Vec {
	if i > 2 {
		panic(mat.ErrRowAccess)
	}
	if m.data == nil {
		return Vec{}
	}
	return Vec{X: m.At(i, 0), Y: m.At(i, 1), Z: m.At(i, 2)}
}

// VecCol returns the elements in the jth column of the receiver.
func (m *Mat) VecCol(j int) Vec {
	if j > 2 {
		panic(mat.ErrColAccess)
	}
	if m.data == nil {
		return Vec{}
	}
	return Vec{X: m.At(0, j), Y: m.At(1, j), Z: m.At(2, j)}
}

// Outer calculates the outer product of the vectors x and y,
// where x and y are treated as column vectors, and stores the result in the receiver.
//  m = alpha * x * yᵀ
func (m *Mat) Outer(alpha float64, x, y Vec) {
	ax := alpha * x.X
	ay := alpha * x.Y
	az := alpha * x.Z
	m.Set(0, 0, ax*y.X)
	m.Set(0, 1, ax*y.Y)
	m.Set(0, 2, ax*y.Z)

	m.Set(1, 0, ay*y.X)
	m.Set(1, 1, ay*y.Y)
	m.Set(1, 2, ay*y.Z)

	m.Set(2, 0, az*y.X)
	m.Set(2, 1, az*y.Y)
	m.Set(2, 2, az*y.Z)
}

// Det calculates the determinant of the receiver using the following formula
//      ⎡a b c⎤
//  m = ⎢d e f⎥
//      ⎣g h i⎦
//  det(m) = a(ei − fh) − b(di − fg) + c(dh − eg)
func (m *Mat) Det() float64 {
	a := m.At(0, 0)
	b := m.At(0, 1)
	c := m.At(0, 2)

	deta := m.At(1, 1)*m.At(2, 2) - m.At(1, 2)*m.At(2, 1)
	detb := m.At(1, 0)*m.At(2, 2) - m.At(1, 2)*m.At(2, 0)
	detc := m.At(1, 0)*m.At(2, 1) - m.At(1, 1)*m.At(2, 0)
	return a*deta - b*detb + c*detc
}

func (m *Mat) Eigs() (r, c []float64) {
	const tol = 1e-12
	if !scalar.EqualWithinAbs(m.At(0, 1), m.At(1, 0), tol) ||
		!scalar.EqualWithinAbs(m.At(1, 2), m.At(2, 1), tol) ||
		!scalar.EqualWithinAbs(m.At(0, 2), m.At(2, 0), tol) {
		panic("non-symmetric eig algorithm not implemented")
	}
	// 3*m = tr(A)
	M := (m.At(0, 0) + m.At(1, 1) + m.At(2, 2)) / 3
	// Calculate  2*q=det(A-m*I)
	nm := NewMat(nil)
	nm.Scale(M, Eye())
	nm.Sub(m, nm)
	q := nm.Det() / 2
	// 6*p = sum of squares of elements of A-m*I
	var p float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			v := nm.At(i, j)
			p += v / 6 * v
		}
	}
	if math.Abs(p) < tol && math.Abs(q) < tol {
		// p == q == 0
		return []float64{M, M, M}, nil
	}
	// sqrt(3)
	const sqrt3 = 1.7320508075688772935274463415058723669428052538103806280558069794
	// phi = 1/3 atan( sqrt(p^3 - q^2)/q ), 0<=phi<=pi
	phi := math.Atan(math.Sqrt(p*p*p-q*q)/q) / 3
	sp, cp := math.Sincos(phi)
	sqrtp := math.Sqrt(p)
	return []float64{
		M + 2*sqrtp*cp,
		M - sqrtp*(cp+sqrt3*sp),
		M - sqrtp*(cp-sqrt3*sp),
	}, nil
}

func (m *Mat) Hessian(p Vec, tol float64, f func(Vec) float64) {
	h2 := tol * tol * 4
	dx := Vec{X: tol}
	dy := Vec{Y: tol}
	dz := Vec{Z: tol}
	fp := f(p)
	diff2 := func(p, d1, d2 Vec, f func(p Vec) float64) float64 {
		return (f(Add(p, Add(d1, d2))) - f(Add(p, d2)) - f(Add(p, d1)) + fp) / h2
	}
	fxx := diff2(p, dx, dx, f)
	fyy := diff2(p, dy, dy, f)
	fzz := diff2(p, dz, dz, f)
	fxy := diff2(p, dx, dy, f)
	fxz := diff2(p, dx, dz, f)
	fyz := diff2(p, dy, dz, f)
	m.Set(0, 0, fxx)
	m.Set(0, 1, fxy)
	m.Set(0, 2, fxz)

	m.Set(1, 0, fxy)
	m.Set(1, 1, fyy)
	m.Set(1, 2, fyz)

	m.Set(2, 0, fxz)
	m.Set(2, 1, fyz)
	m.Set(2, 2, fzz)
}
