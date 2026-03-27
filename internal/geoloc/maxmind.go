package geoloc

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type geoipReader interface {
	City(ip net.IP) (*geoip2.City, error)
	Close() error
}

type MaxMindResolver struct {
	reader geoipReader
}

func NewMaxMindResolver(dbPath string) (*MaxMindResolver, error) {
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening GeoLite2 database: %w", err)
	}
	return &MaxMindResolver{reader: reader}, nil
}

func (r *MaxMindResolver) Close() error {
	return r.reader.Close()
}

func (r *MaxMindResolver) Resolve(ip string) (lat, lon float64, err error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return 0, 0, fmt.Errorf("malformed IP address: %s", ip)
	}

	if parsed.IsPrivate() || parsed.IsLoopback() || parsed.IsUnspecified() {
		return 0, 0, ErrPrivateIP
	}

	record, err := r.reader.City(parsed)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %v", ErrLookupFailed, err)
	}

	if record.Location.AccuracyRadius == 0 {
		return 0, 0, fmt.Errorf("%w: no location data for IP %s", ErrLookupFailed, ip)
	}

	return record.Location.Latitude, record.Location.Longitude, nil
}
