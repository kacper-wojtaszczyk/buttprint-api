package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseButtprintRequest(t *testing.T) {
	tests := []struct {
		name             string
		query            string
		wantErr          bool
		wantCoordsNil    bool
		wantTimestampNil bool
		wantLat          float64
		wantLon          float64
	}{
		{
			name:    "valid full request",
			query:   "lat=52.52&lon=13.40&timestamp=2026-03-08T14:00:00Z",
			wantLat: 52.52,
			wantLon: 13.40,
		},
		{
			name:             "valid coords only",
			query:            "lat=52.52&lon=13.40",
			wantLat:          52.52,
			wantLon:          13.40,
			wantTimestampNil: true,
		},
		{
			name:             "no params",
			query:            "",
			wantCoordsNil:    true,
			wantTimestampNil: true,
		},
		{
			name:    "lat without lon",
			query:   "lat=52.52",
			wantErr: true,
		},
		{
			name:    "lon without lat",
			query:   "lon=13.40",
			wantErr: true,
		},
		{
			name:    "non-numeric lat",
			query:   "lat=abc&lon=13.40",
			wantErr: true,
		},
		{
			name:    "lat out of range high",
			query:   "lat=91&lon=13.40",
			wantErr: true,
		},
		{
			name:    "lat out of range low",
			query:   "lat=-91&lon=13.40",
			wantErr: true,
		},
		{
			name:    "lon out of range high",
			query:   "lat=52.52&lon=181",
			wantErr: true,
		},
		{
			name:    "lon out of range low",
			query:   "lat=52.52&lon=-181",
			wantErr: true,
		},
		{
			name:    "lat boundary max",
			query:   "lat=90&lon=13.40",
			wantLat: 90,
			wantLon: 13.40,
		},
		{
			name:    "lat boundary min",
			query:   "lat=-90&lon=13.40",
			wantLat: -90,
			wantLon: 13.40,
		},
		{
			name:    "lon boundary max",
			query:   "lat=52.52&lon=180",
			wantLat: 52.52,
			wantLon: 180,
		},
		{
			name:    "lon boundary min",
			query:   "lat=52.52&lon=-180",
			wantLat: 52.52,
			wantLon: -180,
		},
		{
			name:    "invalid timestamp format",
			query:   "lat=52.52&lon=13.40&timestamp=2026-03-08",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			br, err := parseButtprintRequest(req)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantCoordsNil {
				if br.Coords != nil {
					t.Errorf("expected nil coords, got %+v", br.Coords)
				}
			} else {
				if br.Coords == nil {
					t.Fatal("expected coords, got nil")
				}
				if br.Coords.Lat != tt.wantLat {
					t.Errorf("expected lat %v, got %v", tt.wantLat, br.Coords.Lat)
				}
				if br.Coords.Lon != tt.wantLon {
					t.Errorf("expected lon %v, got %v", tt.wantLon, br.Coords.Lon)
				}
			}

			if tt.wantTimestampNil && br.Timestamp != nil {
				t.Errorf("expected nil timestamp, got %v", br.Timestamp)
			}
		})
	}
}

func TestParseButtprintRequest_TimestampParsed(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		expected  time.Time
	}{
		{
			name:      "UTC with Z",
			timestamp: "2026-03-08T14:00:00Z",
			expected:  time.Date(2026, 3, 8, 14, 0, 0, 0, time.UTC),
		},
		{
			name:      "positive timezone offset",
			timestamp: "2026-03-08T14:00:00%2B02:00", // + must be URL-encoded in query strings
			expected:  time.Date(2026, 3, 8, 14, 0, 0, 0, time.FixedZone("", 2*60*60)),
		},
		{
			name:      "negative timezone offset",
			timestamp: "2026-03-08T14:00:00-05:00",
			expected:  time.Date(2026, 3, 8, 14, 0, 0, 0, time.FixedZone("", -5*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?lat=52.52&lon=13.40&timestamp="+tt.timestamp, nil)
			br, err := parseButtprintRequest(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if br.Timestamp == nil {
				t.Fatal("expected timestamp, got nil")
			}
			if !br.Timestamp.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, *br.Timestamp)
			}
		})
	}
}
