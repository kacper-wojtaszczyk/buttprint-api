package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ButtprintResponse struct {
	Location           LocationResponse   `json:"location"`
	RequestedTimestamp time.Time          `json:"requested_timestamp"`
	Variables          []VariableResponse `json:"variables"`
	Score              ScoreResponse      `json:"score"`
	SVG                string             `json:"svg"`
}

type LocationResponse struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Name   string  `json:"name,omitempty"`
	Source string  `json:"source"`
}

type VariableResponse struct {
	Name         string           `json:"name"`
	Value        float64          `json:"value"`
	Unit         string           `json:"unit"`
	RefTimestamp time.Time        `json:"ref_timestamp"`
	ActualLat    float64          `json:"actual_lat"`
	ActualLon    float64          `json:"actual_lon"`
	Lineage      *LineageResponse `json:"lineage"`
}

type ScoreResponse struct {
	Thiccness  float64 `json:"thiccness"`
	Sweatiness float64 `json:"sweatiness"`
	Irritation float64 `json:"irritation"`
	Warmth     float64 `json:"warmth"`
}

type LineageResponse struct {
	Source    string    `json:"source"`
	Dataset   string    `json:"dataset"`
	RawFileID uuid.UUID `json:"raw_file_id"`
}

func newButtprintResponse(buttprint domain.Buttprint, coords coords, timestamp time.Time) ButtprintResponse {
	variables := make([]VariableResponse, len(buttprint.Variables))
	for i, v := range buttprint.Variables {
		variables[i] = newVariableResponse(v)
	}
	return ButtprintResponse{
		Location: LocationResponse{
			Lat:    coords.Lat,
			Lon:    coords.Lon,
			Source: "explicit",
		},
		RequestedTimestamp: timestamp,
		Variables:          variables,
		Score: ScoreResponse{
			Thiccness:  buttprint.Score.Thiccness,
			Sweatiness: buttprint.Score.Sweatiness,
			Irritation: buttprint.Score.Irritation,
			Warmth:     buttprint.Score.Warmth,
		},
		SVG: buttprint.SVG,
	}
}

func newVariableResponse(v domain.VariableData) VariableResponse {
	var lineage *LineageResponse
	if v.Lineage != nil {
		lineage = &LineageResponse{
			Source:    v.Lineage.Source,
			Dataset:   v.Lineage.Dataset,
			RawFileID: v.Lineage.RawFileID,
		}
	}
	return VariableResponse{
		Name:         v.Name,
		Value:        v.Value,
		Unit:         v.Unit,
		RefTimestamp: v.RefTimestamp,
		ActualLat:    v.ActualLat,
		ActualLon:    v.ActualLon,
		Lineage:      lineage,
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
