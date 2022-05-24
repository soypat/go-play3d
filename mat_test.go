package main

import "testing"

func TestMatEigs(t *testing.T) {
	m := NewMat([]float64{
		1, 2, 3,
		2, 4, 5,
		3, 5, 6,
	})
	t.Fatal(m.Eigs())
}
