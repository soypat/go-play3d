//go:build js

package main

import (
	"bytes"
	_ "embed"
	"syscall/js"
	"time"

	"github.com/soypat/gwasm"
	"github.com/soypat/sdf3ui/uirender"
	"github.com/soypat/three"
)

//go:generate go run gen_shape.go

//go:embed shape.tri
var _gen_shape []byte

func genShape() []Triangle {
	triangles, err := uirender.DecodeAll(bytes.NewReader(_gen_shape))
	if err != nil {
		panic(err)
	}
	tri := make([]Triangle, len(triangles))
	for i := range tri {
		tri[i] = [3]Vec{
			Vec(triangles[i].V[0]),
			Vec(triangles[i].V[1]),
			Vec(triangles[i].V[2]),
		}
	}
	return tri
}

var (
	scene    three.Scene
	camera   three.PerspectiveCamera
	renderer three.WebGLRenderer
	controls three.TrackballControls
)

const (
	size = 10.0
)

func main() {

	// THREE Initialization.
	gwasm.AddScript("https://threejs.org/build/three.js", "THREE", time.Second)
	gwasm.AddScript("assets/trackball_controls.js", "TrackballControls", time.Second)
	err := three.Init()
	if err != nil {
		panic("three js init failed.")
	}

	document := js.Global().Get("document")
	windowWidth := js.Global().Get("innerWidth").Float()
	windowHeight := js.Global().Get("innerHeight").Float()
	devicePixelRatio := js.Global().Get("devicePixelRatio").Float()
	camera = three.NewPerspectiveCamera(70, windowWidth/windowHeight, size/100, size*100)
	camera.SetPosition(three.NewVector3(size/5, size/4, size/4))
	camera.LookAt(three.NewVector3(0, 0, 0))
	camera.SetUp(three.NewVector3(0, 1, 0))
	scene = three.NewScene()

	light := three.NewDirectionalLight(three.NewColor("white"), 1)
	light.SetPosition(three.NewVector3(size*5, size, 0))
	scene.Add(light)

	ambLight := three.NewAmbientLight(three.NewColorHex(0xbbbbbb), 0.4)
	scene.Add(ambLight)

	renderer = three.NewWebGLRenderer(three.WebGLRendererParam{})
	renderer.SetPixelRatio(devicePixelRatio)
	renderer.SetSize(windowWidth, windowHeight, true)
	rendererElement := renderer.Get("domElement")
	document.Get("body").Call("appendChild", rendererElement)

	scene.Add(makeObjects())

	// Controls to rotate camera around earth
	controls = three.NewTrackballControls(camera, rendererElement)
	controls.SetTarget(three.NewVector3(0, 0, 0))
	controls.SetMaxDistance(size * 10)
	controls.SetMinDistance(0.1)
	controls.SetZoomSpeed(1.2)
	controls.SetPanSpeed(.8)
	controls.SetRotateSpeed(.8)

	animate(js.Null(), nil)
	select {}
}

func animate(_ js.Value, _ []js.Value) interface{} {
	controls.Update()
	renderer.Render(scene, camera)

	// Best practice (soypat's opinion) to request frame after work is done.
	js.Global().Call("requestAnimationFrame", js.FuncOf(animate))
	return nil
}
