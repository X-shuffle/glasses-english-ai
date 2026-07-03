package main

import (
	"log"
	"net/http"
	"time"

	"glasses-english-ai/internal/application"
	"glasses-english-ai/internal/config"
	infraCache "glasses-english-ai/internal/infrastructure/cache"
	"glasses-english-ai/internal/infrastructure/learning"
	infraVision "glasses-english-ai/internal/infrastructure/vision"
	"glasses-english-ai/internal/interfaces/httpapi"
)

func main() {
	cfg := config.Load()

	sceneRepo := infraCache.NewMemorySceneRepository(cfg.CacheMaxItems, 12*time.Hour)
	recognizer := infraVision.NewProvider(infraVision.ProviderConfig{
		Provider:      cfg.VisionProvider,
		Endpoint:      cfg.CloudVisionURL,
		APIKey:        cfg.CloudVisionAPIKey,
		OpenAIAPIKey:  cfg.OpenAIAPIKey,
		OpenAIBaseURL: cfg.OpenAIBaseURL,
		OpenAIModel:   cfg.OpenAIModel,
	})
	dictionary := learning.NewStaticDictionary()
	learningHistoryRepo := learning.NewMemoryHistoryRepository()
	recognizeFrame := application.NewRecognizeFrameUseCase(sceneRepo, recognizer, dictionary)
	learningHistory := application.NewLearningHistoryUseCase(learningHistoryRepo)
	server := httpapi.NewServer(recognizeFrame, learningHistory)

	log.Printf("glasses english ai server listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
