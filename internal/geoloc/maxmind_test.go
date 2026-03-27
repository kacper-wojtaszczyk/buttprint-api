package geoloc

import (
	"errors"
	"os"
	"testing"
)

func TestMaxMindResolver_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dbPath := os.Getenv("GEOLITE2_DB_PATH")
	if dbPath == "" {
		t.Skip("GEOLITE2_DB_PATH not set")
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
