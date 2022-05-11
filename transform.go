package main

import (
	"math"
)

// Transform is a 4x4 matrix.
type Transform struct {
	x00, x01, x02, x03 float64
	x10, x11, x12, x13 float64
	x20, x21, x22, x23 float64
	x30, x31, x32, x33 float64
}

// Identity3d returns a 4x4 identity matrix.
func Identity3d() Transform {
	return Transform{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}
}

func (t *Transform) SetTranslate(v Vec) {
	t.x03 = v.X
	t.x13 = v.Y
	t.x23 = v.Z
}

// Translate3D returns a 4x4 translation matrix.
func Translate(v Vec) Transform {
	return Transform{
		1, 0, 0, v.X,
		0, 1, 0, v.Y,
		0, 0, 1, v.Z,
		0, 0, 0, 1}
}

// Scale3d returns a 4x4 scaling matrix.
// Scaling does not preserve distance. See: ScaleUniform3D()
func Scale3d(v Vec) Transform {
	return Transform{
		v.X, 0, 0, 0,
		0, v.Y, 0, 0,
		0, 0, v.Z, 0,
		0, 0, 0, 1}
}

// Rotate3d returns an orthographic 4x4 rotation matrix (right hand rule).
func Rotate3d(v Vec, a float64) Transform {
	v = Unit(v)
	s, c := math.Sincos(a)
	m := 1 - c
	return Transform{
		m*v.X*v.X + c, m*v.X*v.Y - v.Z*s, m*v.Z*v.X + v.Y*s, 0,
		m*v.X*v.Y + v.Z*s, m*v.Y*v.Y + c, m*v.Y*v.Z - v.X*s, 0,
		m*v.Z*v.X - v.Y*s, m*v.Y*v.Z + v.X*s, m*v.Z*v.Z + c, 0,
		0, 0, 0, 1,
	}
}

// RotateX returns a 4x4 matrix with rotation about the X axis.
func RotateX(a float64) Transform {
	return Rotate3d(Vec{X: 1, Y: 0, Z: 0}, a)
}

// RotateY returns a 4x4 matrix with rotation about the Y axis.
func RotateY(a float64) Transform {
	return Rotate3d(Vec{X: 0, Y: 1, Z: 0}, a)
}

// RotateZ returns a 4x4 matrix with rotation about the Z axis.
func RotateZ(a float64) Transform {
	return Rotate3d(Vec{X: 0, Y: 0, Z: 1}, a)
}

// MirrorXY returns a 4x4 matrix with mirroring across the XY plane.
func MirrorXY() Transform {
	return Transform{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, -1, 0,
		0, 0, 0, 1}
}

// MirrorXZ returns a 4x4 matrix with mirroring across the XZ plane.
func MirrorXZ() Transform {
	return Transform{
		1, 0, 0, 0,
		0, -1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}
}

// MirrorYZ returns a 4x4 matrix with mirroring across the YZ plane.
func MirrorYZ() Transform {
	return Transform{
		-1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}
}

// MirrorXeqY returns a 4x4 matrix with mirroring across the X == Y plane.
func MirrorXeqY() Transform {
	return Transform{
		0, 1, 0, 0,
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1}
}

// equals tests the equality of 4x4 matrices.
func (a Transform) equals(b Transform, tolerance float64) bool {
	return (math.Abs(a.x00-b.x00) < tolerance &&
		math.Abs(a.x01-b.x01) < tolerance &&
		math.Abs(a.x02-b.x02) < tolerance &&
		math.Abs(a.x03-b.x03) < tolerance &&
		math.Abs(a.x10-b.x10) < tolerance &&
		math.Abs(a.x11-b.x11) < tolerance &&
		math.Abs(a.x12-b.x12) < tolerance &&
		math.Abs(a.x13-b.x13) < tolerance &&
		math.Abs(a.x20-b.x20) < tolerance &&
		math.Abs(a.x21-b.x21) < tolerance &&
		math.Abs(a.x22-b.x22) < tolerance &&
		math.Abs(a.x23-b.x23) < tolerance &&
		math.Abs(a.x30-b.x30) < tolerance &&
		math.Abs(a.x31-b.x31) < tolerance &&
		math.Abs(a.x32-b.x32) < tolerance &&
		math.Abs(a.x33-b.x33) < tolerance)
}

// Mul multiplies 4x4 matrices.
func (a Transform) Mul(b Transform) Transform {
	m := Transform{}
	m.x00 = a.x00*b.x00 + a.x01*b.x10 + a.x02*b.x20 + a.x03*b.x30
	m.x10 = a.x10*b.x00 + a.x11*b.x10 + a.x12*b.x20 + a.x13*b.x30
	m.x20 = a.x20*b.x00 + a.x21*b.x10 + a.x22*b.x20 + a.x23*b.x30
	m.x30 = a.x30*b.x00 + a.x31*b.x10 + a.x32*b.x20 + a.x33*b.x30
	m.x01 = a.x00*b.x01 + a.x01*b.x11 + a.x02*b.x21 + a.x03*b.x31
	m.x11 = a.x10*b.x01 + a.x11*b.x11 + a.x12*b.x21 + a.x13*b.x31
	m.x21 = a.x20*b.x01 + a.x21*b.x11 + a.x22*b.x21 + a.x23*b.x31
	m.x31 = a.x30*b.x01 + a.x31*b.x11 + a.x32*b.x21 + a.x33*b.x31
	m.x02 = a.x00*b.x02 + a.x01*b.x12 + a.x02*b.x22 + a.x03*b.x32
	m.x12 = a.x10*b.x02 + a.x11*b.x12 + a.x12*b.x22 + a.x13*b.x32
	m.x22 = a.x20*b.x02 + a.x21*b.x12 + a.x22*b.x22 + a.x23*b.x32
	m.x32 = a.x30*b.x02 + a.x31*b.x12 + a.x32*b.x22 + a.x33*b.x32
	m.x03 = a.x00*b.x03 + a.x01*b.x13 + a.x02*b.x23 + a.x03*b.x33
	m.x13 = a.x10*b.x03 + a.x11*b.x13 + a.x12*b.x23 + a.x13*b.x33
	m.x23 = a.x20*b.x03 + a.x21*b.x13 + a.x22*b.x23 + a.x23*b.x33
	m.x33 = a.x30*b.x03 + a.x31*b.x13 + a.x32*b.x23 + a.x33*b.x33
	return m
}

// Transform bounding boxes - keep them axis aligned
// http://dev.theomader.com/transform-bounding-boxes/

// ApplyPosition multiplies a Vec position with a rotate/translate matrix.
func (a Transform) ApplyPosition(b Vec) Vec {
	return Vec{
		X: a.x00*b.X + a.x01*b.Y + a.x02*b.Z + a.x03,
		Y: a.x10*b.X + a.x11*b.Y + a.x12*b.Z + a.x13,
		Z: a.x20*b.X + a.x21*b.Y + a.x22*b.Z + a.x23}
}

func (a Transform) ApplyTriangle(b Triangle) Triangle {
	for i := range b {
		b[i] = a.ApplyPosition(b[i])
	}
	return b
}

// ApplyBox rotates/translates a 3d bounding box and resizes for axis-alignment.
func (a Transform) ApplyBox(box Box) Box {
	r := Vec{X: a.x00, Y: a.x10, Z: a.x20}
	u := Vec{X: a.x01, Y: a.x11, Z: a.x21}
	b := Vec{X: a.x02, Y: a.x12, Z: a.x22}
	t := Vec{X: a.x03, Y: a.x13, Z: a.x23}

	xa := Scale(box.Min.X, r)
	xb := Scale(box.Max.X, r)
	ya := Scale(box.Min.Y, u)
	yb := Scale(box.Max.Y, u)
	za := Scale(box.Min.Z, b)
	zb := Scale(box.Max.Z, b)
	xa, xb = minElem(xa, xb), maxElem(xa, xb)
	ya, yb = minElem(ya, yb), maxElem(ya, yb)
	za, zb = minElem(za, zb), maxElem(za, zb)
	min := Add(Add(xa, ya), Add(za, t))
	max := Add(Add(xb, yb), Add(zb, t))

	return Box{min, max}
}

// Determinant returns the determinant of a 4x4 matrix.
func (a Transform) Determinant() float64 {
	return (a.x00*a.x11*a.x22*a.x33 - a.x00*a.x11*a.x23*a.x32 +
		a.x00*a.x12*a.x23*a.x31 - a.x00*a.x12*a.x21*a.x33 +
		a.x00*a.x13*a.x21*a.x32 - a.x00*a.x13*a.x22*a.x31 -
		a.x01*a.x12*a.x23*a.x30 + a.x01*a.x12*a.x20*a.x33 -
		a.x01*a.x13*a.x20*a.x32 + a.x01*a.x13*a.x22*a.x30 -
		a.x01*a.x10*a.x22*a.x33 + a.x01*a.x10*a.x23*a.x32 +
		a.x02*a.x13*a.x20*a.x31 - a.x02*a.x13*a.x21*a.x30 +
		a.x02*a.x10*a.x21*a.x33 - a.x02*a.x10*a.x23*a.x31 +
		a.x02*a.x11*a.x23*a.x30 - a.x02*a.x11*a.x20*a.x33 -
		a.x03*a.x10*a.x21*a.x32 + a.x03*a.x10*a.x22*a.x31 -
		a.x03*a.x11*a.x22*a.x30 + a.x03*a.x11*a.x20*a.x32 -
		a.x03*a.x12*a.x20*a.x31 + a.x03*a.x12*a.x21*a.x30)
}

// Inverse returns the inverse of a 4x4 matrix.
func (a Transform) Inverse() Transform {
	m := Transform{}
	d := 1 / a.Determinant()
	m.x00 = (a.x12*a.x23*a.x31 - a.x13*a.x22*a.x31 + a.x13*a.x21*a.x32 - a.x11*a.x23*a.x32 - a.x12*a.x21*a.x33 + a.x11*a.x22*a.x33) * d
	m.x01 = (a.x03*a.x22*a.x31 - a.x02*a.x23*a.x31 - a.x03*a.x21*a.x32 + a.x01*a.x23*a.x32 + a.x02*a.x21*a.x33 - a.x01*a.x22*a.x33) * d
	m.x02 = (a.x02*a.x13*a.x31 - a.x03*a.x12*a.x31 + a.x03*a.x11*a.x32 - a.x01*a.x13*a.x32 - a.x02*a.x11*a.x33 + a.x01*a.x12*a.x33) * d
	m.x03 = (a.x03*a.x12*a.x21 - a.x02*a.x13*a.x21 - a.x03*a.x11*a.x22 + a.x01*a.x13*a.x22 + a.x02*a.x11*a.x23 - a.x01*a.x12*a.x23) * d
	m.x10 = (a.x13*a.x22*a.x30 - a.x12*a.x23*a.x30 - a.x13*a.x20*a.x32 + a.x10*a.x23*a.x32 + a.x12*a.x20*a.x33 - a.x10*a.x22*a.x33) * d
	m.x11 = (a.x02*a.x23*a.x30 - a.x03*a.x22*a.x30 + a.x03*a.x20*a.x32 - a.x00*a.x23*a.x32 - a.x02*a.x20*a.x33 + a.x00*a.x22*a.x33) * d
	m.x12 = (a.x03*a.x12*a.x30 - a.x02*a.x13*a.x30 - a.x03*a.x10*a.x32 + a.x00*a.x13*a.x32 + a.x02*a.x10*a.x33 - a.x00*a.x12*a.x33) * d
	m.x13 = (a.x02*a.x13*a.x20 - a.x03*a.x12*a.x20 + a.x03*a.x10*a.x22 - a.x00*a.x13*a.x22 - a.x02*a.x10*a.x23 + a.x00*a.x12*a.x23) * d
	m.x20 = (a.x11*a.x23*a.x30 - a.x13*a.x21*a.x30 + a.x13*a.x20*a.x31 - a.x10*a.x23*a.x31 - a.x11*a.x20*a.x33 + a.x10*a.x21*a.x33) * d
	m.x21 = (a.x03*a.x21*a.x30 - a.x01*a.x23*a.x30 - a.x03*a.x20*a.x31 + a.x00*a.x23*a.x31 + a.x01*a.x20*a.x33 - a.x00*a.x21*a.x33) * d
	m.x22 = (a.x01*a.x13*a.x30 - a.x03*a.x11*a.x30 + a.x03*a.x10*a.x31 - a.x00*a.x13*a.x31 - a.x01*a.x10*a.x33 + a.x00*a.x11*a.x33) * d
	m.x23 = (a.x03*a.x11*a.x20 - a.x01*a.x13*a.x20 - a.x03*a.x10*a.x21 + a.x00*a.x13*a.x21 + a.x01*a.x10*a.x23 - a.x00*a.x11*a.x23) * d
	m.x30 = (a.x12*a.x21*a.x30 - a.x11*a.x22*a.x30 - a.x12*a.x20*a.x31 + a.x10*a.x22*a.x31 + a.x11*a.x20*a.x32 - a.x10*a.x21*a.x32) * d
	m.x31 = (a.x01*a.x22*a.x30 - a.x02*a.x21*a.x30 + a.x02*a.x20*a.x31 - a.x00*a.x22*a.x31 - a.x01*a.x20*a.x32 + a.x00*a.x21*a.x32) * d
	m.x32 = (a.x02*a.x11*a.x30 - a.x01*a.x12*a.x30 - a.x02*a.x10*a.x31 + a.x00*a.x12*a.x31 + a.x01*a.x10*a.x32 - a.x00*a.x11*a.x32) * d
	m.x33 = (a.x01*a.x12*a.x20 - a.x02*a.x11*a.x20 + a.x02*a.x10*a.x21 - a.x00*a.x12*a.x21 - a.x01*a.x10*a.x22 + a.x00*a.x11*a.x22) * d
	return m
}

// rotateToVector returns the rotation matrix that transforms a onto the same direction as b.
func rotateToVec(a, b Vec) Transform {
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
