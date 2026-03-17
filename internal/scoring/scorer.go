package scoring

import (
	"fmt"
	"strings"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

// Calibration ranges: each variable is linearly normalized to [0, 1].
// See docs/scoring-formula.md for rationale.
const (
	tempMin = -30.0
	tempMax = 48.0
	humMin  = 10.0
	humMax  = 98.0
	dewMin  = -20.0
	dewMax  = 32.0
	pm25Min = 0.0
	pm25Max = 300.0
	pm10Min = 0.0
	pm10Max = 500.0
)

// Thickness weights (must sum to 1.0).
const (
	tempWeight = 0.30
	humWeight  = 0.30
	pm25Weight = 0.25
	pm10Weight = 0.15
)

// Irritation weights (must sum to 1.0).
const (
	pm25IrritationWeight = 0.65
	pm10IrritationWeight = 0.35
)

var requiredVariables = []string{"temperature", "humidity", "dewpoint", "pm2p5", "pm10"}

type Scorer struct{}

func NewScorer() *Scorer {
	return &Scorer{}
}

func (s *Scorer) RequiredVariables() []string {
	return requiredVariables
}

func (s *Scorer) Calculate(variableData []domain.VariableData) (domain.Score, error) {
	vars, err := buildVarMap(variableData)
	if err != nil {
		return domain.Score{}, err
	}

	normTemp := normalize(vars["temperature"], tempMin, tempMax)
	normHum := normalize(vars["humidity"], humMin, humMax)
	normDew := normalize(vars["dewpoint"], dewMin, dewMax)
	normPM25 := normalize(vars["pm2p5"], pm25Min, pm25Max)
	normPM10 := normalize(vars["pm10"], pm10Min, pm10Max)

	return domain.Score{
		Thickness:  tempWeight*normTemp + humWeight*normHum + pm25Weight*normPM25 + pm10Weight*normPM10,
		Sweatiness: normDew,
		Irritation: pm25IrritationWeight*normPM25 + pm10IrritationWeight*normPM10,
		Warmth:     normTemp,
	}, nil
}

func buildVarMap(variableData []domain.VariableData) (map[string]float64, error) {
	vars := make(map[string]float64, len(variableData))
	for _, v := range variableData {
		vars[v.Name] = v.Value
	}

	var missing []string
	for _, name := range requiredVariables {
		if _, ok := vars[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"scorer: missing required variable(s) [%s] — this is a bug; Jackfruit should have failed first",
			strings.Join(missing, ", "),
		)
	}

	return vars, nil
}

func normalize(value, lo, hi float64) float64 {
	n := (value - lo) / (hi - lo)
	if n < 0 {
		return 0
	}
	if n > 1 {
		return 1
	}
	return n
}
