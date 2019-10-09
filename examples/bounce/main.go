// This is a small example (and test) how to use tilegraphics. It displays a
// small yellow square on a black background that bounces to both sides of the display.
package main

import (
	"image/color"
	"time"

	"github.com/aykevl/tilegraphics"
)

func main() {
	println("start")
	configureScreen()
	screenWidth, _ := screen.Size()
	engine := tilegraphics.NewEngine(screen)

	var (
		x      = int16(30)
		y      = int16(30)
		width  = int16(40)
		height = int16(40)
	)
	rect := engine.NewRectangle(x, y, width, height, color.RGBA{255, 255, 0, 255})
	engine.Display()

	lastPrint := time.Now()
	sumElapsed := time.Duration(0)
	numCycles := 0
	move := int16(1)
	for {
		start := time.Now()
		if x+width >= screenWidth {
			move = -1
		}
		if x <= 0 {
			move = 1
		}
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
		if lastPrint.Add(time.Second).Before(start) {
			avgDuration := sumElapsed / time.Duration(numCycles)
			print("drawing: ", time.Second/avgDuration, "fps ", avgDuration.String(), "\r\n")
			sumElapsed = 0
			numCycles = 0
			lastPrint = start
		}
	}
}
