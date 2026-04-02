package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRecoveryMiddleware_PanicReturns500(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	})
	wrapped := RecoveryMiddleware(discardLogger())(panicHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}

	body := w.Body.String()
	var errResp ErrorResponse
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&errResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in body")
	}
	if strings.Contains(body, "kaboom") {
		t.Error("panic message leaked to client")
	}
}

func TestRecoveryMiddleware_PanicAfterWrite(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("partial"))
		panic("mid-write kaboom")
	})
	wrapped := RecoveryMiddleware(discardLogger())(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	// Headers were already sent — recovery must not attempt a second WriteHeader.
	// The original 200 status should be preserved, not overwritten to 500.
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (already sent before panic)", w.Code)
	}
	if !strings.Contains(w.Body.String(), "partial") {
		t.Error("expected partial body from before panic")
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	wrapped := RecoveryMiddleware(discardLogger())(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("body = %q, want \"ok\"", w.Body.String())
	}
}

func TestStatusRecorder_CapturesStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

	rec.WriteHeader(http.StatusNotFound)

	if rec.status != http.StatusNotFound {
		t.Errorf("recorder status = %d, want 404", rec.status)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("underlying status = %d, want 404", w.Code)
	}
}

func TestLoggingMiddleware_Transparent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})
	wrapped := LoggingMiddleware(discardLogger())(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if w.Body.String() != "created" {
		t.Errorf("body = %q, want \"created\"", w.Body.String())
	}
}
