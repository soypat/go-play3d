//go:build !js

package main

import (
	"fmt"
	"sort"

	"gonum.org/v1/gonum/mat"
)

//go:generate go run .

func main() {
	nodes, elems := feaModel()
	// Fiber compliance matrix
	// Cf := orthotropicCompliance(235e3, 14e3, 0.2, 0.25, 28e3)
	// composite filler material compliance matrix
	Cm := isotropicCompliance(4.8e3, 0.34)
	// Calculate Gauss integration points and form functions
	// evaluated at Gauss points.
	upg, wpg := gauss3D(2, 2, 2)
	N := make([]*mat.VecDense, len(upg))
	dN := make([]*mat.Dense, len(upg))
	for ipg, pg := range upg {
		N[ipg] = mat.NewVecDense(8, h8FormFuncs(pg.X, pg.Y, pg.Z))
		dN[ipg] = mat.NewDense(3, 8, h8FormFuncsDiff(pg.X, pg.Y, pg.Z))
	}
	jac := NewMat(nil)
	enod := make([]Vec, 8)
	edofs := make([]int, 3*8)
	dNxyz := mat.NewDense(3, 8, nil)
	B := mat.NewDense(6, 3*8, nil) // number of columns in Compliance x NdofPerNode*nodesperelement
	Ke := mat.NewDense(3*8, 3*8, nil)
	aux1 := mat.NewDense(3*8, 6, nil)
	aux2 := mat.NewDense(3*8, 3*8, nil)
	K := mat.NewDense(3*len(nodes), 3*len(nodes), nil)
	for iele := range elems {
		Ke.Zero()
		enodi := elems[iele][:]
		storeElemNode(enod, nodes, enodi)
		storeElemDofs(edofs, enodi, 3)
		for ipg := range upg {
			jac.Mul(dN[ipg], denseFromR3(enod))
			dNxyz.Solve(jac, dN[ipg])
			for i := 0; i < 8; i++ {
				// First three rows.
				B.Set(0, i*3, dNxyz.At(0, i))
				B.Set(1, i*3+1, dNxyz.At(1, i))
				B.Set(2, i*3+2, dNxyz.At(2, i))
				// Fourth row.
				B.Set(3, i*3, dNxyz.At(1, i))
				B.Set(3, i*3+1, dNxyz.At(0, i))
				// Fifth row.
				B.Set(4, i*3+1, dNxyz.At(2, i))
				B.Set(4, i*3+2, dNxyz.At(1, i))
				// Sixth row.
				B.Set(5, i*3, dNxyz.At(2, i))
				B.Set(5, i*3+2, dNxyz.At(0, i))
			}
			aux1.Mul(B.T(), Cm)
			aux2.Mul(aux1, B)
			aux2.Scale(jac.Det()*wpg[ipg], aux2)
			Ke.Add(Ke, aux2)
		}
		r, c := Ke.Dims()
		for i := 0; i < r; i++ {
			ei := edofs[i]
			for j := 0; j < c; j++ {
				ej := edofs[j]
				K.Set(ei, ej, K.At(ei, ej)+Ke.At(i, j))
			}
		}
	}
	e := mat.Eigen{}
	ok := e.Factorize(Ke, mat.EigenBoth)
	eigs := e.Values(nil)
	sort.Sort(byRealMagnitude(eigs))
	fmt.Println(ok, eigs)
	fmt.Printf("%f", mat.Formatted(K, mat.FormatMATLAB()))
}

// func convertToRenderTriangles(t []Triangle) []render.Triangle3 {
// 	return *(*[]render.Triangle3)(unsafe.Pointer(&t))
// 	// c := make([]render.Triangle3, len(t))
// 	// for i := range t {
// 	// 	c[i] = render.Triangle3(t[i])
// 	// }
// 	// return c
// }
