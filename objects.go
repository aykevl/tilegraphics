package tilegraphics

import "image/color"

// Rectangle is a single rectangle drawn on the display that can be moved
// around.
type Rectangle struct {
	engine              *Engine
	x, y, width, height int16
	color               color.RGBA
}

// Move sets the new position and size of this rectangle.
func (r *Rectangle) Move(x, y, width, height int16) {
	r.invalidate()
	r.x = x
	r.y = y
	r.width = width
	r.height = height
	r.invalidate()
}

// invalidate invalidates all tiles currently under the rectangle.
func (r *Rectangle) invalidate() {
	// Calculate tile coordinates.
	tileX1 := r.x / TileSize
	tileY1 := r.y / TileSize
	tileX2 := (r.x + r.width + TileSize) / TileSize
	tileY2 := (r.y + r.height + TileSize) / TileSize

	for tileX := tileX1; tileX < tileX2; tileX++ {
		if tileX < 0 || int(tileX) >= len(r.engine.cleanTiles) {
			continue
		}
		tileRow := r.engine.cleanTiles[tileX]
		for tileY := tileY1; tileY < tileY2; tileY++ {
			if tileY < 0 || int(tileY) >= len(tileRow) {
				continue
			}
			tileRow[tileY] = false
		}
	}
}

// Paint draws the rectangle to the given tile at coordinates tileX and tileY.
func (r *Rectangle) Paint(t *tile, tileX, tileY int16) {
	x1 := r.x - tileX
	y1 := r.y - tileY
	x2 := r.x - tileX + r.width
	y2 := r.y - tileY + r.height
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
