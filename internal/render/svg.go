package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

// SVGRenderer generates parametric SVG butt visualizations driven by scores.
type SVGRenderer struct{}

func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{}
}

type bodyGeom struct {
	thickness                       float64
	topY, sideX, sideY, bottomY    float64
	cheekY, leftCheekX, rightCheekX float64
}

func newBodyGeom(t float64) bodyGeom {
	sideX := lerp(75, 30, t)
	return bodyGeom{
		thickness:   t,
		topY:        lerp(85, 50, t),
		sideX:       sideX,
		sideY:       lerp(130, 125, t),
		bottomY:     lerp(175, 215, t),
		cheekY:      lerp(135, 145, t),
		leftCheekX:  (120 + sideX) / 2,
		rightCheekX: (120 + (240 - sideX)) / 2,
	}
}

// Render is deterministic: same input always produces identical output.
func (r *SVGRenderer) Render(score domain.Score) (string, error) {
	var b strings.Builder
	geom := newBodyGeom(score.Thickness)

	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 240 260">`)

	writeGradientDefs(&b, score)
	writeBodyShape(&b, geom, score.Warmth)
	writeCrease(&b, geom, score.Warmth)
	writeBlush(&b, geom, score.Irritation)
	writeHighlights(&b, geom, score.Sweatiness)
	writeDroplets(&b, geom, score.Sweatiness)

	b.WriteString(`</svg>`)

	return b.String(), nil
}

func writeGradientDefs(b *strings.Builder, score domain.Score) {
	base := warmthColor(score.Warmth)
	center := base.withLightness(0.10)
	edge := base.withLightness(-0.05)

	b.WriteString(`<defs>`)

	fmt.Fprintf(b, `<radialGradient id="bodyGrad" cx="45%%" cy="40%%" r="60%%">`+
		`<stop offset="0%%" stop-color="%s"/>`+
		`<stop offset="100%%" stop-color="%s"/>`+
		`</radialGradient>`,
		center.toHex(), edge.toHex())

	if score.Irritation >= 0.05 {
		blushColor := blushHex(score.Irritation)
		opacity := ff(lerp(0.02, 0.55, score.Irritation))

		writeFadingGradient(b, "blushL", blushColor, opacity)
		writeFadingGradient(b, "blushR", blushColor, opacity)
	}

	if score.Sweatiness >= 0.05 {
		opacity := ff(lerp(0.02, 0.50, score.Sweatiness))

		writeFadingGradient(b, "highL", "#FFFFFF", opacity)
		writeFadingGradient(b, "highR", "#FFFFFF", opacity)
	}

	b.WriteString(`</defs>`)
}

func strokeColor(warmth float64) string {
	return warmthColor(warmth).withLightness(-0.15).withSaturation(0.10).toHex()
}

// blushHex interpolates from soft pink (#E08080) to angry red (#C02020).
func blushHex(irritation float64) string {
	c := hslColor{
		H: 0,
		S: lerp(0.55, 0.71, irritation),
		L: lerp(0.69, 0.44, irritation),
	}
	return c.toHex()
}

func writeFadingGradient(b *strings.Builder, id, color, opacity string) {
	fmt.Fprintf(b, `<radialGradient id="%s" cx="50%%" cy="50%%" r="50%%">`+
		`<stop offset="0%%" stop-color="%s" stop-opacity="%s"/>`+
		`<stop offset="100%%" stop-color="%s" stop-opacity="0"/>`+
		`</radialGradient>`, id, color, opacity, color)
}

func writeEllipse(b *strings.Builder, cx, cy float64, rx, ry, fill string) {
	fmt.Fprintf(b, `<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="%s"/>`,
		ff(cx), ff(cy), rx, ry, fill)
}

// writeBodyShape draws 4 cubic Bézier segments forming a peach shape, symmetric around x=120.
func writeBodyShape(b *strings.Builder, geom bodyGeom, warmth float64) {
	t := geom.thickness
	path := fmt.Sprintf(
		`M %s %s `+
			`C %s %s, %s %s, %s %s `+
			`C %s %s, %s %s, %s %s `+
			`C %s %s, %s %s, %s %s `+
			`C %s %s, %s %s, %s %s Z`,

		// Start: top center
		ff(120), ff(geom.topY),

		// Seg 1: top → left side
		ff(lerp(95, 75, t)), ff(geom.topY),
		ff(geom.sideX), ff(lerp(90, 75, t)),
		ff(geom.sideX), ff(geom.sideY),

		// Seg 2: left side → bottom center
		ff(geom.sideX), ff(lerp(165, 185, t)),
		ff(lerp(95, 80, t)), ff(lerp(185, 210, t)),
		ff(120), ff(geom.bottomY),

		// Seg 3: bottom center → right side (mirror)
		ff(240-lerp(95, 80, t)), ff(lerp(185, 210, t)),
		ff(240-geom.sideX), ff(lerp(165, 185, t)),
		ff(240-geom.sideX), ff(geom.sideY),

		// Seg 4: right side → top (mirror)
		ff(240-geom.sideX), ff(lerp(90, 75, t)),
		ff(240-lerp(95, 75, t)), ff(geom.topY),
		ff(120), ff(geom.topY),
	)

	strokeWidth := ff(lerp(2.0, 3.5, t))

	fmt.Fprintf(b, `<path d="%s" fill="url(#bodyGrad)" stroke="%s" stroke-width="%s" stroke-linejoin="round"/>`,
		path, strokeColor(warmth), strokeWidth)
}

func writeCrease(b *strings.Builder, geom bodyGeom, warmth float64) {
	t := geom.thickness

	creaseTop := geom.topY + lerp(15, 20, t)
	creaseBottom := geom.bottomY - lerp(5, 10, t)

	// Both control points shift left for a gentle arc.
	bow := lerp(1, 8, t)

	strokeWidth := ff(lerp(2.0, 4.0, t))

	fmt.Fprintf(b, `<path d="M %s %s C %s %s, %s %s, %s %s" fill="none" stroke="%s" stroke-width="%s" stroke-linecap="round"/>`,
		ff(120), ff(creaseTop),
		ff(120-bow), ff(creaseTop+(creaseBottom-creaseTop)*0.33),
		ff(120-bow), ff(creaseTop+(creaseBottom-creaseTop)*0.66),
		ff(120), ff(creaseBottom),
		strokeColor(warmth), strokeWidth)
}

func writeBlush(b *strings.Builder, geom bodyGeom, irritation float64) {
	if irritation < 0.05 {
		return
	}

	rx := ff(lerp(18, 35, irritation))
	ry := ff(lerp(14, 28, irritation))

	writeEllipse(b, geom.leftCheekX, geom.cheekY, rx, ry, "url(#blushL)")
	writeEllipse(b, geom.rightCheekX, geom.cheekY, rx, ry, "url(#blushR)")
}

func writeHighlights(b *strings.Builder, geom bodyGeom, sweatiness float64) {
	if sweatiness < 0.05 {
		return
	}

	rx := ff(lerp(16, 22, sweatiness))
	ry := ff(lerp(12, 16, sweatiness))

	writeEllipse(b, geom.leftCheekX-10, geom.cheekY-22, rx, ry, "url(#highL)")
	writeEllipse(b, geom.rightCheekX+10, geom.cheekY-22, rx, ry, "url(#highR)")
}

// writeDroplets uses a Vogel sunflower spiral mapped to an ellipse covering
// the lower half of the butt, with a parabolic U-lift so outer droplets rise
// toward the outline while center ones hang low.
func writeDroplets(b *strings.Builder, geom bodyGeom, sweatiness float64) {
	count := dropletCount(sweatiness)
	if count == 0 {
		return
	}

	halfWidth := (120 - geom.sideX) * 0.9
	centerY := geom.sideY + (geom.bottomY-geom.sideY)*0.65 // gravity bias
	halfHeight := (geom.bottomY - geom.sideY) / 2 * 0.75
	scale := lerp(0.7, 1.4, geom.thickness)

	goldenAngle := math.Pi * (3 - math.Sqrt(5)) // ≈137.5°
	uLift := (geom.bottomY - geom.sideY) * 0.55

	for i := 0; i < count; i++ {
		// Exponent 0.4 (< sqrt) pushes inner points outward at small N.
		r := math.Pow((float64(i)+0.5)/float64(count), 0.4)
		theta := float64(i)*goldenAngle + math.Pi/3.2 // phase offset breaks diagonal alignment

		x := 120 + r*math.Cos(theta)*halfWidth
		y := centerY + r*math.Sin(theta)*halfHeight

		xFrac := (x - 120) / halfWidth // [-1, 1]
		y -= uLift * xFrac * xFrac     // parabolic U-lift

		writeTeardrop(b, x, y, 0.55+0.15*r, scale)
	}
}

func dropletCount(sweatiness float64) int {
	if sweatiness < 0.25 {
		return 0
	}
	count := int(math.Round(lerp(1, 7, (sweatiness-0.25)/0.75)))
	if count > 7 {
		return 7
	}
	return count
}

func writeTeardrop(b *strings.Builder, x, y, opacity, scale float64) {
	h := 6 * scale    // half-height
	w := 4.5 * scale  // half-width
	bw := 1.5 * scale // bulge inset from top

	fmt.Fprintf(b, `<path d="M %s %s C %s %s, %s %s, %s %s C %s %s, %s %s, %s %s Z" fill="#D0E8FF" opacity="%s"/>`,
		ff(x), ff(y-h), // top point
		ff(x-w), ff(y-bw), // cp1: left bulge
		ff(x-w), ff(y+w), // cp2: left round bottom
		ff(x), ff(y+h), // bottom center
		ff(x+w), ff(y+w), // cp1: right round bottom
		ff(x+w), ff(y-bw), // cp2: right bulge
		ff(x), ff(y-h), // back to top
		ff(opacity))
}
