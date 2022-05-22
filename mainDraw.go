//go:build js

package main

import (
	"github.com/soypat/three"
)

func makeObjects() three.Object3D {
	grp := three.NewGroup()
	return grp
}
