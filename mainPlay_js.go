//go:build js

package main

import (
	"github.com/soypat/three"
	"gonum.org/v1/gonum/mat"
)

func addObjects(grp three.Group) {
	nodes, elems := feaModel()
	// Fiber compliance matrix
	Cf := orthotropicCompliance(235e3, 14e3, 0.2, 0.25, 28e3)
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
	dNxyz := mat.NewDense(3, 8, nil)
	for iele := range elems {
		storeElemNode(enod, nodes, elems[iele][:])
		for ipg := range upg {
			jac.Mul(dN[ipg], denseFromR3(enod))
			dNxyz.Solve(jac, dN[ipg])
		}
	}
}
