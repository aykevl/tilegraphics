package tilegraphics

import (
	"image/color"
)

// Blend takes a fully opaque background color and a foreground color that may
// be semi-transparent and blends them together.
//
// Color blending uses a gamma of 2.0, which is close to the commonly used gamma
// of ~2.2 but is much easier to calculate efficiently. It is slightly off the
// ideal gamma curve, but in practice it looks almost identical.
//
// For more information on why blending isn't trivial in sRGB color space:
// https://www.youtube.com/watch?v=LKnqECcg6Gw
// https://blog.johnnovak.net/2016/09/21/what-every-coder-should-know-about-gamma/
// https://ninedegreesbelow.com/photography/linear-gamma-blur-normal-blend.html
func Blend(bottom, top color.RGBA) color.RGBA {
	return color.RGBA{
		R: encodeGamma((decodeGamma(bottom.R)*uint32(255-top.A))/255 + decodeGamma(top.R)),
		G: encodeGamma((decodeGamma(bottom.G)*uint32(255-top.A))/255 + decodeGamma(top.G)),
		B: encodeGamma((decodeGamma(bottom.B)*uint32(255-top.A))/255 + decodeGamma(top.B)),
		A: 255,
	}
}

// ApplyAlpha takes a color (that may be semi-transparent) and applies the given
// alpha to it, making it even more transparent. It does so while taking gamma
// into account, see Blend.
func ApplyAlpha(c color.RGBA, alpha uint8) color.RGBA {
	return color.RGBA{
		R: encodeGamma(decodeGamma(c.R) * uint32(alpha) / 256),
		G: encodeGamma(decodeGamma(c.G) * uint32(alpha) / 256),
		B: encodeGamma(decodeGamma(c.B) * uint32(alpha) / 256),
		A: uint8(uint32(c.A) * uint32(alpha) / 256),
	}
}

// decodeGamma decodes a single 8-bit gamma-encoded (compressed) value to a
// mostly linear color intensity.
func decodeGamma(component uint8) uint32 {
	// This is the correct decoding formula:
	//     return math.Pow(float64(component)/255, 2.2)
	// However, pow is slow. So alternatively, there is this:
	//     return 0.8*f*f + 0.2*f*f*f
	// Source: https://stackoverflow.com/questions/48903716/fast-image-gamma-correction#48904006
	// But we want it even faster, so use a "close enough" gamma of 2.0 instead of 2.2.
	return uint32(component) * uint32(component)
}

// encodeGamma converts a linear color intensity to an 8-bit gamma-encoded
// (compressed) form.
func encodeGamma(x uint32) uint8 {
	// This is the correct encoding formula:
	//     return uint8(math.Pow(component, 1/2.2) * 255)
	// The following might be a little bit faster, and matches DecodeGamma:
	//     return uint8(math.Sqrt(x) * 256)
	// However, floating point is still slow (even float32). So use a fast
	// approximation instead. The code below roundtrips cleanly with EncodeGamma
	// for all 8-bit values and is very fast on a Cortex-M4.
	//
	// Original code copied from:
	// https://stackoverflow.com/questions/34187171/fast-integer-square-root-approximation/#34187992

	if x == 0 {
		// avoid division by 0
		return 0
	}

	// The starting value 32 results in ~8% faster code than most other starting
	// values. Note that this value has been selected because it roundtrips
	// cleanly with DecodeGamma and requires the least amount of correction rounds below.
	a := uint32(32)
	b := x / a
	a = (a + b) / 2
	b = x / a
	a = (a + b) / 2
	b = x / a
	a = (a + b) / 2
	b = x / a
	a = (a + b) / 2
	b = x / a
	a = (a + b) / 2

	return uint8(a)
}
