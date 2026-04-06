package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestLimiter(t *testing.T, rps float64, burst int) *RateLimiter {
	t.Helper()
	rl, err := NewRateLimiter(rps, burst, discardLogger())
	if err != nil {
		t.Fatalf("NewRateLimiter: %v", err)
	}
	return rl
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := newTestLimiter(t, 10, 10)
	defer rl.Stop()

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newTestLimiter(t, 1, 1)
	defer rl.Stop()

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request — consumes the single token
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "1.2.3.4:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want 200", w1.Code)
	}

	// Second request — bucket empty
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "1.2.3.4:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: status = %d, want 429", w2.Code)
	}
	if w2.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429 response")
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w2.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in 429 body")
	}
}

func TestRateLimiter_IndependentPerIP(t *testing.T) {
	rl := newTestLimiter(t, 1, 1)
	defer rl.Stop()

	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "1.1.1.1:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "2.2.2.2:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w1.Code != http.StatusOK {
		t.Errorf("IP 1: status = %d, want 200", w1.Code)
	}
	if w2.Code != http.StatusOK {
		t.Errorf("IP 2: status = %d, want 200", w2.Code)
	}
}

func TestRateLimiter_CleanupEvictsStale(t *testing.T) {
	rl := newTestLimiter(t, 1, 1)
	defer rl.Stop()

	rl.getLimiter("fresh.ip")
	rl.getLimiter("stale.ip")

	rl.mu.Lock()
	rl.clients["stale.ip"].lastSeen = time.Now().Add(-30 * time.Minute)
	rl.mu.Unlock()

	rl.cleanup()

	rl.mu.Lock()
	defer rl.mu.Unlock()
	if _, ok := rl.clients["stale.ip"]; ok {
		t.Error("stale client was not evicted")
	}
	if _, ok := rl.clients["fresh.ip"]; !ok {
		t.Error("fresh client was incorrectly evicted")
	}
}

func TestNewRateLimiter_RejectsInvalidConfig(t *testing.T) {
	cases := []struct {
		name  string
		rps   float64
		burst int
	}{
		{"zero rps", 0, 10},
		{"negative rps", -1, 10},
		{"zero burst", 10, 0},
		{"negative burst", 10, -5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rl, err := NewRateLimiter(tc.rps, tc.burst, discardLogger())
			if err == nil {
				rl.Stop()
				t.Fatalf("expected error for rps=%v burst=%d, got nil", tc.rps, tc.burst)
			}
			if rl != nil {
				t.Errorf("expected nil RateLimiter on error, got %+v", rl)
			}
		})
	}
}

func TestRateLimiter_StopClosesDoneChannel(t *testing.T) {
	rl := newTestLimiter(t, 1, 1)
	rl.Stop()

	select {
	case <-rl.done:
	default:
		t.Error("done channel was not closed by Stop()")
	}
}

func TestRateLimiter_StopIsIdempotent(t *testing.T) {
	rl := newTestLimiter(t, 1, 1)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("repeated Stop() panicked: %v", r)
		}
	}()

	rl.Stop()
	rl.Stop()
	rl.Stop()
}
