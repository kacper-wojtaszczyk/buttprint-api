package scoring

import (
	"strings"
	"testing"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

func allVars(temp, hum, dew, pm25, pm10 float64) []domain.VariableData {
	return []domain.VariableData{
		{Name: "temperature", Value: temp},
		{Name: "humidity", Value: hum},
		{Name: "dewpoint", Value: dew},
		{Name: "pm2p5", Value: pm25},
		{Name: "pm10", Value: pm10},
	}
}

// TestCalculate_MidRange uses exact midpoints of each clamp range so all
// normalised values are 0.5, making expected outputs trivially verifiable.
// temp=9: (9-(-30))/(48-(-30)) = 39/78 = 0.5
// hum=54: (54-10)/(98-10) = 44/88 = 0.5
// dew=6:  (6-(-20))/(32-(-20)) = 26/52 = 0.5
// pm25=150: 150/300 = 0.5, pm10=250: 250/500 = 0.5
func TestCalculate_MidRange(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(9, 54, 6, 150, 250))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Thickness != 0.5 {
		t.Errorf("Thickness: expected 0.5, got %v", score.Thickness)
	}
	if score.Sweatiness != 0.5 {
		t.Errorf("Sweatiness: expected 0.5, got %v", score.Sweatiness)
	}
	if score.Irritation != 0.5 {
		t.Errorf("Irritation: expected 0.5, got %v", score.Irritation)
	}
	if score.Warmth != 0.5 {
		t.Errorf("Warmth: expected 0.5, got %v", score.Warmth)
	}
}

func TestCalculate_AllMinimum(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(-30, 10, -20, 0, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Thickness != 0.0 {
		t.Errorf("Thickness: expected 0.0, got %v", score.Thickness)
	}
	if score.Sweatiness != 0.0 {
		t.Errorf("Sweatiness: expected 0.0, got %v", score.Sweatiness)
	}
	if score.Irritation != 0.0 {
		t.Errorf("Irritation: expected 0.0, got %v", score.Irritation)
	}
	if score.Warmth != 0.0 {
		t.Errorf("Warmth: expected 0.0, got %v", score.Warmth)
	}
}

func TestCalculate_AllMaximum(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(48, 98, 32, 300, 500))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Thickness != 1.0 {
		t.Errorf("Thickness: expected 1.0, got %v", score.Thickness)
	}
	if score.Sweatiness != 1.0 {
		t.Errorf("Sweatiness: expected 1.0, got %v", score.Sweatiness)
	}
	if score.Irritation != 1.0 {
		t.Errorf("Irritation: expected 1.0, got %v", score.Irritation)
	}
	if score.Warmth != 1.0 {
		t.Errorf("Warmth: expected 1.0, got %v", score.Warmth)
	}
}

func TestCalculate_ClampLow(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(-100, 0, -100, -50, -100))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Thickness != 0.0 {
		t.Errorf("Thickness: expected 0.0, got %v", score.Thickness)
	}
	if score.Sweatiness != 0.0 {
		t.Errorf("Sweatiness: expected 0.0, got %v", score.Sweatiness)
	}
	if score.Irritation != 0.0 {
		t.Errorf("Irritation: expected 0.0, got %v", score.Irritation)
	}
	if score.Warmth != 0.0 {
		t.Errorf("Warmth: expected 0.0, got %v", score.Warmth)
	}
}

func TestCalculate_ClampHigh(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(100, 200, 100, 1000, 1000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Thickness != 1.0 {
		t.Errorf("Thickness: expected 1.0, got %v", score.Thickness)
	}
	if score.Sweatiness != 1.0 {
		t.Errorf("Sweatiness: expected 1.0, got %v", score.Sweatiness)
	}
	if score.Irritation != 1.0 {
		t.Errorf("Irritation: expected 1.0, got %v", score.Irritation)
	}
	if score.Warmth != 1.0 {
		t.Errorf("Warmth: expected 1.0, got %v", score.Warmth)
	}
}

func TestCalculate_EmptyInput(t *testing.T) {
	s := NewScorer()
	_, err := s.Calculate(nil)
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	if !strings.Contains(err.Error(), "missing required variable") {
		t.Errorf("expected 'missing required variable' in error, got %q", err.Error())
	}
}

func TestCalculate_PartialInput(t *testing.T) {
	s := NewScorer()
	vars := []domain.VariableData{{Name: "pm2p5", Value: 100}}
	_, err := s.Calculate(vars)
	if err == nil {
		t.Fatal("expected error for partial input, got nil")
	}
	if !strings.Contains(err.Error(), "missing required variable") {
		t.Errorf("expected 'missing required variable' in error, got %q", err.Error())
	}
}

func TestCalculate_MissingVariables(t *testing.T) {
	tests := []struct {
		omit string
		vars []domain.VariableData
	}{
		{
			omit: "temperature",
			vars: []domain.VariableData{
				{Name: "humidity", Value: 54},
				{Name: "dewpoint", Value: 6},
				{Name: "pm2p5", Value: 150},
				{Name: "pm10", Value: 250},
			},
		},
		{
			omit: "humidity",
			vars: []domain.VariableData{
				{Name: "temperature", Value: 9},
				{Name: "dewpoint", Value: 6},
				{Name: "pm2p5", Value: 150},
				{Name: "pm10", Value: 250},
			},
		},
		{
			omit: "dewpoint",
			vars: []domain.VariableData{
				{Name: "temperature", Value: 9},
				{Name: "humidity", Value: 54},
				{Name: "pm2p5", Value: 150},
				{Name: "pm10", Value: 250},
			},
		},
		{
			omit: "pm2p5",
			vars: []domain.VariableData{
				{Name: "temperature", Value: 9},
				{Name: "humidity", Value: 54},
				{Name: "dewpoint", Value: 6},
				{Name: "pm10", Value: 250},
			},
		},
		{
			omit: "pm10",
			vars: []domain.VariableData{
				{Name: "temperature", Value: 9},
				{Name: "humidity", Value: 54},
				{Name: "dewpoint", Value: 6},
				{Name: "pm2p5", Value: 150},
			},
		},
	}

	for _, tt := range tests {
		t.Run("missing_"+tt.omit, func(t *testing.T) {
			s := NewScorer()
			_, err := s.Calculate(tt.vars)
			if err == nil {
				t.Fatalf("expected error when %q is missing, got nil", tt.omit)
			}
			if !strings.Contains(err.Error(), tt.omit) {
				t.Errorf("expected error to mention %q, got %q", tt.omit, err.Error())
			}
		})
	}
}
