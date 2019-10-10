// +build linux,!baremetal

package testscreen

import (
	"log"

	"github.com/aykevl/tilegraphics/sdlscreen"
)

func NewScreen(name string) *sdlscreen.Screen {
	screen, err := sdlscreen.NewScreen(name, 129, 161)
	if err != nil {
		log.Fatalln("could not create screen:", err)
	}
	return screen
}
