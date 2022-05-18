package main

import (
	"math"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/spatial/r2"
)

func TestTriangleClosest(t *testing.T) {
	ico := icosphere(2)
	rnd := rand.New(rand.NewSource(1))
	randVec := func() Vec {
		v := Vec{rnd.Float64(), rnd.Float64(), rnd.Float64()}
		return Scale(rnd.Float64()*2, Unit(v))
	}
	for _, tri := range ico[:2] { // TODO: TEST ALL TRIANGLES WHEN THIS PASSES
		point := randVec()
		gotclosest := tri.Closest(point)
		gotDist := Norm(Sub(point, gotclosest))
		for i := 0; i < 10000; i++ {
			randTriPoint := tri.randomPoint(rnd)
			randDist := Norm(Sub(point, randTriPoint))
			if randDist < gotDist {
				t.Error("random point closer than result of Closest")
				break
			}
		}
	}
}

func TestTriangle2Closest(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	randVec := func(scale float64) r2.Vec {
		return r2.Scale(2*scale, r2.Vec{rnd.Float64() - .5, rnd.Float64() - .5})
		// return Scale(rnd.Float64()*2, r2.Unit(v))
	}
	randTriangle := func(scale float64) [3]r2.Vec {
		return [3]r2.Vec{randVec(scale), randVec(scale), randVec(scale)}
	}
	for ic := 0; ic < 100; ic++ {
		tri := randTriangle(1)
		point := randVec(3)
		ptxy, _ := closestOnTriangle2(point, tri)
		gotDist := r2.Norm(r2.Sub(ptxy, point))
		for i := 0; i < 1000; i++ {
			randPoint := randomPointOnTriangle2(rnd, tri)
			randDist := r2.Norm(r2.Sub(randPoint, point))
			if randDist < gotDist {
				t.Error("random point closer than result of closestOnTriangle2")
			}
		}
	}
}

func TestCanalisTransform(t *testing.T) {
	const tol = 1e-12
	ico := icosphere(2)
	// rnd := rand.New(rand.NewSource(1))
	// randVec := func() Vec {
	// 	v := Vec{rnd.Float64(), rnd.Float64(), rnd.Float64()}
	// 	return Scale(rnd.Float64()*2, Unit(v))
	// }
	mismatches := 0
	var worstAreaMismatch, worstZmismatch float64
	var worstArea, worstZ Triangle
	for _, tri := range ico {
		T := canalisTransform(tri)
		triT := T.ApplyTriangle(tri)
		wantArea := tri.Area()
		gotArea := triT.Area()
		// Area must stay the same.
		areaDiff := math.Abs(wantArea - gotArea)
		if areaDiff > tol {
			t.Error("area mismatches got/want", gotArea, wantArea)
			mismatches++
			if areaDiff > worstAreaMismatch {
				worstAreaMismatch = areaDiff
				worstArea = tri
			}
		}
		// Test first vertex at origin.
		if !EqualWithin(triT[0], Vec{}, tol) {
			t.Error("first vertex not at origin")
			mismatches++
		}
		// test first two vertices are on X axis
		xDir := Sub(triT[1], triT[0])
		xDirLen := Norm(xDir)
		if !EqualWithin(xDir, Vec{X: xDirLen}, tol) {
			t.Error("first edge not on x Axis")
			mismatches++
		}
		// Test last vertex in XY plane
		if math.Abs(triT[2].Z) > tol {
			t.Error("last edge z component too large", triT[2].Z)
			mismatches++
			if math.Abs(triT[2].Z) > worstZmismatch {
				worstZmismatch = math.Abs(triT[2].Z)
				worstZ = tri
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("worst area %.g %+v", worstAreaMismatch, worstArea)
		t.Errorf("worst Z mismatch %.g %+v", worstZmismatch, worstZ)
		t.Error("got number of mismatches", mismatches)
	}
}

func randomPointOnTriangle2(rnd *rand.Rand, t [3]r2.Vec) r2.Vec {
	a, b := rnd.Float64(), rnd.Float64()
	if a+b >= 1 {
		// Reduce from the quadrilateral case
		a = 1 - a
		b = 1 - b
	}
	return r2.Add(r2.Add(t[0], r2.Scale(b, r2.Sub(t[2], t[0]))), r2.Scale(a, r2.Sub(t[1], t[0])))
}
