// This is a small example (and test) how to use tilegraphics.
package main

import (
	"image/color"
	"time"

	"github.com/aykevl/tilegraphics"
)

const printInfoEvery = time.Second * 3

func main() {
	println("start")
	configureScreen()
	screenWidth, _ := screen.Size()
	engine := tilegraphics.NewEngine(screen)

	var (
		x      = int16(10)
		y      = int16(10)
		width  = int16(40)
		height = int16(40)
	)
	// Green layer.
	layer := engine.NewLayer(21, 10, screenWidth-38, 80, color.RGBA{0, 255, 0, 255})

	// Yellow rectangle.
	rect := layer.NewRectangle(x, y, width, height, color.RGBA{255, 255, 0, 255})

	lastPrint := time.Now()
	sumElapsed := time.Duration(0)
	numCycles := 0
	move := int16(1)
	for {
		start := time.Now()
		if x+width >= screenWidth-20 {
			// Note: this goes a bit outside of the layer.
			move = -1
		}
		if x <= -20 {
			move = 1
		}
		y += move
		x += move
		rect.Move(x, y, width, height)
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
