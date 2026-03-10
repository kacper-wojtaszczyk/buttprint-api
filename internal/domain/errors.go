package domain

import (
	"fmt"
	"time"
)

// ErrNoData is returned when no environmental data exists for the given location and time.
// The handler maps this to HTTP 404.
type ErrNoData struct {
	Lat       float64
	Lon       float64
	Timestamp time.Time
}

func (e ErrNoData) Error() string {
	return fmt.Sprintf("no environmental data for lat=%.4f, lon=%.4f, timestamp=%s", e.Lat, e.Lon, e.Timestamp.Format(time.RFC3339))
}

// ErrUpstream is returned when an external dependency (Jackfruit, renderer, etc.) fails.
// The handler maps this to HTTP 502.
type ErrUpstream struct {
	Service string
	Cause   error
}

func (e ErrUpstream) Error() string {
	return fmt.Sprintf("upstream service %q failed: %v", e.Service, e.Cause)
}

// Unwrap allows errors.Is and errors.As to inspect the underlying cause.
func (e ErrUpstream) Unwrap() error {
	return e.Cause
}
