package scoring

import (
	"fmt"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

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

	tempWeight = 0.30
	humWeight  = 0.30
	pm25Weight = 0.25
	pm10Weight = 0.15

	pm25IrritationWeight = 0.65
	pm10IrritationWeight = 0.35
)

type Scorer struct{}

func NewScorer() *Scorer {
	return &Scorer{}
}

func (s *Scorer) RequiredVariables() []string {
	return []string{"temperature", "humidity", "dewpoint", "pm2p5", "pm10"}
}

func (s *Scorer) Calculate(variableData []domain.VariableData) (domain.Score, error) {
	temp, ok := findVariable(variableData, "temperature")
	if !ok {
		return domain.Score{}, fmt.Errorf("scorer: missing required variable %q — this is a bug; Jackfruit should have failed first", "temperature")
	}
	hum, ok := findVariable(variableData, "humidity")
	if !ok {
		return domain.Score{}, fmt.Errorf("scorer: missing required variable %q — this is a bug; Jackfruit should have failed first", "humidity")
	}
	dew, ok := findVariable(variableData, "dewpoint")
	if !ok {
		return domain.Score{}, fmt.Errorf("scorer: missing required variable %q — this is a bug; Jackfruit should have failed first", "dewpoint")
	}
	pm25, ok := findVariable(variableData, "pm2p5")
	if !ok {
		return domain.Score{}, fmt.Errorf("scorer: missing required variable %q — this is a bug; Jackfruit should have failed first", "pm2p5")
	}
	pm10, ok := findVariable(variableData, "pm10")
	if !ok {
		return domain.Score{}, fmt.Errorf("scorer: missing required variable %q — this is a bug; Jackfruit should have failed first", "pm10")
	}

	normTemp := normalize(temp, tempMin, tempMax)
	normHum := normalize(hum, humMin, humMax)
	normDew := normalize(dew, dewMin, dewMax)
	normPM25 := normalize(pm25, pm25Min, pm25Max)
	normPM10 := normalize(pm10, pm10Min, pm10Max)

	return domain.Score{
		Thickness:  tempWeight*normTemp + humWeight*normHum + pm25Weight*normPM25 + pm10Weight*normPM10,
		Sweatiness: normDew,
		Irritation: pm25IrritationWeight*normPM25 + pm10IrritationWeight*normPM10,
		Warmth:     normTemp,
	}, nil
}

func findVariable(vars []domain.VariableData, name string) (float64, bool) {
	for _, v := range vars {
		if v.Name == name {
			return v.Value, true
		}
	}
	return 0, false
}

func normalize(value, min, max float64) float64 {
	n := (value - min) / (max - min)
	if n < 0 {
		return 0
	}
	if n > 1 {
		return 1
	}
	return n
}
