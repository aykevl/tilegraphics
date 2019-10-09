# Tile based graphics rendering

This package implements tile-based graphics rendering, intended to be used on
small SPI-connected screens with a slow bus. It is designed to reduce the
number of pixels to redraw on every update, by only updating the parts of the
screen that did in fact change.

It is not yet complete. Currently the following objects can be drawn:

  * Rectangles with a solid color.

## License

This project has been licensed under the BSD 2-clause license, see the LICENSE
file for details.
