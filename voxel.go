package main

type VoxelGrid struct {
	voxel         [][]int
	size          [3]int
	origin        Vec
	resolution    float64
	resolutionInv float64
}

func New(resolution float64, size [3]int, origin Vec) *VoxelGrid {
	return &VoxelGrid{
		voxel:         make([][]int, size[0]*size[1]*size[2]),
		size:          size,
		origin:        origin,
		resolution:    resolution,
		resolutionInv: 1 / resolution,
	}
}

func (v *VoxelGrid) MinMax() (min, max Vec) {
	return v.origin, Add(v.origin, Vec{
		float64(v.size[0]) * v.resolution,
		float64(v.size[1]) * v.resolution,
		float64(v.size[2]) * v.resolution,
	})
}

func (v *VoxelGrid) Resolution() float64 {
	return v.resolution
}

func (v *VoxelGrid) Add(p Vec, index int) bool {
	addr, ok := v.Addr(p)
	if !ok {
		return false
	}
	ptr := &v.voxel[addr]
	*ptr = append(*ptr, index)
	return true
}

func (v *VoxelGrid) AddByAddr(a int, index int) {
	ptr := &v.voxel[a]
	*ptr = append(*ptr, index)
}

func (v *VoxelGrid) Get(p Vec) []int {
	addr, ok := v.Addr(p)
	if !ok {
		return nil
	}
	return v.voxel[addr]
}

func (v *VoxelGrid) GetByAddr(a int) []int {
	return v.voxel[a]
}

func (v *VoxelGrid) Addr(p Vec) (int, bool) {
	pos := Sub(p, v.origin)
	x := int(pos.X*v.resolutionInv + 0.5)
	if x < 0 || x >= v.size[0] {
		return 0, false
	}
	y := int(pos.Y*v.resolutionInv + 0.5)
	if y < 0 || y >= v.size[1] {
		return 0, false
	}
	z := int(pos.Z*v.resolutionInv + 0.5)
	if z < 0 || z >= v.size[2] {
		return 0, false
	}
	return x + (y+z*v.size[1])*v.size[0], true
}

func (v *VoxelGrid) AddrByPosInt(p [3]int) (int, bool) {
	x, y, z := p[0], p[1], p[2]
	if x < 0 || y < 0 || z < 0 || x >= v.size[0] || y >= v.size[1] || z >= v.size[2] {
		return 0, false
	}
	return x + (y+z*v.size[1])*v.size[0], true
}

func (v *VoxelGrid) PosInt(p Vec) ([3]int, bool) {
	pos := Sub(p, v.origin)
	x := int(pos.X*v.resolutionInv + 0.5)
	if x < 0 || x >= v.size[0] {
		return [3]int{}, false
	}
	y := int(pos.Y*v.resolutionInv + 0.5)
	if y < 0 || y >= v.size[1] {
		return [3]int{}, false
	}
	z := int(pos.Z*v.resolutionInv + 0.5)
	if z < 0 || z >= v.size[2] {
		return [3]int{}, false
	}
	return [3]int{x, y, z}, true
}

func (v *VoxelGrid) Len() int {
	return v.size[0] * v.size[1] * v.size[2]
}

func (v *VoxelGrid) Indice() []int {
	out := make([]int, 0, 1024)
	for _, g := range v.voxel {
		out = append(out, g...)
	}
	return out
}

func (v *VoxelGrid) Reset() {
	for i := range v.voxel {
		v.voxel[i] = nil
	}
}
