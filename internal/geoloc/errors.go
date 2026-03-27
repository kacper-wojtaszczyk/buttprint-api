package geoloc

import "errors"

var (
	// ErrPrivateIP indicates the IP is private/loopback and cannot be geolocated.
	ErrPrivateIP = errors.New("cannot geolocate private/loopback IP")

	// ErrLookupFailed indicates the IP could not be resolved to a location.
	ErrLookupFailed = errors.New("geolocation lookup failed")
)
