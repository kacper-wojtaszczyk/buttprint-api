package api

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string // empty = header not set
		remoteAddr    string
		want          string
	}{
		{
			name:          "XFF single IP",
			xForwardedFor: "203.0.113.50",
			remoteAddr:    "10.0.0.1:12345",
			want:          "203.0.113.50",
		},
		{
			name:          "XFF multiple IPs takes first",
			xForwardedFor: "203.0.113.50, 70.41.3.18, 150.172.238.178",
			remoteAddr:    "10.0.0.1:12345",
			want:          "203.0.113.50",
		},
		{
			name:          "XFF with spaces around commas",
			xForwardedFor: "  203.0.113.50 , 70.41.3.18 ",
			remoteAddr:    "10.0.0.1:12345",
			want:          "203.0.113.50",
		},
		{
			name:       "no XFF falls back to IPv4 RemoteAddr",
			remoteAddr: "192.168.1.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "no XFF falls back to IPv6 RemoteAddr",
			remoteAddr: "[::1]:12345",
			want:       "::1",
		},
		{
			name:          "empty XFF falls back to RemoteAddr",
			xForwardedFor: "",
			remoteAddr:    "192.168.1.1:54321",
			want:          "192.168.1.1",
		},
		{
			name:          "XFF only whitespace falls back to RemoteAddr",
			xForwardedFor: "   ",
			remoteAddr:    "10.0.0.2:8080",
			want:          "10.0.0.2",
		},
		{
			name:       "RemoteAddr without port returned as-is",
			remoteAddr: "192.168.1.1",
			want:       "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tt.xForwardedFor != "" {
				r.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			r.RemoteAddr = tt.remoteAddr

			got := clientIP(r)
			if got != tt.want {
				t.Errorf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
