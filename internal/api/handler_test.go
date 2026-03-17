package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

// mockButtprintProvider implements the unexported buttprintProvider interface.
type mockButtprintProvider struct {
	result domain.Buttprint
	err    error
}

func (m *mockButtprintProvider) GetButtprint(_ context.Context, _, _ float64, _ time.Time) (domain.Buttprint, error) {
	return m.result, m.err
}

func newTestHandler(provider buttprintProvider) *Handler {
	return NewHandler(provider, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func stubButtprint() domain.Buttprint {
	return domain.Buttprint{
		Variables: []domain.VariableData{
			{Name: "temperature", Value: 25, Unit: "°C"},
		},
		Score: domain.Score{
			Thickness:  0.5,
			Warmth:     0.4,
			Sweatiness: 0.6,
			Irritation: 0.3,
		},
		SVG: "<svg/>",
	}
}

func TestHealthHandler(t *testing.T) {
	mux := http.NewServeMux()
	newTestHandler(nil).RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if body := w.Body.String(); body != "" {
		t.Errorf("expected empty body, got %q", body)
	}
}

func TestHandleButtprint_StatusCodes(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		providerErr error
		wantStatus  int
	}{
		{
			name:       "happy path",
			url:        "/buttprint?lat=52.52&lon=13.40&timestamp=2026-03-08T14:00:00Z",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing coords",
			url:        "/buttprint",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "partial coords lat only",
			url:        "/buttprint?lat=52.52",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "partial coords lon only",
			url:        "/buttprint?lon=13.40",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid lat range",
			url:        "/buttprint?lat=91&lon=13.40",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid lon range",
			url:        "/buttprint?lat=52.52&lon=200",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid timestamp format",
			url:        "/buttprint?lat=52.52&lon=13.40&timestamp=2026-03-08",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no timestamp defaults to now",
			url:        "/buttprint?lat=52.52&lon=13.40",
			wantStatus: http.StatusOK,
		},
		{
			name:        "ErrNoData from provider",
			url:         "/buttprint?lat=52.52&lon=13.40",
			providerErr: domain.ErrNoData{Lat: 52.52, Lon: 13.40},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "ErrUpstream from provider",
			url:         "/buttprint?lat=52.52&lon=13.40",
			providerErr: domain.ErrUpstream{Service: "jackfruit", Cause: errors.New("connection refused")},
			wantStatus:  http.StatusBadGateway,
		},
		{
			name:        "context.DeadlineExceeded from provider",
			url:         "/buttprint?lat=52.52&lon=13.40",
			providerErr: context.DeadlineExceeded,
			wantStatus:  http.StatusGatewayTimeout,
		},
		{
			name:        "generic error from provider",
			url:         "/buttprint?lat=52.52&lon=13.40",
			providerErr: errors.New("something went wrong"),
			wantStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockButtprintProvider{
				result: stubButtprint(),
				err:    tt.providerErr,
			}
			h := newTestHandler(provider)
			mux := http.NewServeMux()
			h.RegisterRoutes(mux)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestHandleButtprint_ResponseShape(t *testing.T) {
	provider := &mockButtprintProvider{result: stubButtprint()}
	h := newTestHandler(provider)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/buttprint?lat=52.52&lon=13.40&timestamp=2026-03-08T14:00:00Z", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var resp ButtprintResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Location.Lat != 52.52 {
		t.Errorf("expected lat 52.52, got %v", resp.Location.Lat)
	}
	if resp.Location.Lon != 13.40 {
		t.Errorf("expected lon 13.40, got %v", resp.Location.Lon)
	}
	if resp.Location.Source != "explicit" {
		t.Errorf("expected source 'explicit', got %q", resp.Location.Source)
	}
	expectedTimestamp := time.Date(2026, 3, 8, 14, 0, 0, 0, time.UTC)
	if !resp.RequestedTimestamp.Equal(expectedTimestamp) {
		t.Errorf("expected requested_timestamp %v, got %v", expectedTimestamp, resp.RequestedTimestamp)
	}
	if resp.Score.Thickness != 0.5 {
		t.Errorf("expected thickness 0.5, got %f", resp.Score.Thickness)
	}
	if resp.Score.Warmth != 0.4 {
		t.Errorf("expected warmth 0.4, got %v", resp.Score.Warmth)
	}
	if resp.Score.Sweatiness != 0.6 {
		t.Errorf("expected sweatiness 0.6, got %v", resp.Score.Sweatiness)
	}
	if resp.Score.Irritation != 0.3 {
		t.Errorf("expected irritation 0.3, got %v", resp.Score.Irritation)
	}
	if resp.SVG != "<svg/>" {
		t.Errorf("expected SVG '<svg/>', got %q", resp.SVG)
	}
	if len(resp.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(resp.Variables))
	}
}

func TestHandleButtprint_ErrorResponseShape(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError string
	}{
		{
			name:      "missing coords",
			url:       "/buttprint",
			wantError: "coords are required (for now)",
		},
		{
			name:      "invalid lat",
			url:       "/buttprint?lat=abc&lon=13.40",
			wantError: "invalid lat: must be a number",
		},
		{
			name:      "invalid timestamp",
			url:       "/buttprint?lat=52.52&lon=13.40&timestamp=not-a-date",
			wantError: "invalid timestamp: must be RFC 3339 format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(&mockButtprintProvider{result: stubButtprint()})
			mux := http.NewServeMux()
			h.RegisterRoutes(mux)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("expected Content-Type application/json, got %q", ct)
			}

			var errResp ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}
			if errResp.Error == "" {
				t.Error("expected non-empty error message")
			}
			if !strings.Contains(errResp.Error, tt.wantError) {
				t.Errorf("expected error containing %q, got %q", tt.wantError, errResp.Error)
			}
		})
	}
}

func TestHandleButtprint_DefaultTimestamp(t *testing.T) {
	before := time.Now()

	provider := &mockButtprintProvider{result: stubButtprint()}
	h := newTestHandler(provider)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/buttprint?lat=52.52&lon=13.40", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	after := time.Now()

	var resp ButtprintResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.RequestedTimestamp.Before(before) || resp.RequestedTimestamp.After(after) {
		t.Errorf("expected timestamp between %v and %v, got %v", before, after, resp.RequestedTimestamp)
	}
}
