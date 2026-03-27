package geoloc

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/oschwald/geoip2-golang"
)

type mockGeoIPReader struct {
	record *geoip2.City
	err    error
}

func (m *mockGeoIPReader) City(_ net.IP) (*geoip2.City, error) {
	return m.record, m.err
}

func (m *mockGeoIPReader) Close() error { return nil }

func TestResolve_AccuracyRadius(t *testing.T) {
	tests := []struct {
		name           string
		accuracyRadius uint16
		lat, lon       float64
		wantErr        error
	}{
		{
			name:           "valid location with accuracy radius",
			accuracyRadius: 100,
			lat:            52.52,
			lon:            13.40,
		},
		{
			name:           "zero accuracy radius returns ErrLookupFailed",
			accuracyRadius: 0,
			lat:            0,
			lon:            0,
			wantErr:        ErrLookupFailed,
		},
		{
			name:           "zero accuracy radius with non-zero coords still fails",
			accuracyRadius: 0,
			lat:            52.52,
			lon:            13.40,
			wantErr:        ErrLookupFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &geoip2.City{}
			record.Location.AccuracyRadius = tt.accuracyRadius
			record.Location.Latitude = tt.lat
			record.Location.Longitude = tt.lon

			resolver := &MaxMindResolver{reader: &mockGeoIPReader{record: record}}

			lat, lon, err := resolver.Resolve("203.0.113.50")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if lat != tt.lat || lon != tt.lon {
				t.Errorf("expected (%v, %v), got (%v, %v)", tt.lat, tt.lon, lat, lon)
			}
		})
	}
}

func TestMaxMindResolver_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dbPath := os.Getenv("MAXMIND_DB_PATH")
	if dbPath == "" {
		t.Skip("MAXMIND_DB_PATH not set")
	}

	resolver, err := NewMaxMindResolver(dbPath)
	if err != nil {
		t.Fatalf("NewMaxMindResolver: %v", err)
	}
	defer resolver.Close()

	t.Run("public IP returns location", func(t *testing.T) {
		lat, lon, err := resolver.Resolve("8.8.8.8")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if lat == 0 && lon == 0 {
			t.Error("expected non-zero coordinates for 8.8.8.8")
		}
	})

	t.Run("private IP returns ErrPrivateIP", func(t *testing.T) {
		_, _, err := resolver.Resolve("192.168.1.1")
		if !errors.Is(err, ErrPrivateIP) {
			t.Errorf("expected ErrPrivateIP, got: %v", err)
		}
	})

	t.Run("loopback returns ErrPrivateIP", func(t *testing.T) {
		_, _, err := resolver.Resolve("127.0.0.1")
		if !errors.Is(err, ErrPrivateIP) {
			t.Errorf("expected ErrPrivateIP, got: %v", err)
		}
	})

	t.Run("IPv6 loopback returns ErrPrivateIP", func(t *testing.T) {
		_, _, err := resolver.Resolve("::1")
		if !errors.Is(err, ErrPrivateIP) {
			t.Errorf("expected ErrPrivateIP, got: %v", err)
		}
	})

	t.Run("unspecified IP returns ErrPrivateIP", func(t *testing.T) {
		_, _, err := resolver.Resolve("0.0.0.0")
		if !errors.Is(err, ErrPrivateIP) {
			t.Errorf("expected ErrPrivateIP, got: %v", err)
		}
	})

	t.Run("malformed IP returns error", func(t *testing.T) {
		_, _, err := resolver.Resolve("not-an-ip")
		if err == nil {
			t.Fatal("expected error for malformed IP")
		}
		if errors.Is(err, ErrPrivateIP) {
			t.Error("malformed IP should not be ErrPrivateIP")
		}
	})
}
