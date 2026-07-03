package main

import (
	"log"
	"net/http"
	"time"

	"glasses-english-ai/internal/cache"
	"glasses-english-ai/internal/config"
	"glasses-english-ai/internal/httpapi"
	"glasses-english-ai/internal/vision"
)

func main() {
	cfg := config.Load()

	sceneCache := cache.NewSceneCache(cfg.CacheMaxItems, 12*time.Hour)
	recognizer := vision.NewMockRecognizer()
	server := httpapi.NewServer(sceneCache, recognizer)

	log.Printf("glasses english ai server listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
