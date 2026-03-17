package render

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

func TestVisualSamples(t *testing.T) {
	if os.Getenv("WRITE_SVG") == "" {
		t.Skip("set WRITE_SVG=1 to generate visual samples")
	}

	r := NewSVGRenderer()
	samples := []struct {
		name  string
		score domain.Score
	}{
		{"all_zero", domain.Score{Thickness: 0, Sweatiness: 0, Irritation: 0, Warmth: 0}},
		{"all_one", domain.Score{Thickness: 1, Sweatiness: 1, Irritation: 1, Warmth: 1}},
		{"mid", domain.Score{Thickness: 0.5, Sweatiness: 0.5, Irritation: 0.5, Warmth: 0.5}},
		{"cold_thin", domain.Score{Thickness: 0.1, Sweatiness: 0.1, Irritation: 0, Warmth: 0.1}},
		{"hot_thick", domain.Score{Thickness: 0.9, Sweatiness: 0.8, Irritation: 0.7, Warmth: 0.95}},
		{"sweaty_cold", domain.Score{Thickness: 0.3, Sweatiness: 0.9, Irritation: 0.1, Warmth: 0.2}},
		{"irritated_warm", domain.Score{Thickness: 0.6, Sweatiness: 0.1, Irritation: 0.9, Warmth: 0.7}},

		// Droplet count series: one sample per count (0–7) at mid thickness/warmth.
		{"drops_0", domain.Score{Thickness: 0.5, Sweatiness: 0.00, Irritation: 0, Warmth: 0.8}},
		{"drops_1", domain.Score{Thickness: 0.5, Sweatiness: 0.25, Irritation: 0, Warmth: 0.8}},
		{"drops_2", domain.Score{Thickness: 0.5, Sweatiness: 0.32, Irritation: 0, Warmth: 0.8}},
		{"drops_3", domain.Score{Thickness: 0.5, Sweatiness: 0.44, Irritation: 0, Warmth: 0.8}},
		{"drops_4", domain.Score{Thickness: 0.5, Sweatiness: 0.57, Irritation: 0, Warmth: 0.8}},
		{"drops_5", domain.Score{Thickness: 0.5, Sweatiness: 0.69, Irritation: 0, Warmth: 0.8}},
		{"drops_6", domain.Score{Thickness: 0.5, Sweatiness: 0.82, Irritation: 0, Warmth: 0.8}},
		{"drops_7", domain.Score{Thickness: 0.5, Sweatiness: 0.94, Irritation: 0, Warmth: 0.8}},
	}

	dir := "../../testdata"
	for _, s := range samples {
		svg, err := r.Render(s.score)
		if err != nil {
			t.Errorf("error rendering %s: %v", s.name, err)
			continue
		}
		path := filepath.Join(dir, s.name+".svg")
		if err := os.WriteFile(path, []byte(svg), 0644); err != nil {
			t.Errorf("error writing %s: %v", path, err)
			continue
		}
		t.Logf("wrote %s (%s)", path, fmt.Sprintf("T=%.2f S=%.2f I=%.2f W=%.2f",
			s.score.Thickness, s.score.Sweatiness, s.score.Irritation, s.score.Warmth))
	}
}
