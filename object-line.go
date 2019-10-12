package tilegraphics

import "image/color"

// Line is an anti-aliased line drawn between two coordinates (inclusive), with
// a given color. It supports transparency in the color.
type Line struct {
	parent         *Layer
	x1, y1, x2, y2 int16
	color          color.RGBA
}

// boundingBox returns the bounding box of this line.
func (l *Line) boundingBox() (x1, y1, x2, y2 int16) {
	// Note: a line takes up much less space than what the bounding box would
	// indicate.
	x1 = l.x1
	y1 = l.y1
	x2 = l.x2
	y2 = l.y2

	// The first coordinate is never to the right of the second coordinate. But
	// the y coordinates could still be swapped.
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return x1, y1, x2 + 1, y2 + 1
}

// invalidate marks the tiles that this line goes over as needing to be
// re-painted.
func (l *Line) invalidate() {
	// Crude hack to invalidate at least the area that this line is drawn in. It
	// is possible to make this far more efficient, by marking just the tiles
	// that need to be redrawn (using Wu's algorithm at the tile level).
	x1, y1, x2, y2 := l.boundingBox()
	r := Rectangle{
		parent: l.parent,
		x1:     x1,
		y1:     y1,
		x2:     x2,
		y2:     y2,
	}
	r.invalidate(x1, y1, x2, y2)
}

// paint draws the line to the given tile at coordinates tileX and tileY.
func (l *Line) paint(t *tile, tileX, tileY int16) {
	switch {
	case l.x1 == l.x2:
		// Easy: paint a vertical line.
		y1, y2 := l.y1, l.y2
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		x := l.x1 - tileX
		y1 -= tileY
		y2 -= tileY
		if x < 0 || x >= TileSize {
			return
		}
		if y1 < 0 {
			y1 = 0
		}
		if y2 >= TileSize {
			y2 = TileSize - 1
		}
		if l.color.A == 0xff {
			// Fast path, directly painting the color into the tile.
			for y := y1; y <= y2; y++ {
				t[y*TileSize+x] = l.color
			}
		} else {
			// Slow path, with color blending.
			for y := y1; y <= y2; y++ {
				t[y*TileSize+x] = Blend(t[y*TileSize+x], l.color)
			}
		}

	case l.y1 == l.y2:
		// Easy: paint a horizontal line.
		x1, x2 := l.x1, l.x2
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		y := l.y1 - tileY
		x1 -= tileX
		x2 -= tileX
		if y < 0 || y >= TileSize {
			return
		}
		if x1 < 0 {
			x1 = 0
		}
		if x2 >= TileSize {
			x2 = TileSize - 1
		}
		if l.color.A == 0xff {
			// Fast path, directly painting the color into the tile.
			for x := x1; x <= x2; x++ {
				t[y*TileSize+x] = l.color
			}
		} else {
			// Slow path, with color blending.
			for x := x1; x <= x2; x++ {
				t[y*TileSize+x] = Blend(t[y*TileSize+x], l.color)
			}
		}

	default:
		// Use Wu's antialiasing algorithm to paint the line. The algorithm is
		// used in a slightly different way than originally proposed, but the
		// basic idea is the same. The biggest difference is around the error
		// accumulator. Instead of counting up and incrementing the other axis
		// each time the accumulator rolls over, the other axis is calculated
		// from scratch each step with a multiply. This allows the algorithm to
		// jump into any position at the to-be-drawn line: perfect for a tile
		// based renderer.
		//
		// The paper:
		// https://scholar.google.nl/scholar?cluster=6182728158740398072
		// An easier to read description of the algorithm:
		// http://archive.gamedev.net/archive/reference/articles/article382.html

		// Starting point, as an offset from the tile.
		x1 := l.x1 - tileX
		x2 := l.x2 - tileX
		y1 := l.y1 - tileY
		y2 := l.y2 - tileY

		width := l.x2 - l.x1
		height := l.y2 - l.y1
		if height < 0 {
			height = -height
		}
		if width > height {
			// The line is more horizontal than vertical.
			// yIncrementQ16 is a 15.16 fixed-point number indicating how far
			// the y coordinate should be incremented each time x is
			// incremented.
			yIncrementQ16 := (int32(y2-y1) << 16) / int32(x2-x1)
			xStart := x1
			if x1 < 0 {
				x1 = 0
			}
			if x2 >= TileSize {
				x2 = TileSize - 1
			}
			for x := x1; x <= x2; x++ {
				// The y coordinate as a 15.16 fixed-point number.
				yQ16 := int32(x-xStart) * yIncrementQ16
				y := y1 + int16(yQ16>>16)
				l.paintPixel(t, x, y, 255-uint8(yQ16>>8))
				l.paintPixel(t, x, y+1, uint8(yQ16>>8))
			}
		} else {
			// The line is more vertical than horizontal.
			// Order the points, with point 2 below point 1.
			if y1 > y2 {
				y1, y2 = y2, y1
				x1, x2 = x2, x1
			}
			// xIncrementQ16 is a Q15.16 fixed-point number indicating how far
			// the x coordinate should be incremented each time y is
			// incremented.
			xIncrementQ16 := (int32(x2-x1) << 16) / int32(y2-y1)
			yStart := y1
			if y1 < 0 {
				y1 = 0
			}
			if y2 >= TileSize {
				y2 = TileSize - 1
			}
			for y := y1; y <= y2; y++ {
				xQ16 := int32(y-yStart) * xIncrementQ16
				x := x1 + int16(xQ16>>16)
				l.paintPixel(t, x, y, 255-uint8(xQ16>>8))
				l.paintPixel(t, x+1, y, uint8(xQ16>>8))
			}
		}
	}
}

func (l *Line) paintPixel(t *tile, x, y int16, weight uint8) {
	if x >= 0 && y >= 0 && x < TileSize && y < TileSize {
		t[y*TileSize+x] = Blend(t[y*TileSize+x], ApplyAlpha(l.color, weight))
	}
}
