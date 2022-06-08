package main

import (
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestBooleanIndexing(t *testing.T) {
	for _, test := range []struct {
		m, expect mat.Matrix
		br, bc    []bool
	}{
		{
			m:  mat.NewDense(2, 2, []float64{1, 2, 3, 4}),
			br: []bool{false, true}, bc: []bool{false, true},
			expect: mat.NewDense(1, 1, []float64{4}),
		},
		{
			m:  mat.NewDense(3, 3, []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}),
			br: []bool{false, true, true}, bc: []bool{false, true, true},
			expect: mat.NewDense(2, 2, []float64{5, 6, 8, 9}),
		},
	} {
		sm := booleanIndexing(test.m, false, test.br, test.bc)
		if !mat.Equal(test.expect, sm) {
			t.Error("matrices not equal")
		}
	}
}
