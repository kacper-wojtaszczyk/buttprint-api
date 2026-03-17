package scoring

import (
	"math"
	"strings"
	"testing"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

const epsilon = 1e-9

func closeTo(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func assertScore(t *testing.T, got, want domain.Score) {
	t.Helper()
	if !closeTo(got.Thickness, want.Thickness) {
		t.Errorf("Thickness: want %v, got %v", want.Thickness, got.Thickness)
	}
	if !closeTo(got.Sweatiness, want.Sweatiness) {
		t.Errorf("Sweatiness: want %v, got %v", want.Sweatiness, got.Sweatiness)
	}
	if !closeTo(got.Irritation, want.Irritation) {
		t.Errorf("Irritation: want %v, got %v", want.Irritation, got.Irritation)
	}
	if !closeTo(got.Warmth, want.Warmth) {
		t.Errorf("Warmth: want %v, got %v", want.Warmth, got.Warmth)
	}
}

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
	assertScore(t, score, domain.Score{Thickness: 0.5, Sweatiness: 0.5, Irritation: 0.5, Warmth: 0.5})
}

func TestCalculate_AllMinimum(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(-30, 10, -20, 0, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScore(t, score, domain.Score{})
}

func TestCalculate_AllMaximum(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(48, 98, 32, 300, 500))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScore(t, score, domain.Score{Thickness: 1.0, Sweatiness: 1.0, Irritation: 1.0, Warmth: 1.0})
}

// TestCalculate_Weights verifies that each weight contributes correctly by
// setting one variable to its maximum and all others to their minimum.
// Only the maxed variable's weight should appear in the relevant score.
func TestCalculate_Weights(t *testing.T) {
	tests := []struct {
		name string
		vars []domain.VariableData
		want domain.Score
	}{
		{
			name: "only temperature at max",
			vars: allVars(48, 10, -20, 0, 0),
			want: domain.Score{Thickness: 0.30, Warmth: 1.0},
		},
		{
			name: "only humidity at max",
			vars: allVars(-30, 98, -20, 0, 0),
			want: domain.Score{Thickness: 0.30},
		},
		{
			name: "only pm2p5 at max",
			vars: allVars(-30, 10, -20, 300, 0),
			want: domain.Score{Thickness: 0.25, Irritation: 0.65},
		},
		{
			name: "only pm10 at max",
			vars: allVars(-30, 10, -20, 0, 500),
			want: domain.Score{Thickness: 0.15, Irritation: 0.35},
		},
		{
			name: "only dewpoint at max",
			vars: allVars(-30, 10, 32, 0, 0),
			want: domain.Score{Sweatiness: 1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScorer()
			score, err := s.Calculate(tt.vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertScore(t, score, tt.want)
		})
	}
}

func TestCalculate_ClampLow(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(-100, 0, -100, -50, -100))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScore(t, score, domain.Score{})
}

func TestCalculate_ClampHigh(t *testing.T) {
	s := NewScorer()
	score, err := s.Calculate(allVars(100, 200, 100, 1000, 1000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScore(t, score, domain.Score{Thickness: 1.0, Sweatiness: 1.0, Irritation: 1.0, Warmth: 1.0})
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

func TestCalculate_DuplicateVariable(t *testing.T) {
	s := NewScorer()
	vars := []domain.VariableData{
		{Name: "temperature", Value: 20},
		{Name: "temperature", Value: 30},
		{Name: "humidity", Value: 50},
		{Name: "dewpoint", Value: 10},
		{Name: "pm2p5", Value: 100},
		{Name: "pm10", Value: 200},
	}
	_, err := s.Calculate(vars)
	if err == nil {
		t.Fatal("expected error for duplicate variable, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate variable") {
		t.Errorf("expected 'duplicate variable' in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "temperature") {
		t.Errorf("expected error to name the duplicate variable, got %q", err.Error())
	}
}

func TestCalculate_ExtraVariablesIgnored(t *testing.T) {
	s := NewScorer()
	vars := append(allVars(9, 54, 6, 150, 250), domain.VariableData{Name: "wind_speed", Value: 42})
	score, err := s.Calculate(vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScore(t, score, domain.Score{Thickness: 0.5, Sweatiness: 0.5, Irritation: 0.5, Warmth: 0.5})
}
