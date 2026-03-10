package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

type buttprintProvider interface {
	GetButtprint(ctx context.Context, lat, lon float64, timestamp time.Time) (domain.Buttprint, error)
}

type Handler struct {
	buttprintProvider buttprintProvider
	logger            *slog.Logger
}

func NewHandler(provider buttprintProvider, logger *slog.Logger) *Handler {
	return &Handler{
		buttprintProvider: provider,
		logger:            logger,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /buttprint", h.handleButtprint)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleButtprint(w http.ResponseWriter, r *http.Request) {
	br, err := ParseButtprintRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var coords Coords
	if br.Coords != nil {
		coords = *br.Coords
	} else {
		writeError(w, http.StatusBadRequest, "coords are required (for now)")
		return
	}

	var timestamp time.Time
	if br.Timestamp != nil {
		timestamp = *br.Timestamp
	} else {
		timestamp = time.Now()
	}

	buttprint, err := h.buttprintProvider.GetButtprint(r.Context(), coords.Lat, coords.Lon, timestamp)
	if err != nil {
		h.logger.Error("retrieving Buttprint failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}

	response := ButtprintResponse{
		Location: LocationResponse{
			Lat:    coords.Lat,
			Lon:    coords.Lon,
			Source: "explicit",
		},
		RequestedTimestamp: timestamp,
		Variables:          make([]VariableResponse, len(buttprint.Variables)),
		Score: ScoreResponse{
			Composite:  buttprint.Score.Composite,
			Thickness:  buttprint.Score.Thickness,
			Sweatiness: buttprint.Score.Sweatiness,
			Irritation: buttprint.Score.Sweatiness,
		},
		SVG: buttprint.SVG,
	}
	for i, variable := range buttprint.Variables {
		var lineageResponse *LineageResponse
		if variable.Lineage != nil {
			lineageResponse = &LineageResponse{
				Source:    variable.Lineage.Source,
				Dataset:   variable.Lineage.Dataset,
				RawFileID: variable.Lineage.RawFileID,
			}
		}
		response.Variables[i] = VariableResponse{
			Name:         variable.Name,
			Value:        variable.Value,
			Unit:         variable.Unit,
			RefTimestamp: variable.RefTimestamp,
			ActualLat:    variable.ActualLat,
			ActualLon:    variable.ActualLon,
			Lineage:      lineageResponse,
		}
	}

	writeJSON(w, http.StatusOK, response)
}
