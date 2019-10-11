package tilegraphics

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
