// Example to show fast show/hide of a layer, by "rolling" it up and down. This
// way, an animation could in fact look smoother than a regular update because
// on each new frame, only a small part of the screen needs to be updated.
package main

import (
	"image/color"
	"time"

	"github.com/aykevl/tilegraphics"
	"github.com/aykevl/tilegraphics/testscreen"
)

const printInfoEvery = time.Second * 3

func main() {
	println("start")
	screen := testscreen.NewScreen("reveal")
	screenWidth, screenHeight := screen.Size()
	engine := tilegraphics.NewEngine(screen)

	layer := engine.NewLayer(0, 0, screenWidth, 0, color.RGBA{255, 255, 255, 255})

	var layerHeight = int16(0)

	// Red rectangle in the middle of the screen (in the layer).
	layer.NewRectangle(screenWidth/2-10, screenHeight/2-10, 20, 20, color.RGBA{255, 0, 0, 255})

	lastPrint := time.Now()
	sumElapsed := time.Duration(0)
	numCycles := 0
	move := int16(1)
	for {
		start := time.Now()
		if layerHeight >= screenHeight {
			move = -4
		}
		if layerHeight == 0 {
			move = 4
		}
		layerHeight += move
		layer.Move(0, 0, screenWidth, layerHeight)
		engine.Display()

		// Sleep for a bit, trying to reach 60fps.
		elapsed := time.Since(start)
		sleepTime := time.Second/60 - elapsed
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}

		sumElapsed += elapsed
		numCycles++
		if lastPrint.Add(printInfoEvery).Before(start) {
			avgDuration := sumElapsed / time.Duration(numCycles)
			print("drawing: ", time.Second/avgDuration, "fps ", avgDuration.String(), "\r\n")
			sumElapsed = 0
			numCycles = 0
			lastPrint = start
		}
	}
}
