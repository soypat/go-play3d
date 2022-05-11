package main

// line is an infinite 3D line
// defined by two points on the line.
type line [2]Vec

// interp takes a value between 0 and 1 to interpolate a point
// on the line.
//  interp(0) returns l[0]
//  interp(1) returns l[1]
func (l line) interp(t float64) Vec {
	lineDir := Sub(l[1], l[0])
	return Add(l[0], Scale(t, lineDir))
}

// equation returns the starting point of the line
// and the direction of the line. This can then be used
// to represent the line in vector, parametric or symmetric form.
func (l line) equation() (p0 Vec, n Vec) {
	// https://math.stackexchange.com/questions/404440/what-is-the-equation-for-a-3d-line
	return l[0], Sub(l[1], l[0])
}

// distance returns the minimum euclidean distance of point p
// to the line.
func (l line) distance(p Vec) float64 {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	num := Norm(Cross(Sub(p, l[0]), Sub(p, l[1])))
	return num / Norm(Sub(l[1], l[0]))
}

// closest returns the closest point on line l to point p.
func (l line) closest(p Vec) Vec {
	// https://mathworld.wolfram.com/Point-LineDistance3-Dimensional.html
	t := -Dot(Sub(l[0], p), Sub(l[1], p)) / Norm2(Sub(l[1], l[0]))
	return l.interp(t)
}
