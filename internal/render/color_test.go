package render

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func closeTo(t *testing.T, got, want float64, msg string) {
	t.Helper()
	if math.Abs(got-want) > epsilon {
		t.Errorf("%s: got %f, want %f", msg, got, want)
	}
}

func TestHSLToHex_KnownValues(t *testing.T) {
	tests := []struct {
		name string
		c    hslColor
		want string
	}{
		{"pure red", hslColor{0, 1.0, 0.5}, "#FF0000"},
		{"pure green", hslColor{120, 1.0, 0.5}, "#00FF00"},
		{"pure blue", hslColor{240, 1.0, 0.5}, "#0000FF"},
		{"black", hslColor{0, 0, 0}, "#000000"},
		{"white", hslColor{0, 0, 1}, "#FFFFFF"},
		{"gray 50%", hslColor{0, 0, 0.5}, "#808080"},
		{"yellow", hslColor{60, 1.0, 0.5}, "#FFFF00"},
		{"cyan", hslColor{180, 1.0, 0.5}, "#00FFFF"},
		{"magenta", hslColor{300, 1.0, 0.5}, "#FF00FF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.toHex()
			if got != tt.want {
				t.Errorf("toHex() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestWithLightness(t *testing.T) {
	base := hslColor{180, 0.5, 0.5}

	brighter := base.withLightness(0.2)
	closeTo(t, brighter.L, 0.7, "L+0.2")
	closeTo(t, brighter.H, 180, "H unchanged")
	closeTo(t, brighter.S, 0.5, "S unchanged")

	darker := base.withLightness(-0.3)
	closeTo(t, darker.L, 0.2, "L-0.3")

	// Clamp at upper bound.
	clamped := base.withLightness(0.8)
	closeTo(t, clamped.L, 1.0, "L clamped to 1")

	// Clamp at lower bound.
	clampedLow := base.withLightness(-0.9)
	closeTo(t, clampedLow.L, 0.0, "L clamped to 0")
}

func TestWithSaturation(t *testing.T) {
	base := hslColor{180, 0.5, 0.5}

	more := base.withSaturation(0.3)
	closeTo(t, more.S, 0.8, "S+0.3")

	less := base.withSaturation(-0.4)
	closeTo(t, less.S, 0.1, "S-0.4")

	clamped := base.withSaturation(0.9)
	closeTo(t, clamped.S, 1.0, "S clamped to 1")

	clampedLow := base.withSaturation(-0.8)
	closeTo(t, clampedLow.S, 0.0, "S clamped to 0")
}

func TestLerp(t *testing.T) {
	closeTo(t, lerp(0, 10, 0.5), 5.0, "midpoint")
	closeTo(t, lerp(0, 10, 0.0), 0.0, "t=0")
	closeTo(t, lerp(0, 10, 1.0), 10.0, "t=1")
	closeTo(t, lerp(5, 15, 0.3), 8.0, "offset range")
}

func TestLerpHue_ShortestArc(t *testing.T) {
	tests := []struct {
		name      string
		h1, h2, t float64
		want      float64
	}{
		{"same side", 10, 50, 0.5, 30},
		{"cross zero forward", 350, 10, 0.5, 0},
		{"cross zero backward", 10, 350, 0.5, 0},
		{"ramp 270 to 25", 270, 25, 0.5, 327.5},
		{"ramp 270 to 25 at t=0", 270, 25, 0, 270},
		{"ramp 270 to 25 at t=1", 270, 25, 1, 25},
		{"identical", 90, 90, 0.5, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lerpHue(tt.h1, tt.h2, tt.t)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("lerpHue(%g, %g, %g) = %g, want %g", tt.h1, tt.h2, tt.t, got, tt.want)
			}
		})
	}
}

func TestWarmthColor_RampStops(t *testing.T) {
	for _, stop := range warmthRamp {
		c := warmthColor(stop.t)
		closeTo(t, c.H, stop.color.H, "H at t="+ff(stop.t))
		closeTo(t, c.S, stop.color.S, "S at t="+ff(stop.t))
		closeTo(t, c.L, stop.color.L, "L at t="+ff(stop.t))
	}
}

func TestWarmthColor_Interpolation(t *testing.T) {
	// Midpoint between stop 0 (t=0.0) and stop 1 (t=0.3) is t=0.15.
	c := warmthColor(0.15)
	// S should be between 0.50 and 0.35.
	if c.S < 0.35 || c.S > 0.50 {
		t.Errorf("S at t=0.15 = %f, want between 0.35 and 0.50", c.S)
	}
	// L should be between 0.62 and 0.65.
	if c.L < 0.62 || c.L > 0.65 {
		t.Errorf("L at t=0.15 = %f, want between 0.62 and 0.65", c.L)
	}
}

func TestWarmthColor_Boundaries(t *testing.T) {
	// At 0.0, should match first stop.
	c0 := warmthColor(0.0)
	closeTo(t, c0.H, 215, "H at 0")

	// At 1.0, should match last stop.
	c1 := warmthColor(1.0)
	closeTo(t, c1.H, 8, "H at 1")

	// Below 0 clamps.
	cNeg := warmthColor(-0.5)
	closeTo(t, cNeg.H, 215, "H below 0 clamped")

	// Above 1 clamps.
	cOver := warmthColor(1.5)
	closeTo(t, cOver.H, 8, "H above 1 clamped")
}

func TestFf(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{3.14159, "3.1"},
		{0.0, "0.0"},
		{120.0, "120.0"},
		{99.99, "100.0"},
		{0.05, "0.1"},
	}

	for _, tt := range tests {
		got := ff(tt.v)
		if got != tt.want {
			t.Errorf("ff(%g) = %s, want %s", tt.v, got, tt.want)
		}
	}
}
