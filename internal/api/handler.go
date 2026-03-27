package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/geoloc"
)

type buttprintProvider interface {
	GetButtprint(ctx context.Context, lat, lon float64, timestamp time.Time) (domain.Buttprint, error)
}

type ipResolver interface {
	Resolve(ip string) (lat, lon float64, err error)
}

type Handler struct {
	buttprintProvider buttprintProvider
	ipResolver        ipResolver
	logger            *slog.Logger
}

func NewHandler(provider buttprintProvider, ipResolver ipResolver, logger *slog.Logger) *Handler {
	return &Handler{
		buttprintProvider: provider,
		ipResolver:        ipResolver,
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
	br, err := parseButtprintRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var c coords
	if br.Coords != nil {
		c = *br.Coords
	} else {
		ip := clientIP(r)
		lat, lon, err := h.ipResolver.Resolve(ip)
		if err != nil {
			if errors.Is(err, geoloc.ErrPrivateIP) {
				h.logger.Info("cannot geolocate private IP", "ip", ip)
				writeError(w, http.StatusBadRequest, err.Error())
			} else if errors.Is(err, geoloc.ErrLookupFailed) {
				h.logger.Warn("ip geolocation lookup failed", "ip", ip, "err", err)
				writeError(w, http.StatusBadRequest, "geolocation lookup failed")
			} else {
				h.logger.Error("unexpected ip resolution error", "ip", ip, "err", err)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		c = coords{Lat: lat, Lon: lon}
	}

	var ts time.Time
	if br.Timestamp != nil {
		ts = *br.Timestamp
	} else {
		ts = time.Now()
	}

	buttprint, err := h.buttprintProvider.GetButtprint(r.Context(), c.Lat, c.Lon, ts)
	if err != nil {
		if _, ok := errors.AsType[domain.ErrNoData](err); ok {
			writeError(w, http.StatusNotFound, "no data available for this location and time")
			return
		}
		h.logger.Error("retrieving Buttprint failed", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusGatewayTimeout, "request timed out")
		} else if _, ok := errors.AsType[domain.ErrUpstream](err); ok {
			writeError(w, http.StatusBadGateway, "upstream service error")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, newButtprintResponse(buttprint, c, ts))
}
