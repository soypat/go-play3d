package main

type onode struct {
	// position of node
	c Vec
	// elements joined to node.
	tetras       []otetra
	connectivity []int
}

type otetra struct {
	tetidx int
	hint   int
}

type omesh struct {
	nodes  []onode
	tetras [][4]int
}

func newOptimesh(nodes []Vec, tetras [][4]int) *omesh {
	onodes := make([]onode, len(nodes))
	for tetidx, tetra := range tetras {
		for i := range tetra {
			n := tetra[i]
			on := &onodes[n]
			if on.tetras == nil {
				*on = onode{c: nodes[n], tetras: make([]otetra, 0, 4*6), connectivity: make([]int, 0, 16)}
			}
			on.tetras = append(on.tetras, otetra{tetidx: tetidx, hint: i})
			// Add tetrahedron's incident nodes to onode connectivity if not present.
			for j := 0; j < 3; j++ {
				var existing int
				c := tetra[(i+j+1)%3]
				// Lot of work goes into making sure connectivity is unique list.
				for _, existing = range on.connectivity {
					if c == existing {
						break
					}
				}
				if c != existing {
					on.connectivity = append(on.connectivity, c)
				}
			}
		}
	}
	return &omesh{
		nodes:  onodes,
		tetras: tetras,
	}
}

func (om *omesh) foreach(f func(i int, on *onode)) {
	for i := range om.nodes {
		if len(om.nodes[i].tetras) == 0 {
			continue
		}
		f(i, &om.nodes[i])
	}
}

func (om *omesh) nodePositions() []Vec {
	nn := make([]Vec, len(om.nodes))
	for i := range om.nodes {
		nn[i] = Vec(om.nodes[i].c)
	}
	return nn
}
