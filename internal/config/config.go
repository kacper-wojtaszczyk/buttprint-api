package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	JackfruitURL       string
	MaxMindDBPath      string
	CORSAllowedOrigins []string
	RateLimitRPS       float64
	RateLimitBurst     int
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

func getEnvFloat64(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func Load() *Config {
	return &Config{
		Port:               getEnvString("PORT", "8080"),
		JackfruitURL:       getEnvString("JACKFRUIT_URL", "http://localhost:8080"),
		MaxMindDBPath:      getEnvString("MAXMIND_DB_PATH", "/data/GeoLite2-City.mmdb"),
		CORSAllowedOrigins: getEnvStringSlice("CORS_ALLOWED_ORIGINS", ",", []string{"http://localhost:5173"}),
		RateLimitRPS:       getEnvFloat64("RATE_LIMIT_RPS", 10),
		RateLimitBurst:     getEnvInt("RATE_LIMIT_BURST", 20),
	}
}
