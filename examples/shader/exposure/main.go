package main

//
// This example is my (thegtproject) own creation... I am not a designer- you've been warned.
// This is an attempt at performing a sort of "long-exposure" post effect.
// See ../assets/shaders/fastblur.frag.glsl for more details and comments
//

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/gopxl/pixel2"
	"github.com/gopxl/pixel2/imdraw"
	"github.com/gopxl/pixel2/pixelgl"
)

var (
	g      = 0.1
	r1     = 180.0
	r2     = 90.0
	m1     = 32.0
	m2     = 8.0
	a1v    = 0.0
	a2v    = 0.0
	a1, a2 = a1a2DefaultValues()
)

func run() {
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Bounds: pixel.R(0, 0, 600, 310),
		VSync:  true,
	})
	if err != nil {
		panic(err)
	}
	CenterWindow(win)
	win.SetSmooth(true)
	modelMatrix := pixel.IM.ScaledXY(pixel.ZV, pixel.V(1, -1)).Moved(pixel.V(300, 300))
	viewMatrix := pixel.IM.Moved(win.Bounds().Center())

	// I am putting all shader example initializing stuff here for
	// easier reference to those learning to use this functionality
	fragSource, err := LoadFileToString("../assets/shaders/exposure.frag.glsl")
	if err != nil {
		panic(err)
	}

	// Here we setup our uniforms. Think of uniforms as global variables
	// we can use inside of our fragment shader source code.
	var uTimeVar float32

	// We'll change this variable around with the arrow keys
	var uAmountVar float32 = 0.2

	// We will update these uniforms often, so use pointer
	EasyBindUniforms(win.Canvas(),
		"uTime", &uTimeVar,
		"uAmount", &uAmountVar,
	)

	// Since we are making a post effect, we want to apply the shader
	// to the entire final render. We will use an intermediate canvas
	// to complete the draw frame and then draw the canvas to the window's
	// canvas for shader processing. Otherwise, our shader would only be
	// running on active vertex positions, in this case, only the line
	// and circle draws generated by IMDraw.
	intermediatecanvas := pixelgl.NewCanvas(win.Bounds())
	intermediatecanvas.SetMatrix(modelMatrix)

	wc := win.Canvas()
	wc.SetFragmentShader(fragSource)

	sqrPos := win.Bounds().Moved(pixel.V(-300, -10))
	start := time.Now()
	for !win.Closed() {
		// Update our uniform variables
		uTimeVar = float32(time.Since(start).Seconds())

		switch {
		case win.Pressed(pixelgl.KeyLeft):
			uAmountVar -= 0.001
			win.SetTitle(fmt.Sprint(uAmountVar))
		case win.Pressed(pixelgl.KeyRight):
			uAmountVar += 0.001
			win.SetTitle(fmt.Sprint(uAmountVar))
		}

		win.SetClosed(win.JustPressed(pixelgl.KeyEscape) || win.JustPressed(pixelgl.KeyQ))
		if win.JustPressed(pixelgl.KeySpace) {
			a1, a2 = a1a2DefaultValues()
		}

		a, b := update()

		imd := imdraw.New(nil)

		// Clearing the background color with a filled rectangle so that
		// opengl will include the entire space in shader processing instead
		// of just where the shape objects are drawn.
		imd.Color = color.NRGBA{44, 44, 84, 255}
		imd.Push(sqrPos.Min, sqrPos.Max)
		imd.Rectangle(0)

		imd.Color = color.NRGBA{64, 64, 122, 255}
		imd.Push(pixel.ZV, a, b)
		imd.Line(3)

		imd.Color = color.NRGBA{51, 217, 178, 255}
		imd.Push(a)
		imd.Circle(m1/2, 0)

		imd.Color = color.NRGBA{255, 0, 0, 255}
		imd.Push(b)
		imd.Circle(m2/2, 0)

		imd.Draw(intermediatecanvas)
		intermediatecanvas.Draw(win, viewMatrix)
		win.Update()
	}
}

func update() (pixel.Vec, pixel.Vec) {
	a1a := a1aCalculation()
	a2a := a2aCalculation()

	a1v += a1a
	a2v += a2a

	a1 += a1v
	a2 += a2v

	a1v *= 0.9996
	a2v *= 0.9996

	a := pixel.V(r1*math.Sin(a1), r1*math.Cos(a1))
	b := pixel.V(a.X+r2*math.Sin(a2), a.Y+r2*math.Cos(a2))

	return a, b
}

func main() {
	pixelgl.Run(run)
}

func a1a2DefaultValues() (float64, float64) {
	return math.Pi / 2, math.Pi / 3
}

func a1aCalculation() float64 {
	num1 := -g * (2*m1 + m2) * math.Sin(a1)
	num2 := -m2 * g * math.Sin(a1-2*a2)
	num3 := -2 * math.Sin(a1-a2) * m2
	num4 := a2v*a2v*r2 + a1v*a1v*r1*math.Cos(a1-a2)
	den := r1 * (2*m1 + m2 - m2*math.Cos(2*a1-2*a2))

	return (num1 + num2 + num3*num4) / den
}

func a2aCalculation() float64 {
	num1 := 2 * math.Sin(a1-a2)
	num2 := (a1v * a1v * r1 * (m1 + m2))
	num3 := g * (m1 + m2) * math.Cos(a1)
	num4 := a2v * a2v * r2 * m2 * math.Cos(a1-a2)
	den := r2 * (2*m1 + m2 - m2*math.Cos(2*a2-2*a2))

	return (num1 * (num2 + num3 + num4)) / den
}
