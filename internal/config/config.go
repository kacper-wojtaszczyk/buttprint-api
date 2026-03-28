package config

import (
	"os"
	"strings"
)

type Config struct {
	Port               string
	JackfruitURL       string
	MaxMindDBPath      string
	CORSAllowedOrigins []string
}

func getEnvString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func getEnvStringSlice(key, sep string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, sep)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func Load() *Config {
	return &Config{
		Port:               getEnvString("PORT", "8080"),
		JackfruitURL:       getEnvString("JACKFRUIT_URL", "http://localhost:8080"),
		MaxMindDBPath:      getEnvString("MAXMIND_DB_PATH", "/data/GeoLite2-City.mmdb"),
		CORSAllowedOrigins: getEnvStringSlice("CORS_ALLOWED_ORIGINS", ",", []string{"http://localhost:5173"}),
	}
}
