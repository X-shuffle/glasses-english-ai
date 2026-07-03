package vision

import (
	"log"
	"net/http"
	"strings"
	"time"

	"glasses-english-ai/internal/domain"
)

type ProviderConfig struct {
	Provider      string
	Endpoint      string
	APIKey        string
	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string
}

func NewProvider(cfg ProviderConfig) domain.RecognitionProvider {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "cloud", "http":
		if cfg.Endpoint == "" {
			log.Println("CLOUD_VISION_ENDPOINT is empty; falling back to mock vision provider")
			return NewMockProvider()
		}
		return NewCloudProvider(cfg.Endpoint, cfg.APIKey, &http.Client{Timeout: 15 * time.Second})
	case "openai":
		if cfg.OpenAIAPIKey == "" {
			log.Println("OPENAI_API_KEY is empty; falling back to mock vision provider")
			return NewMockProvider()
		}
		return NewOpenAIProvider(OpenAIProviderConfig{
			APIKey:  cfg.OpenAIAPIKey,
			BaseURL: cfg.OpenAIBaseURL,
			Model:   cfg.OpenAIModel,
			Client:  &http.Client{Timeout: 30 * time.Second},
		})
	default:
		return NewMockProvider()
	}
}
