package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr              string
	CacheMaxItems     int
	VisionProvider    string
	CloudVisionURL    string
	CloudVisionAPIKey string
}

func Load() Config {
	return Config{
		Addr:              getEnv("APP_ADDR", ":8080"),
		CacheMaxItems:     getEnvInt("CACHE_MAX_ITEMS", 1000),
		VisionProvider:    getEnv("VISION_PROVIDER", "mock"),
		CloudVisionURL:    getEnv("CLOUD_VISION_ENDPOINT", ""),
		CloudVisionAPIKey: getEnv("CLOUD_VISION_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
