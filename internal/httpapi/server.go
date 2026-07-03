package httpapi

import (
	"encoding/json"
	"net/http"

	"glasses-english-ai/internal/cache"
	"glasses-english-ai/internal/vision"
)

type Server struct {
	cache      *cache.SceneCache
	recognizer vision.Recognizer
}

func NewServer(sceneCache *cache.SceneCache, recognizer vision.Recognizer) *Server {
	return &Server{cache: sceneCache, recognizer: recognizer}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /api/vision/recognize", s.recognize)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) recognize(w http.ResponseWriter, r *http.Request) {
	var req vision.RecognitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.LastSceneHash != "" {
		if cached, ok := s.cache.Get(req.LastSceneHash); ok {
			cached.FromCache = true
			writeJSON(w, http.StatusOK, cached)
			return
		}
	}

	result, err := s.recognizer.Recognize(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "vision provider failed")
		return
	}

	if result.SceneHash != "" {
		s.cache.Set(result.SceneHash, result)
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
