package render

import (
	"fmt"
	"math"
)

type hslColor struct {
	H float64 // [0, 360)
	S float64 // [0, 1]
	L float64 // [0, 1]
}

type colorStop struct {
	t     float64
	color hslColor
}

var warmthRamp = []colorStop{
	{0.0, hslColor{215, 0.35, 0.72}},
	{0.12, hslColor{215, 0.12, 0.77}},
	{0.18, hslColor{25, 0.04, 0.79}},
	{0.4, hslColor{25, 0.25, 0.78}},
	{0.5, hslColor{25, 0.50, 0.76}},
	{0.7, hslColor{15, 0.55, 0.65}},
	{1.0, hslColor{8, 0.62, 0.60}},
}

// toHex converts an HSL color to a "#RRGGBB" string.
func (c hslColor) toHex() string {
	h := math.Mod(c.H, 360)
	if h < 0 {
		h += 360
	}

	chroma := (1 - math.Abs(2*c.L-1)) * c.S
	hPrime := h / 60
	x := chroma * (1 - math.Abs(math.Mod(hPrime, 2)-1))

	var r1, g1, b1 float64
	switch {
	case hPrime < 1:
		r1, g1, b1 = chroma, x, 0
	case hPrime < 2:
		r1, g1, b1 = x, chroma, 0
	case hPrime < 3:
		r1, g1, b1 = 0, chroma, x
	case hPrime < 4:
		r1, g1, b1 = 0, x, chroma
	case hPrime < 5:
		r1, g1, b1 = x, 0, chroma
	default:
		r1, g1, b1 = chroma, 0, x
	}

	m := c.L - chroma/2
	r := clamp01(r1+m) * 255
	g := clamp01(g1+m) * 255
	b := clamp01(b1+m) * 255

	return fmt.Sprintf("#%02X%02X%02X", int(math.Round(r)), int(math.Round(g)), int(math.Round(b)))
}

// withLightness returns a copy with lightness adjusted by delta, clamped to [0, 1].
func (c hslColor) withLightness(delta float64) hslColor {
	return hslColor{H: c.H, S: c.S, L: clamp01(c.L + delta)}
}

// withSaturation returns a copy with saturation adjusted by delta, clamped to [0, 1].
func (c hslColor) withSaturation(delta float64) hslColor {
	return hslColor{H: c.H, S: clamp01(c.S + delta), L: c.L}
}

// lerpHue interpolates between two hue values using the shortest arc on the color wheel.
func lerpHue(h1, h2, t float64) float64 {
	delta := h2 - h1
	if delta > 180 {
		h1 += 360
	} else if delta < -180 {
		h2 += 360
	}
	h := h1 + t*(h2-h1)
	return math.Mod(h+360, 360)
}

// warmthColor returns the interpolated HSL color for a warmth value in [0, 1].
func warmthColor(warmth float64) hslColor {
	warmth = clamp01(warmth)

	// At or beyond the last stop.
	if warmth >= warmthRamp[len(warmthRamp)-1].t {
		return warmthRamp[len(warmthRamp)-1].color
	}

	// Find the two surrounding stops.
	for i := 0; i < len(warmthRamp)-1; i++ {
		lo := warmthRamp[i]
		hi := warmthRamp[i+1]
		if warmth >= lo.t && warmth <= hi.t {
			localT := (warmth - lo.t) / (hi.t - lo.t)
			return hslColor{
				H: lerpHue(lo.color.H, hi.color.H, localT),
				S: lerp(lo.color.S, hi.color.S, localT),
				L: lerp(lo.color.L, hi.color.L, localT),
			}
		}
	}

	return warmthRamp[0].color
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
