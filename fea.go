package main

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"gonum.org/v1/gonum/mat"
)

var (
	//go:embed assets/crettonRUC_70_elements.tsv
	_rucElem string
	//go:embed assets/crettonRUC_70_nodes.tsv
	_rucNodes string
)

type hexa8 [8]Vec

func feaModel() (nodes []Vec, h8 [][8]int) {
	// preallocate reasonable size for warm start to appends.
	nodes = make([]Vec, 0, 512)
	h8 = make([][8]int, 0, 256)
	r := csv.NewReader(strings.NewReader(_rucNodes))
	r.Comma = '\t'
	r.ReuseRecord = true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		x, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			panic(err)
		}
		y, _ := strconv.ParseFloat(record[2], 64)
		z, _ := strconv.ParseFloat(record[3], 64)
		nodes = append(nodes, Vec{X: x, Y: y, Z: z})
	}
	r = csv.NewReader(strings.NewReader(_rucElem))
	r.Comma = '\t'
	r.ReuseRecord = true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		var elem [8]int
		for i := 0; i < 8; i++ {
			elem[i], err = strconv.Atoi(record[i+1])
			if err != nil {
				panic(err)
			}
			elem[i]-- // to account for 1 indexing.
		}
		h8 = append(h8, elem)
	}
	return nodes, h8
}

func isotropicCompliance(E, nu float64) *mat.Dense {
	return mat.NewDense(6, 6, []float64{
		((1 - nu) * E / ((1 + nu) * (1 - 2*nu))), nu * E / ((1 + nu) * (1 - 2*nu)), nu * E / ((1 + nu) * (1 - 2*nu)), 0, 0, 0,
		nu * E / ((1 + nu) * (1 - 2*nu)), ((1 - nu) * E / ((1 + nu) * (1 - 2*nu))), nu * E / ((1 + nu) * (1 - 2*nu)), 0, 0, 0,
		nu * E / ((1 + nu) * (1 - 2*nu)), nu * E / ((1 + nu) * (1 - 2*nu)), ((1 - nu) * E / ((1 + nu) * (1 - 2*nu))), 0, 0, 0,
		0, 0, 0, E / (2 * (1 + nu)), 0, 0,
		0, 0, 0, 0, E / (2 * (1 + nu)), 0,
		0, 0, 0, 0, 0, E / (2 * (1 + nu)),
	})
}

func orthotropicCompliance(E1, E2, nu12, nu23, G12 float64) *mat.Dense {
	m := mat.NewDense(6, 6, []float64{
		1 / E1, -nu12 / E1, -nu12 / E1, 0, 0, 0,
		-nu12 / E1, 1 / E2, -nu23 / E2, 0, 0, 0,
		-nu12 / E1, -nu23 / E2, 1 / E2, 0, 0, 0,
		0, 0, 0, 1 / G12, 0, 0,
		0, 0, 0, 0, 1 / (E2 / (2 * (nu23 + 1))), 0,
		0, 0, 0, 0, 0, 1 / G12,
	})
	m.Inverse(m)
	return m
}

func gauss3D(nx, ny, nz int) (pos []Vec, w []float64) {
	w = make([]float64, nx*ny*nz)
	pos = make([]Vec, nx*ny*nz)
	x, wx := gauss1D(nx)
	y, wy := gauss1D(ny)
	z, wz := gauss1D(nz)
	count := 0
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			for k := 0; k < nz; k++ {
				w[count] = wx[i] * wy[j] * wz[k]
				pos[count] = Vec{X: x[i], Y: y[j], Z: z[k]}
				count++
			}
		}
	}
	return pos, w
}

func gauss1D(n int) (x, w []float64) {
	switch n {
	case 1:
		x = []float64{0}
		w = []float64{2}
	case 2:
		a := math.Sqrt(3) / 3
		x = []float64{-a, a}
		w = []float64{1, 1}
	case 3:
		a := math.Sqrt(3.0 / 5.0)
		x = []float64{-a, 0, a}
		w = []float64{5 / 9, 8 / 9, 5 / 9}
	case 4:
		sqrt := 2 * math.Sqrt(6/5)
		sqrt30 := math.Sqrt(30)
		a := math.Sqrt((3 - sqrt) / 7)
		b := math.Sqrt((3 + sqrt) / 7)
		wa := (18 + sqrt30) / 36
		wb := (18 - sqrt30) / 36
		x = []float64{-b, -a, a, b}
		w = []float64{wb, wa, wa, wb}
	case 5:
		sqrt107 := 2 * math.Sqrt(10.0/7.0)
		sqrt70 := math.Sqrt(70)
		a := 1.0 / 3 * math.Sqrt(5-2*sqrt107)
		b := 1.0 / 3 * math.Sqrt(5+2*sqrt107)
		wa := (322 + 13*sqrt70) / 900
		wb := (322 - 13*sqrt70) / 900
		x = []float64{-b, -a, 0, a, b}
		w = []float64{wb, wa, 128 / 225, wa, wb}
	case 6:
		a := 0.932469514203152
		b := 0.661209386466265
		c := 0.238619186083197
		wa := 0.171324492379170
		wb := 0.360761573048139
		wc := 0.467913934572691
		x = []float64{-a, -b, -c, c, b, a}
		w = []float64{wa, wb, wc, wc, wb, wa}
	default:
		panic("bad argument to Gauss1D")
	}
	return x, w
}

func h8FormFuncs(ksi, eta, dseta float64) []float64 {
	return []float64{(dseta*eta)/8 - eta/8 - ksi/8 - dseta/8 + (dseta*ksi)/8 + (eta*ksi)/8 - (dseta*eta*ksi)/8 + 1.0/8.0, ksi/8 - eta/8 - dseta/8 + (dseta*eta)/8 - (dseta*ksi)/8 - (eta*ksi)/8 + (dseta*eta*ksi)/8 + 1.0/8.0, eta/8 - dseta/8 + ksi/8 - (dseta*eta)/8 - (dseta*ksi)/8 + (eta*ksi)/8 - (dseta*eta*ksi)/8 + 1.0/8.0, eta/8 - dseta/8 - ksi/8 - (dseta*eta)/8 + (dseta*ksi)/8 - (eta*ksi)/8 + (dseta*eta*ksi)/8 + 1.0/8.0, dseta/8 - eta/8 - ksi/8 - (dseta*eta)/8 - (dseta*ksi)/8 + (eta*ksi)/8 + (dseta*eta*ksi)/8 + 1.0/8.0, dseta/8 - eta/8 + ksi/8 - (dseta*eta)/8 + (dseta*ksi)/8 - (eta*ksi)/8 - (dseta*eta*ksi)/8 + 1.0/8.0, dseta/8 + eta/8 + ksi/8 + (dseta*eta)/8 + (dseta*ksi)/8 + (eta*ksi)/8 + (dseta*eta*ksi)/8 + 1.0/8.0, dseta/8 + eta/8 - ksi/8 + (dseta*eta)/8 - (dseta*ksi)/8 - (eta*ksi)/8 - (dseta*eta*ksi)/8 + 1.0/8.0}
}

func h8FormFuncsDiff(ksi, eta, dseta float64) []float64 {
	return []float64{dseta/8 + eta/8 - (dseta*eta)/8 - 1.0/8.0, (dseta*eta)/8 - eta/8 - dseta/8 + 1.0/8.0, eta/8 - dseta/8 - (dseta*eta)/8 + 1.0/8.0, dseta/8 - eta/8 + (dseta*eta)/8 - 1.0/8.0, eta/8 - dseta/8 + (dseta*eta)/8 - 1.0/8.0, dseta/8 - eta/8 - (dseta*eta)/8 + 1.0/8.0, dseta/8 + eta/8 + (dseta*eta)/8 + 1.0/8.0, -dseta/8 - eta/8 - (dseta*eta)/8 - 1.0/8.0,
		dseta/8 + ksi/8 - (dseta*ksi)/8 - 1.0/8.0, dseta/8 - ksi/8 + (dseta*ksi)/8 - 1.0/8.0, ksi/8 - dseta/8 - (dseta*ksi)/8 + 1.0/8.0, (dseta*ksi)/8 - ksi/8 - dseta/8 + 1.0/8.0, ksi/8 - dseta/8 + (dseta*ksi)/8 - 1.0/8.0, -dseta/8 - ksi/8 - (dseta*ksi)/8 - 1.0/8.0, dseta/8 + ksi/8 + (dseta*ksi)/8 + 1.0/8.0, dseta/8 - ksi/8 - (dseta*ksi)/8 + 1.0/8.0,
		eta/8 + ksi/8 - (eta*ksi)/8 - 1.0/8.0, eta/8 - ksi/8 + (eta*ksi)/8 - 1.0/8.0, -eta/8 - ksi/8 - (eta*ksi)/8 - 1.0/8.0, ksi/8 - eta/8 + (eta*ksi)/8 - 1.0/8.0, (eta*ksi)/8 - ksi/8 - eta/8 + 1.0/8.0, ksi/8 - eta/8 - (eta*ksi)/8 + 1.0/8.0, eta/8 + ksi/8 + (eta*ksi)/8 + 1.0/8.0, eta/8 - ksi/8 - (eta*ksi)/8 + 1.0/8.0}
}

func denseFromR3(v []Vec) *mat.Dense {
	data := make([]float64, 3*len(v))
	for i := range v {
		offset := i * 3
		data[offset] = v[i].X
		data[offset+1] = v[i].Y
		data[offset+2] = v[i].Z
	}
	return mat.NewDense(len(v), 3, data)
}

func storeElemNode(dst, allNodes []Vec, elem []int) {
	if len(dst) != len(elem) {
		panic("bad length")
	}
	for i := range elem {
		en := elem[i]
		dst[i] = allNodes[en]
	}
}

func storeElemDofs(dst []int, elem []int, ndofPerNode int) {
	if len(dst) != len(elem)*ndofPerNode {
		panic("bad length")
	}
	if ndofPerNode != 3 {
		panic("only works for 3 dimensions for now")
	}
	for i, node := range elem {
		id := i * ndofPerNode
		nodeDof := node * ndofPerNode
		dst[id] = nodeDof
		dst[id+1] = nodeDof + 1
		dst[id+2] = nodeDof + 2
	}
}

type byRealMagnitude []complex128

func (b byRealMagnitude) Less(i, j int) bool {
	return real(b[i]) < real(b[j])
}
func (b byRealMagnitude) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b byRealMagnitude) Len() int { return len(b) }

func imposedDisplacementForRUC(rucCase int, displacement float64) *mat.Dense {
	var dx, dy, dz, dxy, dxz, dyz float64
	switch rucCase {
	case 0:
		dx = displacement
	case 1:
		dy = displacement
	case 2:
		dz = displacement
	case 3:
		dxy = displacement / 2
	case 4:
		dxz = displacement / 2
	case 5:
		dyz = displacement / 2
	default:
		panic("invalid RUC case")
	}
	return mat.NewDense(3, 3, []float64{
		dx, dxy, dxz,
		dxy, dy, dyz,
		dxz, dyz, dz,
	})
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

func booleanSetVec(dst *mat.VecDense, src mat.Vector, inv bool, br []bool) {
	if len(br) != dst.Len() {
		panic("bad []bool len. must match dst")
	}
	sum := 0
	for i := range br {
		if br[i] != inv {
			sum++
		}
	}
	if sum != src.Len() {
		panic("amount of true values in br must match length of src")
	}
	sum = 0
	for i := range br {
		if br[i] != inv {
			dst.SetVec(i, src.AtVec(sum))
			sum++
		}
	}
}

func booleanIndexing(m mat.Matrix, inv bool, br, bc []bool) *subMat {
	r, c := m.Dims()
	if len(br) != r || len(bc) != c {
		panic("bad dim")
	}
	sm := subMat{
		ridx: make([]int, 0, r),
		cidx: make([]int, 0, c),
		m:    m,
	}
	for i, b := range br {
		if b != inv {
			sm.ridx = append(sm.ridx, i)
		}
	}
	for i, b := range bc {
		if b != inv {
			sm.cidx = append(sm.cidx, i)
		}
	}
	return &sm
}

type subMat struct {
	ridx, cidx []int
	m          mat.Matrix
}

func (bm *subMat) At(i, j int) float64 { return bm.m.At(bm.ridx[i], bm.cidx[j]) }
func (bm *subMat) AtVec(i int) float64 { return bm.m.At(bm.ridx[i], 0) }
func (bm *subMat) Len() int            { return len(bm.ridx) }
func (bm *subMat) Dims() (int, int)    { return len(bm.ridx), len(bm.cidx) }
func (bm *subMat) T() mat.Matrix {
	return mat.Transpose{Matrix: bm}
}

func copyBlocks(dst *mat.Dense, rows, cols int, src []mat.Matrix) error {
	if len(src) != rows*cols {
		return mat.ErrShape
	}
	var tr, tc int
	for i := 0; i < rows; i++ {
		r, _ := src[i*cols].Dims()
		tr += r
	}
	for j := 0; j < cols; j++ {
		_, c := src[j].Dims()
		tc += c
	}
	dst.ReuseAs(tr, tc)

	var br int
	for i := 0; i < rows; i++ {
		var bc int
		h, _ := src[i*cols].Dims()
		for j := 0; j < cols; j++ {
			r, c := src[i*cols+j].Dims()
			if r != h {
				return fmt.Errorf("matrix at %d,%d is wrong height: %d != %d:  %w", i, j, r, h, mat.ErrShape)
			}
			if i != 0 {
				_, w := src[j].Dims()
				if c != w {
					return fmt.Errorf("matrix at %d,%d is wrong width: %d != %d:  %w", i, j, c, w, mat.ErrShape)
				}
			}
			dst.Slice(br, br+r, bc, bc+c).(*mat.Dense).Copy(src[i*cols+j])
			bc += c
		}
		br += h
	}
	return nil
}

type eye int

func (m eye) At(i, j int) float64 {
	if i == j {
		return 1
	}
	return 0
}
func (m eye) Dims() (r, c int) { return int(m), int(m) }
func (m eye) T() mat.Matrix    { return m }

type zero struct{ r, c int }

func (m zero) At(i, j int) float64 { return 0 }
func (m zero) Dims() (r, c int)    { return m.r, m.c }
func (m zero) T() mat.Matrix       { return zero{m.c, m.r} }

func ExampleDense_copyBlocks() {
	var m mat.Dense
	r := mat.NewDense(5, 3, []float64{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
		10, 11, 12,
		13, 14, 15,
	})
	err := copyBlocks(&m, 2, 3, []mat.Matrix{
		eye(5), zero{5, 3}, r,
		zero{3, 5}, eye(3), zero{3, 3},
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mat.Formatted(&m))
}

func saveMatToFile(filename string, m mat.Matrix) {
	fp, _ := os.Create(filename)
	tstart := time.Now()
	defer fp.Close()
	fmt.Fprintf(fp, "%0.16g", mat.Formatted(m, mat.FormatMATLAB()))
	fmt.Println("writing "+filename+" took ", time.Since(tstart))
}
