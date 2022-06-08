//go:build !js

package main

import (
	"fmt"
	"unsafe"

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

	modelSize := Vec{X: 10, Y: 10, Z: 10}
	// RUC surfaces
	var sx, sX, sy, sY, sz, sZ []int
	// RUC edges.
	var exy, eXy, exY, eXY, exz, eXz, exZ, eXZ, eyz, eYz, eyZ, eYZ []int
	// RUC corners.
	var cxyz, cXyz, cxYz, cXYz, cxyZ, cXyZ, cxYZ, cXYZ int
	for i, n := range nodes {
		// Sure this loop is ugly, but it should consume less energy
		// than having multiple loops since less compares. Save the trees?
		xeq0 := n.X == 0
		yeq0 := n.Y == 0
		zeq0 := n.Z == 0
		xeqL := n.X == modelSize.X
		yeqL := n.Y == modelSize.Y
		zeqL := n.Z == modelSize.Z
		if xeq0 {
			sx = append(sx, i)
			if yeq0 {
				exy = append(exy, i)
			} else if yeqL {
				exY = append(exY, i)
			}
			if zeq0 {
				exz = append(exz, i)
			} else if zeqL {
				exZ = append(exZ, i)
			}
		}
		if xeqL {
			sX = append(sX, i)
		}
		if yeq0 {
			sy = append(sy, i)
			if xeqL {
				eXy = append(eXy, i)
			}
			if zeq0 {
				eyz = append(eyz, i)
			} else if zeqL {
				eyZ = append(eyZ, i)
			}
		}
		if yeqL {
			sY = append(sY, i)
			if xeqL {
				eXY = append(eXY, i)
			}
			if zeq0 {
				eYz = append(eYz, i)
			} else if zeqL {
				eYZ = append(eYZ, i)
			}
		}
		if zeq0 {
			sz = append(sz, i)
			if xeqL {
				eXz = append(eXz, i)
			}
			if yeq0 && xeq0 {
				cxyz = i
			} else if yeq0 && xeqL {
				cXyz = i
			} else if yeqL && xeq0 {
				cxYz = i
			} else if yeqL && xeqL {
				cXYz = i
			}
		}
		if zeqL {
			sZ = append(sZ, i)
			if xeqL {
				eXZ = append(eXZ, i)
			}
			if yeq0 && xeq0 {
				cxyZ = i
			} else if yeq0 && xeqL {
				cXyZ = i
			} else if yeqL && xeq0 {
				cxYZ = i
			} else if yeqL && xeqL {
				cXYZ = i
			}
		}
	}
	surfSize := len(sx) + len(sX) + len(sy) + len(sY) + len(sz) + len(sZ)
	// lagrange := surfSize +
	// 	len(exy) + len(eXy) + len(exY) + len(eXY) + len(exz) + len(eXz) +
	// 	len(exZ) + len(eXZ) + len(eyz) + len(eYz) + len(eyZ) + len(eYZ) + 1
	rows := 0
	NN := mat.NewDense(surfSize+1, 3*len(nodes), nil)
	// X Surface displacement constraint.
	var nsx, nsy, nsz int
	for _, ix := range sx {
		p := nodes[ix]
		for _, iX := range sX {
			P := nodes[iX]
			if p.Z == P.Z && p.Y == P.Y && 0 < p.Z && p.Z < modelSize.Z &&
				0 < p.Y && p.Y < modelSize.Y {
				// Set displacement constraint on opposite nodes.
				constrainDisplacements(NN, rows, ix, iX)
				rows++
				nsx++
			}
		}
	}
	// Y Surface displacement constraint.
	for _, iy := range sy {
		p := nodes[iy]
		for _, iY := range sY {
			P := nodes[iY]
			if p.Z == P.Z && p.X == P.X && 0 < p.Z && p.Z < modelSize.Z &&
				0 < p.X && p.X < modelSize.X {
				// Set displacement constraint on opposite nodes.
				constrainDisplacements(NN, rows, iy, iY)
				rows++
				nsy++
			}
		}
	}
	// Z Surface displacement constraint.
	for _, iz := range sz {
		p := nodes[iz]
		for _, iZ := range sZ {
			P := nodes[iZ]
			if p.Y == P.Y && p.X == P.X && 0 < p.X && p.X < modelSize.X &&
				0 < p.Y && p.Y < modelSize.Y {
				// Set displacement constraint on opposite nodes.
				constrainDisplacements(NN, rows, iz, iZ)
				rows++
				nsz++
			}
		}
	}
	// Constrain edges.
	var dim = func(r rune) int { return int(r - 'X') } // returns 0,1,2 with arguments 'X', 'Y' and 'Z'
	ne1 := constrainRUCEdge(NN, nodes, exZ, eXz, rows, dim('Y'), modelSize)
	rows += ne1
	ne2 := constrainRUCEdge(NN, nodes, exz, eXZ, rows, dim('Y'), modelSize)
	rows += ne2
	ne3 := constrainRUCEdge(NN, nodes, exy, eXY, rows, dim('Z'), modelSize)
	rows += ne3
	ne4 := constrainRUCEdge(NN, nodes, exY, eXy, rows, dim('Z'), modelSize)
	rows += ne4
	ne5 := constrainRUCEdge(NN, nodes, eyZ, eYz, rows, dim('X'), modelSize)
	rows += ne5
	ne6 := constrainRUCEdge(NN, nodes, eyz, eYZ, rows, dim('X'), modelSize)
	rows += ne6
	// Constrain corners.
	constrainDisplacements(NN, rows, cXYz, cxyZ)
	rows++
	constrainDisplacements(NN, rows, cxyz, cXYZ)
	rows++
	constrainDisplacements(NN, rows, cxYZ, cXyz)
	rows++
	constrainDisplacements(NN, rows, cXyZ, cxYz)
	rows++
	r, _ := NN.Dims()
	if rows*3 != r {
		panic(r)
	}
	imposedLoads := make([]float64, (nsx+nsy+nsz)*3)

	for rucCase := 0; rucCase < 1; rucCase++ {
		rows = 0
		disp := imposedDisplacementForRUC(rucCase, 0.1)
		dx := disp.At(0, 0)
		dy := disp.At(1, 1)
		dz := disp.At(2, 2)
		dxy := disp.At(0, 1)
		dxz := disp.At(0, 2)
		dyz := disp.At(1, 2)
		var loads [3]float64
		// Surface load imposition (Lagrange).
		loads = [3]float64{dx * modelSize.X, dxy * modelSize.X, dxz * modelSize.X}
		for i := 0; i < nsx; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		loads = [3]float64{dxy * modelSize.Y, dy * modelSize.Y, dyz * modelSize.Y}
		for i := 0; i < nsy; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		loads = [3]float64{dxz * modelSize.Z, dyz * modelSize.Z, dz * modelSize.Z}
		for i := 0; i < nsz; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		// Edge load imposition (Lagrange).
		loads = [3]float64{modelSize.X*dx - modelSize.Z*dxz, modelSize.X*dxy - modelSize.Z*dyz, modelSize.X*dxz - modelSize.Z*dz}
		for i := 0; i < ne1; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		loads = [3]float64{modelSize.X*dx + modelSize.Z*dxz, modelSize.X*dxy + modelSize.Z*dyz, modelSize.X*dxz + modelSize.Z*dz}
		for i := 0; i < ne2; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		loads = [3]float64{modelSize.X*dx + modelSize.Y*dxy, modelSize.X*dxy + modelSize.Y*dy, modelSize.X*dxz + modelSize.Y*dyz}
		for i := 0; i < ne3; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		loads = [3]float64{modelSize.X*dx - modelSize.Y*dxy, modelSize.X*dxy - modelSize.Y*dy, modelSize.X*dxz - modelSize.Y*dyz}
		for i := 0; i < ne4; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		loads = [3]float64{modelSize.Y*dxy - modelSize.Z*dxz, modelSize.Y*dy - modelSize.Z*dyz, modelSize.Y*dyz - modelSize.Z*dz}
		for i := 0; i < ne5; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		loads = [3]float64{modelSize.Y*dxy + modelSize.Z*dxz, modelSize.Y*dy + modelSize.Z*dyz, modelSize.Y*dyz + modelSize.Z*dz}
		for i := 0; i < ne6; i++ {
			copy(imposedLoads[rows*3:], loads[:])
			rows++
		}
		// Corner load imposition (Lagrange).
		loads = [3]float64{modelSize.X*dx + modelSize.Y*dxy - modelSize.Z*dxz, modelSize.X*dxy + modelSize.Y*dy - modelSize.Z*dyz, modelSize.X*dxz + modelSize.Y*dyz - modelSize.Z*dz}
		rows += copy(imposedLoads[rows*3:], loads[:])
		loads = [3]float64{modelSize.X*dx + modelSize.Y*dxy + modelSize.Z*dxz, modelSize.X*dxy + modelSize.Y*dy + modelSize.Z*dyz, modelSize.X*dxz + modelSize.Y*dyz + modelSize.Z*dz}
		rows += copy(imposedLoads[rows*3:], loads[:])
		loads = [3]float64{-modelSize.X*dx + modelSize.Y*dxy + modelSize.Z*dxz, -modelSize.X*dxy + modelSize.Y*dy + modelSize.Z*dyz, -modelSize.X*dxz + modelSize.Y*dyz + modelSize.Z*dz}
		rows += copy(imposedLoads[rows*3:], loads[:])
		loads = [3]float64{modelSize.X*dx - modelSize.Y*dxy + modelSize.Z*dxz, modelSize.X*dxy - modelSize.Y*dy + modelSize.Z*dyz, modelSize.X*dxz - modelSize.Y*dyz + modelSize.Z*dz}
		rows += copy(imposedLoads[rows*3:], loads[:])
	}
	freeDofs := make([]bool, 3*len(nodes))
	// set last node to fixed
	freeDofs[len(freeDofs)-1] = true
	freeDofs[len(freeDofs)-2] = true
	freeDofs[len(freeDofs)-3] = true
	// KG := booleanIndexing(K, true, freeDofs, freeDofs)
	fmt.Println(NN)
	_ = sx
	_ = sX
	_, _, _, _, _, _, _, _ = cxyz, cXyz, cxYz, cXYz, cxyZ, cXyZ, cxYZ, cXYZ
}

func findNodes(nodes []Vec, f func(n Vec) bool) (idxs []int) {
	for i := range nodes {
		if f(nodes[i]) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func constrainRUCEdge(NN *mat.Dense, nodes []Vec, e1, e2 []int, rows, crossDim int, modelSize Vec) (proc int) {
	if crossDim < 0 || crossDim > 2 {
		panic("bad cross dimension (0,1,2 corresponds to X,Y,Z")
	}
	const (
		VecSize   = unsafe.Sizeof(Vec{})
		VecOffset = unsafe.Alignof(Vec{}.X)
	)
	rowsStart := rows
	nodePtr := uintptr(unsafe.Pointer(&nodes[0]))
	offset := uintptr(VecOffset * uintptr(crossDim))
	modelPtr := uintptr(unsafe.Pointer(&modelSize))
	modelDim := *(*float64)(unsafe.Pointer(modelPtr + offset))
	for _, i1 := range e1 {
		// very unsafe. very sharp.
		pdim := *(*float64)(unsafe.Pointer(nodePtr + VecSize*uintptr(i1) + offset))
		for _, i2 := range e2 {
			Pdim := *(*float64)(unsafe.Pointer(nodePtr + VecSize*uintptr(i2) + offset))
			if pdim == Pdim && 0 < pdim && pdim < modelDim {
				constrainDisplacements(NN, rows, i1, i2)
				rows++
			}
		}
	}
	return rows - rowsStart
}

func constrainDisplacements(NN *mat.Dense, r, i1, i2 int) {
	NN.Set(r*3, i1*3, -1)
	NN.Set(r*3, i2*3, 1)
	NN.Set(r*3+1, i1*3+1, -1)
	NN.Set(r*3+1, i2*3+1, 1)
	NN.Set(r*3+2, i1*3+2, -1)
	NN.Set(r*3+2, i2*3+2, 1)
}

// func convertToRenderTriangles(t []Triangle) []render.Triangle3 {
// 	return *(*[]render.Triangle3)(unsafe.Pointer(&t))
// 	// c := make([]render.Triangle3, len(t))
// 	// for i := range t {
// 	// 	c[i] = render.Triangle3(t[i])
// 	// }
// 	// return c
// }

/*
 7388.1         3806         3806            0            0            0
 3806       7388.1         3806            0            0            0
 3806         3806       7388.1            0            0            0
 0            0            0         1791            0            0
 0            0            0            0         1791            0
 0            0            0            0            0         1791
*/
