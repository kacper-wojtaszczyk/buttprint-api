package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	Composite  float64 `json:"composite"`
	Thickness  float64 `json:"thickness"`
	Sweatiness float64 `json:"sweatiness"`
	Irritation float64 `json:"irritation"`
}

type LineageResponse struct {
	Source    string    `json:"source"`
	Dataset   string    `json:"dataset"`
	RawFileID uuid.UUID `json:"raw_file_id"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
