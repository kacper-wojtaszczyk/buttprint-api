package config

import "os"

type Config struct {
	Port          string
	JackfruitURL  string
	MaxMindDBPath string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		JackfruitURL:  getEnv("JACKFRUIT_URL", "http://localhost:8080"),
		MaxMindDBPath: getEnv("MAXMIND_DB_PATH", "/data/GeoLite2-City.mmdb"),
	}
}
