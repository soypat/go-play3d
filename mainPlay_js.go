//go:build js

package main

import (
	"github.com/soypat/sdf/form3/must3"
	"github.com/soypat/three"
	"gonum.org/v1/gonum/spatial/r3"
)

func addObjects(grp three.Group) {
	grp.Add(three.NewAxesHelper(1))
	const quality = 20
	s := must3.Cylinder(2, .5, .1)

	grp.Add(sdf3Obj(s, quality, "red", .5))
	sbb := s.Bounds()
	bb := Box{Vec(sbb.Min), Vec(sbb.Max)}

	// grp.Add(boxesObj(bb.Octree(), lineColor("green")))
	mainBox := CenteredBox(bb.Center(), bb.Scale(Vec{1.1, 1.1, 1.1}).Size())
	boxes := boxDivide(mainBox, 10)
	norms := make([][2]Vec, len(boxes))
	for i, box := range boxes {
		const tol = 1e-3
		c := box.Center()
		curvature := sdfCurvature(s, c, tol)
		norms[i] = [2]Vec{c, Add(c, Scale(curvature*1e5, sdfNormal(s, c, tol)))}
	}
	mesh := maketmesh(mainBox, .1)
	// grp.Add(boxesObj(boxes, lineColor("blue")))
	// grp.Add(linesObj(norms, lineColor("gold")))
	// grp.Add(boxesObj(mesh.boxes(), lineColor("green")))
	nodes, tetras := mesh.meshTetraBCC()
	eval := func(v Vec) float64 { return s.Evaluate(r3.Vec(v)) }
	newtetras := make([][4]int, 0, len(tetras))

	for _, tetra := range tetras {
		nd := Tetra{nodes[tetra[0]], nodes[tetra[1]], nodes[tetra[2]], nodes[tetra[3]]}
		// aspect := nd.aspect()
		// if aspect > 1 {
		// 	// fmt.Println(nd.altitudes())
		// 	fmt.Println(aspect)
		// 	newtetras = append(newtetras, tetra)
		// }
		if eval(nd[0]) < 0 || eval(nd[1]) < 0 || eval(nd[2]) < 0 || eval(nd[3]) < 0 {
			newtetras = append(newtetras, tetra)
		}

	}
	// grp.Add(pointsObj(nodes, pointColor("cyan")))
	grp.Add(triangleMesh(tetraTriangles(nodes, newtetras), phongMaterial("orange", 0.5)))
}

func tetraTriangles(nodes []Vec, tetras [][4]int) []Triangle {
	var triangles []Triangle
	for _, tetra := range tetras {
		nd := [4]Vec{nodes[tetra[0]], nodes[tetra[1]], nodes[tetra[2]], nodes[tetra[3]]}
		triangles = append(triangles,
			Triangle{nd[0], nd[1], nd[2]},
			Triangle{nd[1], nd[3], nd[2]},
			Triangle{nd[0], nd[3], nd[1]},
			Triangle{nd[0], nd[2], nd[3]},
		)
	}
	return triangles
}
