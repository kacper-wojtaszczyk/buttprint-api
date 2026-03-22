package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/scoring"
)

type scenario struct {
	name         string
	description  string
	variableData []domain.VariableData
}

func makeVars(temp, hum, dew, pm25, pm10 float64) []domain.VariableData {
	return []domain.VariableData{
		{Name: "temperature", Value: temp, Unit: "°C"},
		{Name: "humidity", Value: hum, Unit: "%"},
		{Name: "dewpoint", Value: dew, Unit: "°C"},
		{Name: "pm2p5", Value: pm25, Unit: "µg/m³"},
		{Name: "pm10", Value: pm10, Unit: "µg/m³"},
	}
}

var scenarios = []scenario{
	{
		name:         "siberian_winter",
		description:  "Siberian Winter · −20°C, 25% RH",
		variableData: makeVars(-20, 25, -25, 5, 15),
	},
	{
		name:         "temperate_spring",
		description:  "Temperate Spring · 15°C, 60% RH",
		variableData: makeVars(15, 60, 8, 20, 40),
	},
	{
		name:         "mediterranean_summer",
		description:  "Mediterranean Summer · 28°C, 45% RH",
		variableData: makeVars(28, 45, 14, 15, 30),
	},
	{
		name:         "new_orleans_summer",
		description:  "New Orleans Summer · 30°C, 85% RH",
		variableData: makeVars(30, 85, 26, 35, 60),
	},
	{
		name:         "humid_tropics",
		description:  "Humid Tropics · 32°C, 92% RH, PM2.5 45",
		variableData: makeVars(32, 92, 29, 45, 90),
	},
	{
		name:         "beijing_winter_smog",
		description:  "Beijing Winter Smog · −5°C, PM2.5 150",
		variableData: makeVars(-5, 30, -15, 150, 280),
	},
	{
		name:         "delhi_haze",
		description:  "Delhi Haze · 35°C, 70% RH, PM2.5 200",
		variableData: makeVars(35, 70, 20, 200, 350),
	},
	{
		name:         "wildfire_smoke",
		description:  "Wildfire Smoke · 25°C, 35% RH, PM2.5 200",
		variableData: makeVars(25, 35, 5, 200, 350),
	},
}

func TestVisualSamples(t *testing.T) {
	if os.Getenv("WRITE_SVG") == "" {
		t.Skip("set WRITE_SVG=1 to generate visual samples")
	}

	dir := "../../_visual_output"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}

	scorer := scoring.NewScorer()
	renderer := NewSVGRenderer()

	var gallery strings.Builder
	gallery.WriteString(galleryHeader)
	gallery.WriteString("<div class=\"grid\">\n")

	for _, sc := range scenarios {
		score, err := scorer.Calculate(sc.variableData)
		if err != nil {
			t.Errorf("scoring %s: %v", sc.name, err)
			continue
		}

		svg, err := renderer.Render(score)
		if err != nil {
			t.Errorf("rendering %s: %v", sc.name, err)
			continue
		}

		path := filepath.Join(dir, sc.name+".svg")
		if err := os.WriteFile(path, []byte(svg), 0644); err != nil {
			t.Errorf("writing %s: %v", path, err)
			continue
		}
		t.Logf("wrote %s (T%.2f S%.2f I%.2f W%.2f)", path,
			score.Thiccness, score.Sweatiness, score.Irritation, score.Warmth)

		fmt.Fprintf(&gallery,
			"  <div class=\"sample\">\n    <img src=\"%s.svg\">\n    <div class=\"name\">%s</div>\n    <div class=\"scores\">T%.2f S%.2f I%.2f W%.2f</div>\n  </div>\n",
			sc.name, sc.description,
			score.Thiccness, score.Sweatiness, score.Irritation, score.Warmth)
	}

	gallery.WriteString("</div>\n")
	gallery.WriteString("</body>\n</html>\n")

	htmlPath := filepath.Join(dir, "gallery.html")
	if err := os.WriteFile(htmlPath, []byte(gallery.String()), 0644); err != nil {
		t.Errorf("writing %s: %v", htmlPath, err)
	}
	t.Logf("wrote %s", htmlPath)
}

const galleryHeader = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Buttprint Visual Gallery</title>
<style>
  body {
    margin: 0;
    padding: 32px;
    background: #2a2725;
    font-family: "SF Mono", "Fira Code", monospace;
    color: #b0a89f;
  }
  .grid {
    display: flex;
    flex-wrap: wrap;
    gap: 24px;
  }
  .sample {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }
  .sample img {
    width: 120px;
    height: 132px;
  }
  .sample .name {
    font-size: 12px;
    color: #d4ccc4;
  }
  .sample .scores {
    font-size: 10px;
    color: #7a726a;
  }
</style>
</head>
<body>

`
