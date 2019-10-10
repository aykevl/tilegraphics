package tilegraphics

import "image/color"

// object is something that can be drawn on the screen.
type object interface {
	// Paint draws this object on the given tile. The tile coordinates are the
	// offsets of the tile from the coordinates of the objects relative to the
	// parent.
	paint(t *tile, tileX, tileY int16)

	// boundingBox returns the bounding box of this object. It should be as
	// small as possible for maximum performance, but drawing outside the
	// bounding box will lead to unexpected results.
	// The x2 and y2 values are the coordinates that lie just outside of the
	// bounding box, so (2, 2, 3, 4) will cover just two pixels.
	boundingBox() (x1, y1, x2, y2 int16)
}

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
	// Calculate tile coordinates.
	tileX1 := x / TileSize
	tileY1 := y / TileSize
	tileX2 := (x + (x2 - x1) + TileSize) / TileSize
	tileY2 := (y + (y2 - y1) + TileSize) / TileSize

	for tileX := tileX1; tileX < tileX2; tileX++ {
		if tileX < 0 || int(tileX) >= len(r.parent.engine.cleanTiles) {
			continue
		}
		tileRow := r.parent.engine.cleanTiles[tileX]
		for tileY := tileY1; tileY < tileY2; tileY++ {
			if tileY < 0 || int(tileY) >= len(tileRow) {
				continue
			}
			tileRow[tileY] = false
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

// Layer contains other objects, but makes sure no containing objects will draw
// outside of its boundaries. This object provides some encapsulation and
// improves performance when rendering many objects on a screen.
type Layer struct {
	rect    Rectangle
	engine  *Engine
	parent  *Layer // may be nil for the root
	objects []object
}

// boundingBox returns the exact bounding box of this layer.
func (l *Layer) boundingBox() (x1, y1, x2, y2 int16) {
	return l.rect.boundingBox()
}

// SetBackgroundColor updates the background color of this layer.
func (l *Layer) SetBackgroundColor(background color.RGBA) {
	l.rect.color = background
	l.rect.invalidate(l.rect.x1, l.rect.y1, l.rect.x2, l.rect.y2)
}

// NewRectangle adds a new rectangle to the layer with the given color.
func (l *Layer) NewRectangle(x, y, width, height int16, c color.RGBA) *Rectangle {
	r := &Rectangle{
		parent: l,
		x1:     x,
		y1:     y,
		x2:     x + width,
		y2:     y + height,
		color:  c,
	}
	l.objects = append(l.objects, r)
	r.invalidate(r.x1, r.y1, r.x2, r.y2)
	return r
}

// NewLayer returns a new layer inside this layer, with the given coordinates
// (relative to the parent layer) and the given background color.
func (l *Layer) NewLayer(x, y, width, height int16, background color.RGBA) *Layer {
	child := &Layer{
		rect: Rectangle{
			x1:    x,
			y1:    y,
			x2:    x + width,
			y2:    y + height,
			color: background,
		},
		engine: l.engine,
		parent: l,
	}
	child.rect.parent = child
	child.rect.invalidate(child.rect.x1, child.rect.y1, child.rect.x2, child.rect.y2)
	l.objects = append(l.objects, child)
	return child
}

// paint draws the layer (and nothing outside the layer) to the tile at
// coordinates tileX and tileY.
func (l *Layer) paint(t *tile, tileX, tileY int16) {
	// Get a new tile to paint on from the tile pool, to avoid a heap
	// allocation.
	subtile := l.engine.getTile()

	// Paint the background. The background works from the parent coordinates
	// (because it defines the rect coordinates), so don't adjust tileX and
	// tileY.
	l.rect.paint(subtile, tileX, tileY)

	// Move the tile coordinates into the layer coordinate system.
	tileX -= l.rect.x1
	tileY -= l.rect.y1

	// Draw all objects in this tile.
	for _, obj := range l.objects {
		x1, y1, x2, y2 := obj.boundingBox()
		if x1 > tileX+TileSize || y1 > tileY+TileSize || x2 < tileX || y2 < tileY {
			// Object falls outside of this layer, so don't draw.
			continue
		}
		obj.paint(subtile, tileX, tileY)
	}

	// Determine the bounds of the tile that should be painted to.
	var (
		x1 = int16(0)
		x2 = int16(TileSize)
		y1 = int16(0)
		y2 = int16(TileSize)
	)
	if tileX < 0 {
		x1 = -tileX
	}
	if tileY < 0 {
		y1 = -tileY
	}
	if tileX+TileSize > l.rect.x2-l.rect.x1 {
		x2 = (l.rect.x2 - l.rect.x1) - tileX
	}
	if tileY+TileSize > l.rect.y2-l.rect.y1 {
		y2 = (l.rect.y2 - l.rect.y1) - tileY
	}

	// Paint the parts of the layer tile that are part of the layer to the underlying tile.
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			t[y*TileSize+x] = subtile[y*TileSize+x]
		}
	}

	// Give the temporary tile back to the pool.
	l.engine.putTile(subtile)
}
