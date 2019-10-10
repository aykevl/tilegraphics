// Package tilegraphics implements a graphics rendering using small tiles. This
// is especially useful for devices with a slow bus to the display, like screens
// driven over SPI.
//
// Instantiate an Engine object using NewEngine and draw objects on it. The
// objects can be changed and moved after creation. These changes are recorded,
// but not directly sent to the screen. They're only sent on the next call to
// Display(). It is therefore recommended to only call Display() when the
// display should really be updated, to send all the updates in a single batch
// for improved performance.
package tilegraphics

import "image/color"

// TileSize is the size (width and height) of a tile. A tile will take up
// TileSize*TileSize*4 bytes of memory during rendering.
const TileSize = 8

// engineDebug can be set to true for extra logging.
const engineDebug = false

// Displayer is the display interface required by the rendering engine.
type Displayer interface {
	// Size returns the display size in pixels. It must never change.
	Size() (int16, int16)

	// Display sends the last updates to the screen, if needed.
	// The display might be updated directly or only after this method has been
	// called, depending on the implementation.
	Display() error

	// FillRectangle fills the given rectangle with the given color, and returns
	// an error if something went wrong.
	FillRectangle(x, y, width, height int16, c color.RGBA) error

	// FillRectangleWithBuffer fills the given rectangle with a slice of colors.
	// The buffer is stored in row major order.
	FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error
}

// tile encapsulates a single tile with colors in row major order.
type tile [TileSize * TileSize]color.RGBA

// Engine is the actual rendering engine. Use NewEngine to construct a new rendering engine.
type Engine struct {
	// display is the backing display to which all pixels will be drawn once
	// the Display method is called.
	display Displayer

	// The root layer, that stores the background color and the list of objects
	// (in order) that should be drawn on each tile.
	root Layer

	// cleanTiles stores for each tile whether it should be redrawn. True means
	// it is up-to-date, false means it should be redrawn.
	cleanTiles [][]bool

	// tile is a tile that is re-used for all root tiles.
	tile *tile

	// tilePool is a slice of re-usable tiles. They can be used for layer
	// drawing, without allocating a new tile every time or allocating a big
	// object on the stack (if it gets stack-allocated at all).
	tilePool []*tile
}

// NewEngine creates a new rendering engine based on the displayer interface.
func NewEngine(display Displayer) *Engine {
	// Store which tiles are currently up-to-date and which aren't.
	width, height := display.Size()
	cleanTiles := make([][]bool, width/TileSize)
	for i := int16(0); i < width/TileSize; i++ {
		cleanTiles[i] = make([]bool, height/TileSize)
	}

	e := &Engine{
		display:    display,
		tile:       &tile{},
		cleanTiles: cleanTiles,
	}
	e.root = Layer{
		rect: Rectangle{
			x1:    0,
			y1:    0,
			x2:    width,
			y2:    height,
			color: color.RGBA{0, 0, 0, 255}, // black background by default
		},
		engine: e,
	}
	e.root.rect.parent = &e.root
	return e
}

// SetBackgroundColor updates the background color of the display. Note that the
// alpha channel should be 100% (255) and will be ignored.
func (e *Engine) SetBackgroundColor(background color.RGBA) {
	e.root.SetBackgroundColor(background)
}

// NewRectangle adds a new rectangle to the display with the given color.
func (e *Engine) NewRectangle(x, y, width, height int16, c color.RGBA) *Rectangle {
	return e.root.NewRectangle(x, y, width, height, c)
}

// NewLayer creates a new layer to the display with the given background color.
func (e *Engine) NewLayer(x, y, width, height int16, background color.RGBA) *Layer {
	return e.root.NewLayer(x, y, width, height, background)
}

// getTile returns a reusable tile from the tile pool, without allocating a new
// tile. It should be returned to the tile pool after use with putTile.
func (e *Engine) getTile() *tile {
	if len(e.tilePool) != 0 {
		// A reusable tile was found.
		t := e.tilePool[len(e.tilePool)-1]
		e.tilePool = e.tilePool[:len(e.tilePool)-1]
		return t
	}
	// No reusable tile was found, make a new one.
	return &tile{}
}

// putTile returns a tile back to the tile pool that isn't used anymore.
func (e *Engine) putTile(t *tile) {
	e.tilePool = append(e.tilePool, t)
}

// Display updates the display with all the changes that have been done since
// the last update.
func (e *Engine) Display() {
	tilesDrawn := 0
	for row, cleanTilesRow := range e.cleanTiles {
		for col, cleanTile := range cleanTilesRow {
			if cleanTile {
				// Already updated.
				continue
			}
			// Will be true after this loop body finishes.
			cleanTilesRow[col] = true
			tilesDrawn++

			// Paint tile.
			tileX := int16(row * TileSize)
			tileY := int16(col * TileSize)
			e.root.paint(e.tile, tileX, tileY)

			// Draw tile in screen.
			e.display.FillRectangleWithBuffer(tileX, tileY, TileSize, TileSize, e.tile[:])
		}
	}

	if engineDebug {
		println("tiles drawn:", tilesDrawn)
	}

	// Send the update to the screen. Not all Displayer implementations need this.
	e.display.Display()
}
