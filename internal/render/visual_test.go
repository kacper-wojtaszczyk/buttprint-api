package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

func TestVisualSamples(t *testing.T) {
	if os.Getenv("WRITE_SVG") == "" {
		t.Skip("set WRITE_SVG=1 to generate visual samples")
	}

	dir := "../../_visual_output"
	r := NewSVGRenderer()

	levels := []float64{0, 0.5, 1}

	var gallery strings.Builder
	gallery.WriteString(galleryHeader)
	gallery.WriteString("<div class=\"grid\">\n")

	for _, th := range levels {
		for _, sw := range levels {
			for _, ir := range levels {
				for _, wa := range levels {
					s := domain.Score{
						Thiccness:  th,
						Sweatiness: sw,
						Irritation: ir,
						Warmth:     wa,
					}
					name := fmt.Sprintf("t%.1f_s%.1f_i%.1f_w%.1f", th, sw, ir, wa)

					svg, err := r.Render(s)
					if err != nil {
						t.Errorf("error rendering %s: %v", name, err)
						continue
					}

					path := filepath.Join(dir, name+".svg")
					if err := os.WriteFile(path, []byte(svg), 0644); err != nil {
						t.Errorf("error writing %s: %v", path, err)
						continue
					}
					t.Logf("wrote %s", path)

					label := fmt.Sprintf("T%.0f S%.0f I%.0f W%.0f", th*10, sw*10, ir*10, wa*10)
					fmt.Fprintf(&gallery,
						"  <div class=\"sample\">\n    <img src=\"%s.svg\">\n    <div class=\"name\">%s</div>\n  </div>\n",
						name, label)
				}
			}
		}
	}

	gallery.WriteString("</div>\n")
	gallery.WriteString("</body>\n</html>\n")

	htmlPath := filepath.Join(dir, "gallery.html")
	if err := os.WriteFile(htmlPath, []byte(gallery.String()), 0644); err != nil {
		t.Errorf("error writing %s: %v", htmlPath, err)
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
  h2 {
    font-weight: 400;
    font-size: 12px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #7a726a;
    margin: 32px 0 16px;
  }
  .grid {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
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
</style>
</head>
<body>

`
