package render

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

var renderer = NewSVGRenderer()

func mustRender(t *testing.T, score domain.Score) string {
	t.Helper()
	svg, err := renderer.Render(score)
	if err != nil {
		t.Fatalf("Render(%+v) returned error: %v", score, err)
	}
	return svg
}

func assertValidXML(t *testing.T, svg string) {
	t.Helper()
	var parsed interface{}
	if err := xml.Unmarshal([]byte(svg), &parsed); err != nil {
		t.Errorf("SVG is not valid XML: %v\nSVG (first 500 chars): %s", err, svg[:min(len(svg), 500)])
	}
}

func assertContains(t *testing.T, svg, substr string) {
	t.Helper()
	if !strings.Contains(svg, substr) {
		t.Errorf("SVG should contain %q but does not.\nSVG (first 500 chars): %s", substr, svg[:min(len(svg), 500)])
	}
}

func assertNotContains(t *testing.T, svg, substr string) {
	t.Helper()
	if strings.Contains(svg, substr) {
		t.Errorf("SVG should NOT contain %q but does", substr)
	}
}

// score is a shorthand constructor to keep test literals concise.
func score(thiccness, sweatiness, irritation, warmth float64) domain.Score {
	return domain.Score{
		Thiccness:  thiccness,
		Sweatiness: sweatiness,
		Irritation: irritation,
		Warmth:     warmth,
	}
}

func TestRender_XMLValidity(t *testing.T) {
	scores := []domain.Score{
		score(0, 0, 0, 0),
		score(1, 1, 1, 1),
		score(0.5, 0.5, 0.5, 0.5),
		score(1, 0, 1, 0),
		score(0, 1, 0, 1),
	}
	for _, s := range scores {
		t.Run("", func(t *testing.T) {
			svg := mustRender(t, s)
			assertValidXML(t, svg)
		})
	}
}

func TestRender_Determinism(t *testing.T) {
	s := score(0.73, 0.41, 0.88, 0.55)
	svg1 := mustRender(t, s)
	svg2 := mustRender(t, s)
	if svg1 != svg2 {
		t.Error("same score produced different SVG outputs")
	}
}

func TestRender_DifferentScoresProduceDifferentSVGs(t *testing.T) {
	scores := []domain.Score{
		score(0, 0, 0, 0),
		score(1, 1, 1, 1),
		score(0.5, 0.5, 0.5, 0.5),
	}
	svgs := make([]string, len(scores))
	for i, s := range scores {
		svgs[i] = mustRender(t, s)
	}
	for i := 0; i < len(svgs); i++ {
		for j := i + 1; j < len(svgs); j++ {
			if svgs[i] == svgs[j] {
				t.Errorf("scores %+v and %+v produced identical SVGs", scores[i], scores[j])
			}
		}
	}
}

func TestRender_StructuralAssertions(t *testing.T) {
	svg := mustRender(t, score(0.5, 0.5, 0.5, 0.5))

	assertContains(t, svg, `<svg`)
	assertContains(t, svg, `xmlns="http://www.w3.org/2000/svg"`)
	assertContains(t, svg, `viewBox="0 0 240 260"`)
	assertContains(t, svg, `<path`) // body
	assertContains(t, svg, `</svg>`)

	// No width/height on the SVG element (stroke-width is fine).
	svgTag := svg[:strings.Index(svg, ">")]
	if strings.Contains(svgTag, ` width=`) || strings.Contains(svgTag, ` height=`) {
		t.Error("SVG element should not have width or height attributes")
	}
}

func TestRender_AllZeros(t *testing.T) {
	svg := mustRender(t, score(0, 0, 0, 0))
	assertValidXML(t, svg)

	// Body must be present.
	assertContains(t, svg, `<path`)
	assertContains(t, svg, `url(#bodyGrad)`)

	// No blush (irritation < 0.05).
	assertNotContains(t, svg, `url(#blushL)`)
	assertNotContains(t, svg, `url(#blushR)`)

	// No highlights (sweatiness < 0.05).
	assertNotContains(t, svg, `url(#highL)`)
	assertNotContains(t, svg, `url(#highR)`)

	// No droplets (sweatiness < 0.25).
	assertNotContains(t, svg, `#D0E8FF`)
}

func TestRender_AllOnes(t *testing.T) {
	svg := mustRender(t, score(1, 1, 1, 1))
	assertValidXML(t, svg)

	assertContains(t, svg, `url(#bodyGrad)`)
	assertContains(t, svg, `url(#blushL)`)
	assertContains(t, svg, `url(#blushR)`)
	assertContains(t, svg, `url(#highL)`)
	assertContains(t, svg, `url(#highR)`)
	assertContains(t, svg, `#D0E8FF`)
}

func TestRender_NoBlushAtZeroIrritation(t *testing.T) {
	svg := mustRender(t, score(0.5, 0.5, 0.0, 0.5))
	assertNotContains(t, svg, `url(#blushL)`)
	assertNotContains(t, svg, `url(#blushR)`)
}

func TestRender_BlushAppearsAboveThreshold(t *testing.T) {
	svg := mustRender(t, score(0.5, 0.5, 0.1, 0.5))
	assertContains(t, svg, `url(#blushL)`)
	assertContains(t, svg, `url(#blushR)`)
}

func TestRender_NoHighlightsAtZeroSweatiness(t *testing.T) {
	svg := mustRender(t, score(0.5, 0.0, 0.5, 0.5))
	assertNotContains(t, svg, `url(#highL)`)
	assertNotContains(t, svg, `url(#highR)`)
}

func TestRender_HighlightsAppearAboveThreshold(t *testing.T) {
	svg := mustRender(t, score(0.5, 0.1, 0.5, 0.5))
	assertContains(t, svg, `url(#highL)`)
	assertContains(t, svg, `url(#highR)`)
}

func TestRender_NoDropletsBelowThreshold(t *testing.T) {
	svg := mustRender(t, score(0.5, 0.2, 0.5, 0.5))
	assertNotContains(t, svg, `#D0E8FF`)
}

func TestRender_DropletsAppearAboveThreshold(t *testing.T) {
	svg := mustRender(t, score(0.5, 0.5, 0.5, 0.5))
	assertContains(t, svg, `#D0E8FF`)
}

func TestRender_MaxDropletsAtHighSweatiness(t *testing.T) {
	svg := mustRender(t, score(0.5, 1.0, 0.5, 0.5))
	count := strings.Count(svg, `#D0E8FF`)
	if count != 7 {
		t.Errorf("expected 7 droplets at sweatiness=1.0, got %d occurrences of #D0E8FF", count)
	}
}

func TestRender_MixedExtremes(t *testing.T) {
	// High thiccness + irritation, low sweatiness + warmth.
	svg1 := mustRender(t, score(1, 0, 1, 0))
	assertValidXML(t, svg1)
	assertContains(t, svg1, `url(#blushL)`)   // irritation=1
	assertNotContains(t, svg1, `#D0E8FF`)     // sweatiness=0
	assertNotContains(t, svg1, `url(#highL)`) // sweatiness=0

	// Low thiccness + irritation, high sweatiness + warmth.
	svg2 := mustRender(t, score(0, 1, 0, 1))
	assertValidXML(t, svg2)
	assertNotContains(t, svg2, `url(#blushL)`) // irritation=0
	assertContains(t, svg2, `#D0E8FF`)         // sweatiness=1
	assertContains(t, svg2, `url(#highL)`)     // sweatiness=1
}

func TestDropletCount(t *testing.T) {
	tests := []struct {
		sweatiness float64
		want       int
	}{
		{0.0, 0},
		{0.24, 0},
		{0.25, 1},
		{0.40, 2},
		{0.55, 3},
		{0.70, 5},
		{0.85, 6},
		{1.0, 7},
	}
	for _, tt := range tests {
		got := dropletCount(tt.sweatiness)
		if got != tt.want {
			t.Errorf("dropletCount(%g) = %d, want %d", tt.sweatiness, got, tt.want)
		}
	}
}
