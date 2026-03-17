package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Buttprint struct {
	Variables []VariableData
	Score     Score
	SVG       string
}

type VariableData struct {
	Name         string
	Value        float64
	Unit         string
	RefTimestamp time.Time
	ActualLat    float64
	ActualLon    float64
	Lineage      *Lineage
}

type Lineage struct {
	Source    string
	Dataset   string
	RawFileID uuid.UUID
}

type Score struct {
	Thickness  float64
	Warmth     float64
	Sweatiness float64
	Irritation float64
}

type EnvironmentalDataProvider interface {
	GetEnvironmentalData(ctx context.Context, lat, lon float64, timestamp time.Time, variables []string) ([]VariableData, error)
}

type Scorer interface {
	Calculate(variableData []VariableData) (Score, error)
	RequiredVariables() []string
}

type Renderer interface {
	Render(Score) (string, error)
}

type Service struct {
	environmentalDataProvider EnvironmentalDataProvider
	scorer                    Scorer
	renderer                  Renderer
}

func NewService(environmentalDataProvider EnvironmentalDataProvider, scorer Scorer, renderer Renderer) *Service {
	return &Service{
		environmentalDataProvider: environmentalDataProvider,
		scorer:                    scorer,
		renderer:                  renderer,
	}
}

func (s *Service) GetButtprint(ctx context.Context, lat, lon float64, timestamp time.Time) (Buttprint, error) {
	variables, err := s.environmentalDataProvider.GetEnvironmentalData(ctx, lat, lon, timestamp, s.scorer.RequiredVariables())
	if err != nil {
		return Buttprint{}, err
	}

	score, err := s.scorer.Calculate(variables)
	if err != nil {
		return Buttprint{}, err
	}

	svg, err := s.renderer.Render(score)
	if err != nil {
		return Buttprint{}, err
	}

	return Buttprint{
		Variables: variables,
		Score:     score,
		SVG:       svg,
	}, nil
}
