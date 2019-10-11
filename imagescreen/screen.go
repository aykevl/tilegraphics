// Package imagescreen implements a fake screen (actually, an image) that can be
// used for testing the tilegraphics package.
package imagescreen

import (
	"errors"
	"image"
	"image/color"
)

var (
	// ErrBufferSizeMismatch is returned when the size of the buffer passed to
	// FillRectangleWithBuffer doesn't match the to-be-updated area.
	ErrBufferSizeMismatch = errors.New("imagescreen: buffer size did not match width*height")
)

// Screen is a wrapper around an image, to implement the tilegraphics Displayer
// interface. It is used for testing.
type Screen struct {
	*image.RGBA
}

// NewScreen returns an in-memory memory buffer that acts as a screen,
// implementing the Displayer interface.
func NewScreen(width, height int16) *Screen {
	return &Screen{
		image.NewRGBA(image.Rect(0, 0, int(width), int(height))),
	}
}

// Size returns the width and height of this screen.
func (s *Screen) Size() (int16, int16) {
	rect := s.Bounds()
	return int16(rect.Max.X), int16(rect.Max.Y)
}

// Display implements the Displayer interface but is a no-op: data is stored
// directly to the image.
func (s *Screen) Display() error {
	return nil
}

// FillRectangle fills the given rectangle with the given color.
func (s *Screen) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	for pixelY := y; pixelY < y+height; pixelY++ {
		for pixelX := x; pixelX < x+width; pixelX++ {
			s.Set(int(pixelX), int(pixelY), c)
		}
	}
	return nil
}

// FillRectangleWithBuffer fills the given rectangle with a slice of colors. The
// buffer must be in row major order.
func (s *Screen) FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error {
	if len(buffer) != int(width*height) {
		return ErrBufferSizeMismatch
	}
	for pixelY := 0; pixelY < int(height); pixelY++ {
		for pixelX := 0; pixelX < int(width); pixelX++ {
			s.Set(int(x)+pixelX, int(y)+pixelY, buffer[pixelY*int(width)+pixelX])
		}
	}
	return nil
}
