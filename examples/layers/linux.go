// +build linux,!baremetal

package main

import (
	"log"

	"github.com/aykevl/tilegraphics/sdlscreen"
)

var screen *sdlscreen.Screen

func init() {
	var err error
	screen, err = sdlscreen.NewScreen("layers", 129, 161)
	if err != nil {
		log.Fatalln("could not create screen:", err)
	}
}

func configureScreen() {
	// Nothing to do here.
}
