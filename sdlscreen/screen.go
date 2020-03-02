// Package sdlscreen implements a Displayer interface as required by
// tilegraphics and outputs using SDL2.
//
// This package does not use SDL2 entirely correctly, nor is it really intended
// for serious use. It is mainly useful for testing the tilegraphics package
// without constantly reflashing a microcontroller.
package sdlscreen

import (
	"image/color"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// Screen is a window implemented using SDL2 that can be drawn to.
type Screen struct {
	window  *sdl.Window
	surface *sdl.Surface
}

// NewScreen creates a new window with the given width and height.
func NewScreen(name string, width, height int16) (*Screen, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow(name, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(width), int32(height), sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	surface, err := window.GetSurface()
	if err != nil {
		return nil, err
	}
	s := &Screen{
		window:  window,
		surface: surface,
	}

	// Note: it is technically not allowed to do this in a separate goroutine,
	// but it seems to work in practice on Linux for receiving events.
	// See: https://wiki.libsdl.org/CategoryThread
	go s.background()

	return s, nil
}

// background runs in a goroutine and polls for events, like the Quit button or
// redraw events.
func (s *Screen) background() {
	for {
		event := sdl.PollEvent()
		if event == nil {
			time.Sleep(30 * time.Millisecond)
			continue
		}
		switch event.(type) {
		case *sdl.QuitEvent:
			os.Exit(0)
		case *sdl.WindowEvent:
			s.window.UpdateSurface()
		}
	}
}

// Size returns the window content size.
func (s *Screen) Size() (int16, int16) {
	return int16(s.surface.W), int16(s.surface.H)
}

// Display updates the window surface. This is necessary to actually write to
// the screen what was drawn using FillRectangle, for example.
func (s *Screen) Display() error {
	return s.window.UpdateSurface()
}

// FillScreen sets the whole screen to the given color.
func (s *Screen) FillScreen(c color.RGBA) error {
	width, height := s.Size()
	s.FillRectangle(0, 0, width, height, c)
	return nil
}

// FillRectangle fills the given rectangle with the given color, and returns an
// error if something went wrong.
func (s *Screen) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	rect := sdl.Rect{
		X: int32(x),
		Y: int32(y),
		W: int32(width),
		H: int32(height),
	}
	return s.surface.FillRect(&rect, colorToInt(c))
}

// FillRectangleWithBuffer fills the given rectangle with a slice of colors. The
// buffer must be in row major order.
func (s *Screen) FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error {
	if s.surface.MustLock() {
		s.surface.Lock()
		defer s.surface.Unlock()
	}
	for bufferX := int16(0); bufferX < width; bufferX++ {
		for bufferY := int16(0); bufferY < height; bufferY++ {
			s.SetPixel(bufferX+x, bufferY+y, buffer[bufferX+bufferY*width])
		}
	}
	return nil
}

// SetPixel sets the pixel at the given coordinates to the given color. Setting
// a pixel out of bounds is allowed: it won't do anything. An error may be
// returned if setting the pixel failed.
func (s *Screen) SetPixel(x, y int16, c color.RGBA) {
	surfaceX := int(x)
	surfaceY := int(y)
	if surfaceX >= 0 && surfaceY >= 0 && surfaceX < int(s.surface.W) || surfaceY < int(s.surface.H) {
		s.surface.Set(surfaceX, surfaceY, c)
	}
}

// Close closes the window.
func (s *Screen) Close() error {
	return s.window.Destroy()
}

// colorToInt is a small helper function to transform a color.RGBA to a color
// integer as expected by SDL2.
func colorToInt(c color.RGBA) uint32 {
	return uint32(c.A)<<24 | uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B)
}
