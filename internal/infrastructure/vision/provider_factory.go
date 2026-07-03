package vision

import (
	"log"
	"net/http"
	"strings"
	"time"

	"glasses-english-ai/internal/domain"
)

type ProviderConfig struct {
	Provider string
	Endpoint string
	APIKey   string
}

func NewProvider(cfg ProviderConfig) domain.RecognitionProvider {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "cloud", "http":
		if cfg.Endpoint == "" {
			log.Println("CLOUD_VISION_ENDPOINT is empty; falling back to mock vision provider")
			return NewMockProvider()
		}
		return NewCloudProvider(cfg.Endpoint, cfg.APIKey, &http.Client{Timeout: 15 * time.Second})
	default:
		return NewMockProvider()
	}
}
