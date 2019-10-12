package tilegraphics

import "image/color"

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

// Move sets the new position and size of this layer.
func (l *Layer) Move(x, y, width, height int16) {
	if x != l.rect.x1 || y != l.rect.y1 {
		// The layer was moved, so all containing objects must be redrawn.
		l.rect.invalidate(l.rect.x1, l.rect.y1, l.rect.x2, l.rect.y2)
	}

	// The layer wasn't moved, only its size changed. The objects that need to
	// be redrawn will be redrawn anyway with the standard algorithm.
	l.rect.Move(x, y, width, height)
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

// NewLine creates a new line with the given coordinates and line color. There
// is no restriction on the order of coordinates. Note that x2, y2 is inclusive,
// not exlusive, meaning that those pixels will get painted as well.
func (l *Layer) NewLine(x1, y1, x2, y2 int16, stroke color.RGBA) *Line {
	// Let the first coordinate always be to the left of the second coordinate.
	if x1 > x2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}
	line := &Line{
		parent: l,
		x1:     x1,
		y1:     y1,
		x2:     x2,
		y2:     y2,
		color:  stroke,
	}
	l.objects = append(l.objects, line)
	line.invalidate()
	return line
}

// paint draws the layer (and nothing outside the layer) to the tile at
// coordinates tileX and tileY.
func (l *Layer) paint(t *tile, tileX, tileY int16) {
	// Get a new tile to paint on from the tile pool, to avoid a heap
	// allocation.
	subtile := l.engine.getTile()

	// Paint the background, simply by filling this subtile with the layer
	// background color. Blending takes place when painting this tile on the
	// background, so don't blend here.
	for y := 0; y < TileSize; y++ {
		for x := 0; x < TileSize; x++ {
			subtile[y*TileSize+x] = l.rect.color
		}
	}

	// Draw all objects in this tile.
	l.paintObjects(subtile, tileX, tileY)

	// Move the tile coordinates into the layer coordinate system.
	tileX -= l.rect.x1
	tileY -= l.rect.y1

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

	// Paint the underlying tile using the temporary tile.
	if l.rect.color.A == 0xff {
		// Fast path: tile is fully opaque. We can draw directly in the passed
		// in tile.
		for x := x1; x < x2; x++ {
			for y := y1; y < y2; y++ {
				t[y*TileSize+x] = subtile[y*TileSize+x]
			}
		}
	} else {
		// Slow path. The background of this tile is at least partially
		// transparent, so blend the temporary tile with the passed in tile.
		for x := x1; x < x2; x++ {
			for y := y1; y < y2; y++ {
				t[y*TileSize+x] = Blend(t[y*TileSize+x], subtile[y*TileSize+x])
			}
		}
	}

	// Give the temporary tile back to the pool.
	l.engine.putTile(subtile)
}

// paintObjects will paint the objects in this layer into the given tile, at the
// given coordinates.
func (l *Layer) paintObjects(t *tile, tileX, tileY int16) {
	// Move the tile coordinates into the layer coordinate system.
	tileX -= l.rect.x1
	tileY -= l.rect.y1

	// Draw all objects in this tile.
	for _, obj := range l.objects {
		x1, y1, x2, y2 := obj.boundingBox()
		if x1 > tileX+TileSize || y1 > tileY+TileSize || x2 <= tileX || y2 <= tileY {
			// Object falls outside of this layer, so don't draw.
			continue
		}
		obj.paint(t, tileX, tileY)
	}
}
