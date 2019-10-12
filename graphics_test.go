package tilegraphics

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"testing"

	"github.com/aykevl/tilegraphics/imagescreen"
)

var flagUpdate = flag.Bool("update", false, "Update images based on test output.")

// Draw three rectangles on a screen, slightly overlapping the window border.
// Check whether the resulting image looks as expected.
func TestRectBasic(t *testing.T) {
	screen := imagescreen.NewScreen(100, 100)
	engine := NewEngine(screen)
	engine.NewRectangle(10, 10, 30, 30, color.RGBA{255, 255, 0, 255})
	engine.NewRectangle(-10, -10, 30, 30, color.RGBA{255, 150, 0, 255})
	engine.NewRectangle(90, 90, 30, 30, color.RGBA{0, 150, 0, 255})
	engine.Display()

	matchImage(t, screen, "testdata/rect1.png")
}

// Move a rectangle around and see whether the resulting image is the same as
// when it has been drawn on a new surface. This tests the rectangle
// invalidation logic.
func TestRectUpdate(t *testing.T) {
	// Get a deterministic randomness source.
	rand := rand.New(rand.NewSource(1))

	// These parameters will be changed at random.
	var (
		x            = int16(0)
		y            = int16(0)
		width        = int16(30)
		height       = int16(30)
		screenWidth  = int16(100)
		screenHeight = int16(100)
		rectColor    = color.RGBA{200, 200, 0, 200}
	)

	screen := imagescreen.NewScreen(screenWidth, screenHeight)
	engine := NewEngine(screen)
	engine.SetBackgroundColor(color.RGBA{50, 50, 50, 255})
	rect := engine.NewRectangle(x, y, width, height, rectColor)

	// Do fewer iterations if the -short parameter has been passed.
	max := 100
	if testing.Short() {
		max = 20
	}

	for i := 0; i < max; i++ {
		newX := int16(rand.Int31()%200 - 50)
		newY := int16(rand.Int31()%200 - 50)
		newWidth := int16(rand.Int31() % 50)
		newHeight := int16(rand.Int31() % 50)
		rect.Move(newX, newY, newWidth, newHeight)
		engine.Display()

		reference := imagescreen.NewScreen(screenWidth, screenHeight)
		referenceEngine := NewEngine(reference)
		referenceEngine.SetBackgroundColor(color.RGBA{50, 50, 50, 255})
		referenceEngine.NewRectangle(newX, newY, newWidth, newHeight, rectColor)
		referenceEngine.Display()
		if err := sameImage(screen, reference); err != nil {
			t.Errorf("moving rectangle with x=%d, y=%d, width=%d, height=%d resulted in a different image from creating it from scratch: %v", newX, newY, newWidth, newHeight, err)
			t.Errorf("previous rectangle: x=%d y=%d width=%d height=%d", x, y, width, height)
			saveTemporaryImages(t, "RectUpdate", i, screen, reference)
			x = newX
			y = newY
			width = newWidth
			height = newHeight
		}

		if i%10 == 9 {
			// Test with different screen sizes, to detect bugs on the screen boundary.
			screenWidth = int16(rand.Uint32()%20 + 100)
			screenHeight = int16(rand.Uint32()%20 + 100)
			if i%20 == 19 {
				rectColor = color.RGBA{255, 255, 0, 255}
			} else {
				rectColor = color.RGBA{200, 200, 0, 200}
			}
			screen = imagescreen.NewScreen(screenWidth, screenHeight)
			engine = NewEngine(screen)
			engine.SetBackgroundColor(color.RGBA{50, 50, 50, 255})
			rect = engine.NewRectangle(x, y, width, height, rectColor)
		}
	}
}

// Move a layer with an enclosed rectangle around in various ways, testing for
// inconsistent update behavior.
func TestLayerUpdate(t *testing.T) {
	// Get a deterministic randomness source.
	rand := rand.New(rand.NewSource(1))

	// Test different display sizes/configurations.
	const screenCycleMax = 20
	for screenCycle := 0; screenCycle < screenCycleMax; screenCycle++ {
		screenWidth := int16(rand.Uint32()%20 + 100)
		screenHeight := int16(rand.Uint32()%20 + 100)
		screen := imagescreen.NewScreen(screenWidth,
			screenHeight)
		engine := NewEngine(screen)
		engine.SetBackgroundColor(color.RGBA{50, 50, 50, 255})
		layerColor := color.RGBA{100, 0, 0, uint8(rand.Uint32()%1)*55 + 100}
		layer := engine.NewLayer(0, 0, 10, 10, layerColor)
		rectX := int16(rand.Int31()%200 - 50)
		rectY := int16(rand.Int31()%200 - 50)
		rectWidth := int16(rand.Int31() % 50)
		rectHeight := int16(rand.Int31() % 50)
		rectColor := color.RGBA{0, 100, 0, uint8(rand.Uint32()%1)*55 + 100}
		layer.NewRectangle(rectX, rectY, rectWidth, rectHeight, rectColor)

		// Move the layer around a few times, checking with a reference each
		// time that's built from scratch.
		const layerCycleMax = 8
		for layerCycle := 0; layerCycle < layerCycleMax; layerCycle++ {
			layerX := int16(rand.Int31()%200 - 50)
			layerY := int16(rand.Int31()%200 - 50)
			layerWidth := int16(rand.Int31() % 50)
			layerHeight := int16(rand.Int31() % 50)
			layer.Move(layerX, layerY, layerWidth, layerHeight)
			engine.Display()

			reference := imagescreen.NewScreen(screenWidth, screenHeight)
			referenceEngine := NewEngine(reference)
			referenceEngine.SetBackgroundColor(color.RGBA{50, 50, 50, 255})
			referenceLayer := referenceEngine.NewLayer(layerX, layerY, layerWidth, layerHeight, layerColor)
			referenceLayer.NewRectangle(rectX, rectY, rectWidth, rectHeight, rectColor)
			referenceEngine.Display()

			if err := sameImage(screen, reference); err != nil {
				t.Errorf("moving a layer didn't invalidate the correct area")
				saveTemporaryImages(t, "LayerUpdate", screenCycle*screenCycleMax+layerCycle, screen, reference)
			}
		}
	}
}

// Test drawing a few transparent rectangles partially over each other, and
// check whether the image output matches the expected output.
func TestRectTransparent(t *testing.T) {
	screen := imagescreen.NewScreen(100, 100)
	engine := NewEngine(screen)
	engine.NewRectangle(10, 10, 50, 50, color.RGBA{127, 0, 0, 127})
	engine.NewRectangle(25, 25, 50, 50, color.RGBA{0, 127, 0, 127})
	engine.NewRectangle(40, 40, 50, 50, color.RGBA{0, 0, 127, 127})
	engine.Display()

	matchImage(t, screen, "testdata/rect2.png")
}

// matchImage compares the given image with the PNG stored at the path, and will
// log an error if they don't match. Testing can continue on errors.
func matchImage(t *testing.T, screen *imagescreen.Screen, path string) {
	if *flagUpdate {
		err := saveImage(path, screen)
		if err != nil {
			t.Error("could not save image:", err)
		}
		return
	}

	reference, err := loadImage(path)
	if err != nil {
		t.Error("could not load image:", err)
		return
	}
	if err := sameImage(screen, reference); err != nil {
		t.Errorf("image %s didn't match: %s", path, err)
	}
}

// saveTemporaryImages tries to store two images to a temporary directory (for
// investigating test failures), and prints their paths if that is successful.
func saveTemporaryImages(t *testing.T, name string, num int, image1, image2 *imagescreen.Screen) {
	path1 := fmt.Sprintf("/tmp/graphics-%s-%d-moved.png", name, num)
	if saveImage(path1, image1) == nil {
		t.Error("\timage 1:", path1)
	}
	path2 := fmt.Sprintf("/tmp/graphics-%s-%d-reference.png", name, num)
	if saveImage(path2, image2) == nil {
		t.Error("\timage 2:", path2)
	}
}

// sameImage returns nil if both images are the same, or an error when they
// aren't.
func sameImage(screen *imagescreen.Screen, reference image.Image) error {
	width, height := screen.Size()
	referenceRect := reference.Bounds()
	if referenceRect.Min.X != 0 || referenceRect.Min.Y != 0 || referenceRect.Max.X != int(width) || referenceRect.Max.Y != int(height) {
		return fmt.Errorf("image is the wrong size: width=%d height=%d versus reference width=%d height=%d", width, height, referenceRect.Max.X, referenceRect.Max.Y)
	}

	for y := 0; y < referenceRect.Max.Y; y++ {
		for x := 0; x < referenceRect.Max.X; x++ {
			if reference.At(x, y) != screen.At(x, y) {
				return fmt.Errorf("pixel mismatch at X=%d Y=%d", x, y)
			}
		}
	}

	return nil // the same image
}

// loadImage is a helper function to load a single PNG image by filename.
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

// saveImage is a helper function to save a single PNG image by filename.
func saveImage(path string, image *imagescreen.Screen) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, image.RGBA)
}
