package api

import (
	"context"
	"errors"
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

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
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
		if _, ok := errors.AsType[domain.ErrNoData](err); ok {
			writeError(w, http.StatusNotFound, "no data available for this location and time")
		} else if _, ok := errors.AsType[domain.ErrUpstream](err); ok {
			writeError(w, http.StatusBadGateway, "upstream service error")
		} else if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusGatewayTimeout, "request timed out")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, newButtprintResponse(buttprint, coords, timestamp))
}
