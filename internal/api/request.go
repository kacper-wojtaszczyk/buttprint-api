package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type buttprintRequest struct {
	Coords    *coords
	Timestamp *time.Time
}

type coords struct {
	Lat float64
	Lon float64
}

func parseButtprintRequest(r *http.Request) (*buttprintRequest, error) {
	query := r.URL.Query()
	lat, err := parseFloat64(query.Get("lat"))
	if err != nil {
		return nil, fmt.Errorf("invalid lat: must be a number: %w", err)
	}
	lon, err := parseFloat64(query.Get("lon"))
	if err != nil {
		return nil, fmt.Errorf("invalid lon: must be a number: %w", err)
	}

	if lat != nil && (*lat < -90 || *lat > 90) {
		return nil, fmt.Errorf("lat must be between -90 and 90, got %v", *lat)
	}
	if lon != nil && (*lon < -180 || *lon > 180) {
		return nil, fmt.Errorf("lon must be between -180 and 180, got %v", *lon)
	}

	var c *coords

	if lat == nil && lon == nil {
		c = nil
	} else if lat != nil && lon != nil {
		c = &coords{Lat: *lat, Lon: *lon}
	} else {
		return nil, fmt.Errorf("lat and lon must be either both set or both empty, lat: %v, lon: %v", lat, lon)
	}

	timestamp, err := parseTime(query.Get("timestamp"))
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: must be RFC 3339 format: %w", err)
	}

	return &buttprintRequest{
		Coords:    c,
		Timestamp: timestamp,
	}, nil
}

func parseTime(timeString string) (*time.Time, error) {
	if timeString == "" {
		return nil, nil
	}
	timestamp, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return nil, err
	}

	return &timestamp, nil
}

func parseFloat64(floatString string) (*float64, error) {
	if floatString == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(floatString, 64)
	if err != nil {
		return nil, err
	}

	return &f, nil
}
