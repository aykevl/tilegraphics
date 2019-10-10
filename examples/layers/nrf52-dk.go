// +build pca10040

package main

import (
	"machine"

	"tinygo.org/x/drivers/st7735"
)

var screen *st7735.Device

func init() {
	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       29,
		MOSI:      30,
		MISO:      machine.NoPin,
		Frequency: 8000000,
	})
	s := st7735.New(machine.SPI0, 17, 18, 31, 19)
	screen = &s
}

func configureScreen() {
	// Work around what looks like a bug in TinyGo.
	screen.Configure(st7735.Config{
		Width:        129,
		Height:       161,
		RowOffset:    -1,
		ColumnOffset: -1,
	})
}
