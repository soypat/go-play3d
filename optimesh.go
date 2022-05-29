package main

type onode struct {
	// position of node
	c Vec
	// elements joined to node.
	tetras []otetra
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
				*on = onode{c: nodes[n], tetras: make([]otetra, 0, 4*6)}
			}
			on.tetras = append(on.tetras, otetra{tetidx: tetidx, hint: i})
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
