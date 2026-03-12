package jackfruit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

// mustJSON marshals v to a JSON string, failing the test on error.
func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustJSON: %v", err)
	}
	return string(b)
}

// newTestClient creates an httptest server with handler and returns a Client pointing at it.
// The server is closed via t.Cleanup when the test ends.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(http.DefaultClient, srv.URL)
}

func TestGetEnvironmentalData(t *testing.T) {
	refTime := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)
	fileID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Pre-build JSON response bodies so mustJSON failures surface at the parent test level.
	singleVarBody := mustJSON(t, environmentalResponse{
		Variables: []variableResponse{{
			Name: "pm2p5", Value: 12.5, Unit: "µg/m³",
			RefTimestamp: refTime, ActualLat: 52.5, ActualLon: 13.4,
		}},
	})
	multiVarBody := mustJSON(t, environmentalResponse{
		Variables: []variableResponse{
			{Name: "pm2p5", Value: 12.5, Unit: "µg/m³", RefTimestamp: refTime, ActualLat: 52.5, ActualLon: 13.4},
			{Name: "pm10", Value: 20.0, Unit: "µg/m³", RefTimestamp: refTime, ActualLat: 52.5, ActualLon: 13.4},
			{Name: "temperature", Value: 295.15, Unit: "K", RefTimestamp: refTime, ActualLat: 52.5, ActualLon: 13.4},
			{Name: "humidity", Value: 75.0, Unit: "%", RefTimestamp: refTime, ActualLat: 52.5, ActualLon: 13.4},
		},
	})
	withLineageBody := mustJSON(t, environmentalResponse{
		Variables: []variableResponse{{
			Name: "pm2p5", Value: 12.5, Unit: "µg/m³",
			RefTimestamp: refTime, ActualLat: 52.5, ActualLon: 13.4,
			Lineage: &lineageResponse{Source: "cams", Dataset: "global", RawFileID: fileID},
		}},
	})
	withoutLineageBody := mustJSON(t, environmentalResponse{
		Variables: []variableResponse{{
			Name: "pm2p5", Value: 12.5, Unit: "µg/m³",
			RefTimestamp: refTime, ActualLat: 52.5, ActualLon: 13.4,
			Lineage: nil,
		}},
	})

	jsonResp := func(status int, body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}
	}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		// client overrides newTestClient when set (e.g. for the network error case).
		client func(t *testing.T) *Client
		// ctx overrides context.Background() when set.
		ctx   func(t *testing.T) context.Context
		check func(t *testing.T, vars []domain.VariableData, err error)
	}{
		{
			name:    "success single variable",
			handler: jsonResp(http.StatusOK, singleVarBody),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(vars) != 1 {
					t.Fatalf("want 1 variable, got %d", len(vars))
				}
				v := vars[0]
				if v.Name != "pm2p5" {
					t.Errorf("name: got %q, want pm2p5", v.Name)
				}
				if v.Value != 12.5 {
					t.Errorf("value: got %v, want 12.5", v.Value)
				}
				if v.Unit != "µg/m³" {
					t.Errorf("unit: got %q, want µg/m³", v.Unit)
				}
				if !v.RefTimestamp.Equal(refTime) {
					t.Errorf("ref_timestamp: got %v, want %v", v.RefTimestamp, refTime)
				}
				if v.ActualLat != 52.5 {
					t.Errorf("actual_lat: got %v, want 52.5", v.ActualLat)
				}
				if v.ActualLon != 13.4 {
					t.Errorf("actual_lon: got %v, want 13.4", v.ActualLon)
				}
				if v.Lineage != nil {
					t.Error("lineage: want nil")
				}
			},
		},
		{
			name:    "success multiple variables",
			handler: jsonResp(http.StatusOK, multiVarBody),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(vars) != 4 {
					t.Fatalf("want 4 variables, got %d", len(vars))
				}
				wantNames := []string{"pm2p5", "pm10", "temperature", "humidity"}
				for i, name := range wantNames {
					if vars[i].Name != name {
						t.Errorf("vars[%d].Name: got %q, want %q", i, vars[i].Name, name)
					}
				}
			},
		},
		{
			name:    "success with lineage",
			handler: jsonResp(http.StatusOK, withLineageBody),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(vars) != 1 {
					t.Fatalf("want 1 variable, got %d", len(vars))
				}
				l := vars[0].Lineage
				if l == nil {
					t.Fatal("lineage: want non-nil")
				}
				if l.Source != "cams" {
					t.Errorf("source: got %q, want cams", l.Source)
				}
				if l.Dataset != "global" {
					t.Errorf("dataset: got %q, want global", l.Dataset)
				}
				if l.RawFileID != fileID {
					t.Errorf("raw_file_id: got %v, want %v", l.RawFileID, fileID)
				}
			},
		},
		{
			name:    "success without lineage (null)",
			handler: jsonResp(http.StatusOK, withoutLineageBody),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(vars) == 0 {
					t.Fatal("want 1 variable, got 0")
				}
				if vars[0].Lineage != nil {
					t.Error("lineage: want nil")
				}
			},
		},
		{
			name:    "jackfruit 400",
			handler: jsonResp(http.StatusBadRequest, `{"error":"bad request"}`),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				var upErr domain.ErrUpstream
				if !errors.As(err, &upErr) {
					t.Fatalf("want ErrUpstream, got %T: %v", err, err)
				}
				if upErr.Service != "jackfruit" {
					t.Errorf("service: got %q, want jackfruit", upErr.Service)
				}
			},
		},
		{
			name:    "jackfruit 404",
			handler: jsonResp(http.StatusNotFound, `{"error":"not found"}`),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				var noData domain.ErrNoData
				if !errors.As(err, &noData) {
					t.Fatalf("want ErrNoData, got %T: %v", err, err)
				}
				if noData.Lat != 52.52 {
					t.Errorf("lat: got %v, want 52.52", noData.Lat)
				}
				if noData.Lon != 13.405 {
					t.Errorf("lon: got %v, want 13.405", noData.Lon)
				}
				if !noData.Timestamp.Equal(refTime) {
					t.Errorf("timestamp: got %v, want %v", noData.Timestamp, refTime)
				}
			},
		},
		{
			name:    "jackfruit 500",
			handler: jsonResp(http.StatusInternalServerError, `{"error":"internal error"}`),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				var upErr domain.ErrUpstream
				if !errors.As(err, &upErr) {
					t.Fatalf("want ErrUpstream, got %T: %v", err, err)
				}
			},
		},
		{
			name:    "invalid JSON",
			handler: jsonResp(http.StatusOK, `{not valid json`),
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				var upErr domain.ErrUpstream
				if !errors.As(err, &upErr) {
					t.Fatalf("want ErrUpstream, got %T: %v", err, err)
				}
			},
		},
		{
			// Verifies fix: Do errors must wrap in ErrUpstream (not bare fmt.Errorf → 500).
			name: "network error",
			client: func(t *testing.T) *Client {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				srv.Close() // closed before any request lands
				return NewClient(http.DefaultClient, srv.URL)
			},
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				var upErr domain.ErrUpstream
				if !errors.As(err, &upErr) {
					t.Fatalf("want ErrUpstream, got %T: %v", err, err)
				}
			},
		},
		{
			name: "request timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(500 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"variables":[]}`))
			},
			ctx: func(t *testing.T) context.Context {
				// 100 ms deadline — well inside the handler's 500 ms sleep.
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("want DeadlineExceeded, got %v", err)
				}
			},
		},
		{
			name:    "context cancellation",
			handler: jsonResp(http.StatusOK, `{"variables":[]}`),
			ctx: func(t *testing.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // cancelled before the call
				return ctx
			},
			check: func(t *testing.T, vars []domain.VariableData, err error) {
				if !errors.Is(err, context.Canceled) {
					t.Errorf("want Canceled, got %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var client *Client
			if tc.client != nil {
				client = tc.client(t)
			} else {
				client = newTestClient(t, tc.handler)
			}

			ctx := context.Background()
			if tc.ctx != nil {
				ctx = tc.ctx(t)
			}

			vars, err := client.GetEnvironmentalData(ctx, 52.52, 13.405, refTime, []string{"pm2p5", "pm10"})
			tc.check(t, vars, err)
		})
	}
}

// TestGetEnvironmentalData_URLConstruction verifies the client constructs the correct
// path and query parameters when calling the Jackfruit API.
func TestGetEnvironmentalData_URLConstruction(t *testing.T) {
	refTime := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)

	var gotReq *http.Request
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"variables":[]}`))
	})

	_, _ = client.GetEnvironmentalData(
		context.Background(),
		52.52, 13.405,
		refTime,
		[]string{"pm2p5", "pm10", "temperature", "humidity"},
	)

	if gotReq == nil {
		t.Fatal("handler was not called")
	}
	if gotReq.URL.Path != "/v1/environmental" {
		t.Errorf("path: got %q, want /v1/environmental", gotReq.URL.Path)
	}

	q := gotReq.URL.Query()
	if got := q.Get("lat"); got != "52.52" {
		t.Errorf("lat: got %q, want 52.52", got)
	}
	if got := q.Get("lon"); got != "13.405" {
		t.Errorf("lon: got %q, want 13.405", got)
	}
	if got, want := q.Get("timestamp"), refTime.Format(time.RFC3339); got != want {
		t.Errorf("timestamp: got %q, want %q", got, want)
	}
	if got := q.Get("variables"); got != "pm2p5,pm10,temperature,humidity" {
		t.Errorf("variables: got %q, want pm2p5,pm10,temperature,humidity", got)
	}
}

func TestGetEnvironmentalData_URLConstruction_EdgeCases(t *testing.T) {
	refTime := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		lat     float64
		lon     float64
		wantLat string
		wantLon string
	}{
		{
			name:    "negative coordinates",
			lat:     -33.8688,
			lon:     151.2093,
			wantLat: "-33.8688",
			wantLon: "151.2093",
		},
		{
			name:    "zero coordinates",
			lat:     0,
			lon:     0,
			wantLat: "0",
			wantLon: "0",
		},
		{
			name:    "high-precision coordinates",
			lat:     52.520008,
			lon:     13.404954,
			wantLat: "52.520008",
			wantLon: "13.404954",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotReq *http.Request
			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				gotReq = r
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"variables":[]}`))
			})

			_, _ = client.GetEnvironmentalData(
				context.Background(),
				tc.lat, tc.lon,
				refTime,
				[]string{"pm2p5"},
			)

			if gotReq == nil {
				t.Fatal("handler was not called")
			}

			q := gotReq.URL.Query()
			if got := q.Get("lat"); got != tc.wantLat {
				t.Errorf("lat: got %q, want %q", got, tc.wantLat)
			}
			if got := q.Get("lon"); got != tc.wantLon {
				t.Errorf("lon: got %q, want %q", got, tc.wantLon)
			}
		})
	}
}
