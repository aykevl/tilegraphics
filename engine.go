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

	// The objects slice is an array of objects that should be drawn (in order)
	// to the display, with backgroundColor as the background color.
	objects         []*Rectangle
	backgroundColor color.RGBA

	// cleanTiles stores for each tile whether it should be redrawn. True means
	// it is up-to-date, false means it should be redrawn.
	cleanTiles [][]bool

	// tile is a tile that is re-used for every drawing operation.
	tile *tile
}

// NewEngine creates a new rendering engine based on the displayer interface.
func NewEngine(display Displayer) *Engine {
	// Store which tiles are currently up-to-date and which aren't.
	width, height := display.Size()
	cleanTiles := make([][]bool, width/TileSize)
	for i := int16(0); i < width/TileSize; i++ {
		cleanTiles[i] = make([]bool, height/TileSize)
	}

	return &Engine{
		display:         display,
		tile:            &tile{},
		cleanTiles:      cleanTiles,
		backgroundColor: color.RGBA{0, 0, 0, 255},
	}
}

// SetBackgroundColor updates the background color of the display.
func (e *Engine) SetBackgroundColor(background color.RGBA) {
	e.backgroundColor = background
}

// NewRectangle adds a new rectangle to the display with the given color.
func (e *Engine) NewRectangle(x, y, width, height int16, c color.RGBA) *Rectangle {
	r := &Rectangle{
		engine: e,
		x:      x,
		y:      y,
		width:  width,
		height: height,
		color:  c,
	}
	e.objects = append(e.objects, r)
	r.invalidate()
	return r
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

			tileX := int16(row * TileSize)
			tileY := int16(col * TileSize)

			// Draw background.
			bg := Rectangle{e, tileX, tileY, TileSize, TileSize, e.backgroundColor}
			bg.Paint(e.tile, tileX, tileY)

			// Draw all objects in this tile.
			for _, obj := range e.objects {
				obj.Paint(e.tile, tileX, tileY)
			}

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
