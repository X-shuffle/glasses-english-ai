package httpapi

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"

	"glasses-english-ai/internal/application"
	"glasses-english-ai/internal/domain"
)

//go:embed static
var staticFiles embed.FS

type Server struct {
	recognizeFrame *application.RecognizeFrameUseCase
}

func NewServer(recognizeFrame *application.RecognizeFrameUseCase) *Server {
	return &Server{recognizeFrame: recognizeFrame}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.demo)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(mustStaticSubtree()))))
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /api/vision/recognize", s.recognize)
	return mux
}

func (s *Server) demo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFileFS(w, r, staticFiles, "static/index.html")
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) recognize(w http.ResponseWriter, r *http.Request) {
	var req recognitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	result, err := s.recognizeFrame.Execute(r.Context(), application.RecognizeFrameCommand{
		DeviceID:      req.DeviceID,
		FrameID:       req.FrameID,
		ImageBase64:   req.ImageBase64,
		LastSceneHash: req.LastSceneHash,
		OfflineOK:     req.OfflineOK,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSceneNotFound) {
			writeError(w, http.StatusNotFound, "scene not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, newRecognitionResponse(result))
}

type recognitionRequest struct {
	DeviceID      string `json:"device_id"`
	FrameID       string `json:"frame_id"`
	ImageBase64   string `json:"image_base64"`
	LastSceneHash string `json:"last_scene_hash,omitempty"`
	OfflineOK     bool   `json:"offline_ok,omitempty"`
}

type recognitionResponse struct {
	SceneHash string           `json:"scene_hash"`
	FromCache bool             `json:"from_cache"`
	Objects   []objectResponse `json:"objects"`
}

type objectResponse struct {
	Letter      string       `json:"letter"`
	Name        string       `json:"name"`
	English     string       `json:"english"`
	Meaning     string       `json:"meaning"`
	Chinese     string       `json:"chinese"`
	Phonetic    string       `json:"phonetic"`
	Sentence    string       `json:"sentence"`
	DisplayText string       `json:"display_text"`
	SpeakText   string       `json:"speak_text"`
	Box         boxResponse  `json:"box"`
	Score       float64      `json:"score"`
	Learning    learningCard `json:"learning"`
}

type boxResponse struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type learningCard struct {
	English         string `json:"english"`
	Chinese         string `json:"chinese"`
	Phonetic        string `json:"phonetic"`
	ExampleSentence string `json:"example_sentence"`
	ExampleMeaning  string `json:"example_meaning"`
}

func newRecognitionResponse(scene domain.SceneRecognition) recognitionResponse {
	objects := make([]objectResponse, 0, len(scene.Objects))
	for _, object := range scene.Objects {
		objects = append(objects, objectResponse{
			Letter:      object.Letter,
			Name:        object.English,
			English:     object.English,
			Meaning:     object.Chinese,
			Chinese:     object.Chinese,
			Phonetic:    object.Phonetic,
			Sentence:    object.Sentence,
			DisplayText: object.DisplayText,
			SpeakText:   object.SpeakText,
			Box: boxResponse{
				X:      object.Box.X,
				Y:      object.Box.Y,
				Width:  object.Box.Width,
				Height: object.Box.Height,
			},
			Score: object.Score,
			Learning: learningCard{
				English:         object.Learning.English,
				Chinese:         object.Learning.Chinese,
				Phonetic:        object.Learning.Phonetic,
				ExampleSentence: object.Learning.ExampleSentence,
				ExampleMeaning:  object.Learning.ExampleMeaning,
			},
		})
	}
	return recognitionResponse{
		SceneHash: scene.SceneHash,
		FromCache: scene.FromCache,
		Objects:   objects,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func mustStaticSubtree() fs.FS {
	subtree, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return subtree
}
