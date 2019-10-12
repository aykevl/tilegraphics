package tilegraphics

import (
	"image/color"
	"math"
	"testing"

	"github.com/aykevl/tilegraphics/imagescreen"
)

func TestBlend(t *testing.T) {
	screen := imagescreen.NewScreen(256, 64*3)

	// Create 3 colored bands in the background: red, green, and blue.
	var (
		red   = color.RGBA{255, 0, 0, 255}
		green = color.RGBA{0, 255, 0, 255}
		blue  = color.RGBA{0, 0, 255, 255}
	)
	screen.FillRectangle(0, 0, 256, 64, red)
	screen.FillRectangle(0, 64, 256, 64, green)
	screen.FillRectangle(0, 128, 256, 64, blue)

	// Blend each band with the next color.
	// Note that each band is split in two: one with the correct gamma=2.2 for
	// comparison, and one with the custom (faster) 2.0 gamma.
	for x := 0; x <= 255; x++ {
		// Blend red and green.
		screen.FillRectangle(int16(x), 0, 1, 32, blendFloat(red, color.RGBA{0, uint8(x), 0, uint8(x)}))
		screen.FillRectangle(int16(x), 32, 1, 32, Blend(red, color.RGBA{0, uint8(x), 0, uint8(x)}))
		// Blend green and blue.
		screen.FillRectangle(int16(x), 64, 1, 32, blendFloat(green, color.RGBA{0, 0, uint8(x), uint8(x)}))
		screen.FillRectangle(int16(x), 96, 1, 32, Blend(green, color.RGBA{0, 0, uint8(x), uint8(x)}))
		// Blend blue and red.
		screen.FillRectangle(int16(x), 128, 1, 32, blendFloat(blue, color.RGBA{uint8(x), 0, 0, uint8(x)}))
		screen.FillRectangle(int16(x), 160, 1, 32, Blend(blue, color.RGBA{uint8(x), 0, 0, uint8(x)}))
	}

	matchImage(t, screen, "testdata/blend1.png")
}

// TestGamma checks whether all values decoded with decodeGamma are encoded to
// the same value with encodeGamma.
func TestGamma(t *testing.T) {
	for n := 0; n <= 255; n++ {
		linear := decodeGamma(uint8(n))
		n2 := encodeGamma(linear)
		if n2 != uint8(n) {
			t.Errorf("gamma conversion roundtrip failed for: %d -> %d -> %d", n, linear, n2)
		}
	}
}

// blendFloat takes in two colors and blends them together. The bottom color
// must have an opacity of 100% (A=255).
//
// This implementation is a reference to test against.
func blendFloat(bottom, top color.RGBA) color.RGBA {
	return color.RGBA{
		R: encodeGammaFloat(decodeGammaFloat(bottom.R)*(float64(255-top.A)/255) + decodeGammaFloat(top.R)),
		G: encodeGammaFloat(decodeGammaFloat(bottom.G)*(float64(255-top.A)/255) + decodeGammaFloat(top.G)),
		B: encodeGammaFloat(decodeGammaFloat(bottom.B)*(float64(255-top.A)/255) + decodeGammaFloat(top.B)),
		A: 255,
	}
}

func decodeGammaFloat(component uint8) float64 {
	return math.Pow(float64(component)/255, 2.2)
}

func encodeGammaFloat(component float64) uint8 {
	return uint8(math.Pow(component, 1/2.2) * 255)
}
