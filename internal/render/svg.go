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

// Render is deterministic: same input always produces identical output.
func (r *SVGRenderer) Render(score domain.Score) (string, error) {
	var b strings.Builder

	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 240 260">`)

	writeGradientDefs(&b, score)
	writeBodyShape(&b, score.Thickness, score.Warmth)
	writeCrease(&b, score.Thickness, score.Warmth)
	writeBlush(&b, score.Thickness, score.Irritation)
	writeHighlights(&b, score.Thickness, score.Sweatiness)
	writeDroplets(&b, score.Thickness, score.Sweatiness)

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

		fmt.Fprintf(b, `<radialGradient id="blushL" cx="50%%" cy="50%%" r="50%%">`+
			`<stop offset="0%%" stop-color="%s" stop-opacity="%s"/>`+
			`<stop offset="100%%" stop-color="%s" stop-opacity="0"/>`+
			`</radialGradient>`, blushColor, opacity, blushColor)

		fmt.Fprintf(b, `<radialGradient id="blushR" cx="50%%" cy="50%%" r="50%%">`+
			`<stop offset="0%%" stop-color="%s" stop-opacity="%s"/>`+
			`<stop offset="100%%" stop-color="%s" stop-opacity="0"/>`+
			`</radialGradient>`, blushColor, opacity, blushColor)
	}

	if score.Sweatiness >= 0.05 {
		opacity := ff(lerp(0.02, 0.50, score.Sweatiness))

		fmt.Fprintf(b, `<radialGradient id="highL" cx="50%%" cy="50%%" r="50%%">`+
			`<stop offset="0%%" stop-color="#FFFFFF" stop-opacity="%s"/>`+
			`<stop offset="100%%" stop-color="#FFFFFF" stop-opacity="0"/>`+
			`</radialGradient>`, opacity)

		fmt.Fprintf(b, `<radialGradient id="highR" cx="50%%" cy="50%%" r="50%%">`+
			`<stop offset="0%%" stop-color="#FFFFFF" stop-opacity="%s"/>`+
			`<stop offset="100%%" stop-color="#FFFFFF" stop-opacity="0"/>`+
			`</radialGradient>`, opacity)
	}

	b.WriteString(`</defs>`)
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

// writeBodyShape draws 4 cubic Bézier segments forming a peach shape, symmetric around x=120.
func writeBodyShape(b *strings.Builder, thickness, warmth float64) {
	t := thickness

	topY := lerp(85, 50, t)
	sideX := lerp(75, 30, t)
	sideY := lerp(130, 125, t)
	bottomY := lerp(175, 215, t)
	path := fmt.Sprintf(
		`M %s %s `+
			`C %s %s, %s %s, %s %s `+
			`C %s %s, %s %s, %s %s `+
			`C %s %s, %s %s, %s %s `+
			`C %s %s, %s %s, %s %s Z`,

		// Start: top center
		ff(120), ff(topY),

		// Seg 1: top → left side
		ff(lerp(95, 75, t)), ff(topY),
		ff(sideX), ff(lerp(90, 75, t)),
		ff(sideX), ff(sideY),

		// Seg 2: left side → bottom center
		ff(sideX), ff(lerp(165, 185, t)),
		ff(lerp(95, 80, t)), ff(lerp(185, 210, t)),
		ff(120), ff(bottomY),

		// Seg 3: bottom center → right side (mirror)
		ff(240-lerp(95, 80, t)), ff(lerp(185, 210, t)),
		ff(240-sideX), ff(lerp(165, 185, t)),
		ff(240-sideX), ff(sideY),

		// Seg 4: right side → top (mirror)
		ff(240-sideX), ff(lerp(90, 75, t)),
		ff(240-lerp(95, 75, t)), ff(topY),
		ff(120), ff(topY),
	)

	base := warmthColor(warmth)
	strokeColor := base.withLightness(-0.15).withSaturation(0.10).toHex()
	strokeWidth := ff(lerp(2.0, 3.5, thickness))

	fmt.Fprintf(b, `<path d="%s" fill="url(#bodyGrad)" stroke="%s" stroke-width="%s" stroke-linejoin="round"/>`,
		path, strokeColor, strokeWidth)
}

func writeCrease(b *strings.Builder, thickness, warmth float64) {
	t := thickness

	topY := lerp(85, 50, t)
	bottomY := lerp(175, 215, t)

	creaseTop := topY + lerp(15, 20, t)
	creaseBottom := bottomY - lerp(5, 10, t)

	// Both control points shift left for a gentle arc.
	bow := lerp(1, 8, t)

	base := warmthColor(warmth)
	strokeColor := base.withLightness(-0.15).withSaturation(0.10).toHex()
	strokeWidth := ff(lerp(2.0, 4.0, t))

	fmt.Fprintf(b, `<path d="M %s %s C %s %s, %s %s, %s %s" fill="none" stroke="%s" stroke-width="%s" stroke-linecap="round"/>`,
		ff(120), ff(creaseTop),
		ff(120-bow), ff(creaseTop+(creaseBottom-creaseTop)*0.33),
		ff(120-bow), ff(creaseTop+(creaseBottom-creaseTop)*0.66),
		ff(120), ff(creaseBottom),
		strokeColor, strokeWidth)
}

func writeBlush(b *strings.Builder, thickness, irritation float64) {
	if irritation < 0.05 {
		return
	}

	t := thickness

	sideX := lerp(75, 30, t)
	leftCX := (120 + sideX) / 2
	rightCX := (120 + (240 - sideX)) / 2
	cheekY := lerp(135, 145, t)

	rx := ff(lerp(18, 35, irritation))
	ry := ff(lerp(14, 28, irritation))

	fmt.Fprintf(b, `<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="url(#blushL)"/>`,
		ff(leftCX), ff(cheekY), rx, ry)
	fmt.Fprintf(b, `<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="url(#blushR)"/>`,
		ff(rightCX), ff(cheekY), rx, ry)
}

func writeHighlights(b *strings.Builder, thickness, sweatiness float64) {
	if sweatiness < 0.05 {
		return
	}

	t := thickness

	sideX := lerp(75, 30, t)
	leftCX := (120+sideX)/2 - 10
	rightCX := (120+(240-sideX))/2 + 10
	cheekY := lerp(135, 145, t) - 22

	rx := ff(lerp(16, 22, sweatiness))
	ry := ff(lerp(12, 16, sweatiness))

	fmt.Fprintf(b, `<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="url(#highL)"/>`,
		ff(leftCX), ff(cheekY), rx, ry)
	fmt.Fprintf(b, `<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="url(#highR)"/>`,
		ff(rightCX), ff(cheekY), rx, ry)
}

// writeDroplets uses a Vogel sunflower spiral mapped to an ellipse covering
// the lower half of the butt, with a parabolic U-lift so outer droplets rise
// toward the outline while center ones hang low.
func writeDroplets(b *strings.Builder, thickness, sweatiness float64) {
	count := dropletCount(sweatiness)
	if count == 0 {
		return
	}

	t := thickness

	sideX := lerp(75, 30, t)
	halfWidth := (120 - sideX) * 0.9
	midY := lerp(130, 125, t)
	bottomY := lerp(175, 215, t)
	centerY := midY + (bottomY-midY)*0.65 // gravity bias
	halfHeight := (bottomY - midY) / 2 * 0.75
	scale := lerp(0.7, 1.4, t)

	goldenAngle := math.Pi * (3 - math.Sqrt(5)) // ≈137.5°
	uLift := (bottomY - midY) * 0.55

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
