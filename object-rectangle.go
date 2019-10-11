package tilegraphics

import "image/color"

// Rectangle is a single rectangle drawn on the display that can be moved
// around.
type Rectangle struct {
	parent         *Layer // nil for the root
	x1, y1, x2, y2 int16
	color          color.RGBA
}

// boundingBox returns the exact bounding box of the rectangle.
func (r *Rectangle) boundingBox() (x1, y1, x2, y2 int16) {
	return r.x1, r.y1, r.x2, r.y2
}

// Move sets the new position and size of this rectangle.
func (r *Rectangle) Move(x, y, width, height int16) {
	newX1 := x
	newY1 := y
	newX2 := x + width
	newY2 := y + height

	if newX1 > r.x2 || newY1 > r.y2 || newX2 < r.x1 || newY2 < r.y1 {
		// Not overlapping. Simply invalidate the old and new rectangle.
		// https://stackoverflow.com/questions/306316/determine-if-two-rectangles-overlap-each-other
		r.invalidate(r.x1, r.y1, r.x2, r.y2)
		r.invalidate(newX1, newY1, newX2, newY2)

	} else {
		// Overlapping rectangles. Only redraw the parts that should be redrawn.
		// Background: https://magcius.github.io/xplain/article/regions.html
		// Essentially we need to invalidate the xor regions. There can be up
		// to 4 of them when two rectangles overlap.

		maxY1 := r.y1
		if newY1 > maxY1 {
			maxY1 = newY1
		}

		minY2 := r.y2
		if newY2 < minY2 {
			minY2 = newY2
		}

		if newX1 != r.x1 {
			// Invalidate the block on the left side of the rectangle.
			r.invalidateMiddleBlock(newX1, maxY1, r.x1, minY2)
		}
		if newX2 != r.x2 {
			// Invalidate the block on the right side of the rectangle.
			r.invalidateMiddleBlock(newX2, maxY1, r.x2, minY2)
		}
		if newY1 != r.y1 {
			// Invalidate the block on the top of the rectangle.
			if newY1 > r.y1 {
				// y1 moved down
				r.invalidate(r.x1, r.y1, r.x2, newY1)
			} else {
				// y1 moved up
				r.invalidate(newX1, newY1, newX2, r.y1)
			}
		}
		if newY2 != r.y2 {
			// Invalidate the block on the bottom of the rectangle.
			if newY2 > r.y2 {
				// y2 moved down
				r.invalidate(newX1, r.y2, newX2, newY2)
			} else {
				// y2 moved up
				r.invalidate(r.x1, newY2, r.x2, r.y2)
			}
		}
	}

	r.x1 = newX1
	r.y1 = newY1
	r.x2 = newX2
	r.y2 = newY2
}

// invalidateMiddleBlock invalidates an area where the two X coordinates might
// be swapped.
func (r *Rectangle) invalidateMiddleBlock(xA, maxY1, xB, minY2 int16) {
	if xA > xB {
		xA, xB = xB, xA
	}
	r.invalidate(xA, maxY1, xB, minY2)
}

// invalidate invalidates all tiles currently under the rectangle.
func (r *Rectangle) invalidate(x1, y1, x2, y2 int16) {
	x, y := r.absolutePos(x1, y1)
	// Calculate tile grid indices.
	tileX1 := x / TileSize
	tileY1 := y / TileSize
	tileX2 := (x + (x2 - x1) + TileSize) / TileSize
	tileY2 := (y + (y2 - y1) + TileSize) / TileSize

	// Limit the tile grid indices to the screen.
	if tileY1 < 0 {
		tileY1 = 0
	}
	if int(tileY2) >= len(r.parent.engine.cleanTiles) {
		tileY2 = int16(len(r.parent.engine.cleanTiles))
	}
	if tileX1 < 0 {
		tileX1 = 0
	}
	if int(tileX2) >= len(r.parent.engine.cleanTiles[0]) {
		tileX2 = int16(len(r.parent.engine.cleanTiles[0]))
	}

	// Set all tiles in bounds as needing an update.
	for tileY := tileY1; tileY < tileY2; tileY++ {
		tileRow := r.parent.engine.cleanTiles[tileY]
		for tileX := tileX1; tileX < tileX2; tileX++ {
			tileRow[tileX] = false
		}
	}
}

// paint draws the rectangle to the given tile at coordinates tileX and tileY.
func (r *Rectangle) paint(t *tile, tileX, tileY int16) {
	x1 := r.x1 - tileX
	y1 := r.y1 - tileY
	x2 := r.x2 - tileX
	y2 := r.y2 - tileY
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > TileSize {
		x2 = TileSize
	}
	if y2 > TileSize {
		y2 = TileSize
	}
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			t[x+y*TileSize] = r.color
		}
	}
}

// absolutePos returns the x and y coordinate of this rectangle in the screen.
func (r *Rectangle) absolutePos(x, y int16) (int16, int16) {
	layer := r.parent
	if &layer.rect == r {
		layer = layer.parent
	}
	for layer != nil {
		x += layer.rect.x1
		y += layer.rect.y1
		layer = layer.parent
	}
	return x, y
}
