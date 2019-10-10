// +build pca10040

package testscreen

import (
	"machine"

	"tinygo.org/x/drivers/st7735"
)

func NewScreen(name string) *st7735.Device {
	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       29,
		MOSI:      30,
		MISO:      machine.NoPin,
		Frequency: 8000000,
	})
	screen := st7735.New(machine.SPI0, 17, 18, 31, 19)
	screen.Configure(st7735.Config{
		Width:        129,
		Height:       161,
		RowOffset:    -1,
		ColumnOffset: -1,
	})
	return &screen
}
