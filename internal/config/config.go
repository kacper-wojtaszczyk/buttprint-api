package config

import "os"

type Config struct {
	Port         string
	JackfruitUrl string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		JackfruitUrl: getEnv("JACKFRUIT_URL", "localhost:8080"),
	}
}
