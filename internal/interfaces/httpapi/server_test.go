package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"glasses-english-ai/internal/application"
	infraCache "glasses-english-ai/internal/infrastructure/cache"
	"glasses-english-ai/internal/infrastructure/learning"
	infraVision "glasses-english-ai/internal/infrastructure/vision"
)

func TestRecognizeReturnsBilingualDisplayObjects(t *testing.T) {
	server := newTestServer()
	body := bytes.NewBufferString(`{"device_id":"glass_001","frame_id":"f_1","image_base64":"desk demo"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/vision/recognize", body)
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result recognitionResponse
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Objects) < 3 {
		t.Fatalf("expected multiple recognized objects, got %d", len(result.Objects))
	}
	first := result.Objects[0]
	if first.Letter != "A" {
		t.Fatalf("expected first object letter A, got %q", first.Letter)
	}
	if first.English != "cup" || first.Chinese != "杯子" {
		t.Fatalf("expected cup bilingual result, got english=%q chinese=%q", first.English, first.Chinese)
	}
	if first.DisplayText != "A cup / 杯子" {
		t.Fatalf("expected display text, got %q", first.DisplayText)
	}
	if first.Learning.ExampleSentence == "" {
		t.Fatal("expected learning card example sentence")
	}
}

func TestRecognizeReusesSceneCache(t *testing.T) {
	server := newTestServer()
	body := bytes.NewBufferString(`{"device_id":"glass_001","frame_id":"f_1","image_base64":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/vision/recognize", body)
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	var first recognitionResponse
	if err := json.NewDecoder(rec.Body).Decode(&first); err != nil {
		t.Fatal(err)
	}

	body = bytes.NewBufferString(`{"device_id":"glass_001","frame_id":"f_2","last_scene_hash":"` + first.SceneHash + `"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/vision/recognize", body)
	rec = httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	var cached recognitionResponse
	if err := json.NewDecoder(rec.Body).Decode(&cached); err != nil {
		t.Fatal(err)
	}
	if !cached.FromCache {
		t.Fatal("expected cached response")
	}
	if cached.SceneHash != first.SceneHash {
		t.Fatalf("expected scene hash %q, got %q", first.SceneHash, cached.SceneHash)
	}
}

func TestDemoPageAndStaticAssetsAreServed(t *testing.T) {
	server := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected demo status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Glasses English AI") {
		t.Fatal("expected demo page content")
	}
	if !strings.Contains(rec.Body.String(), "cameraFeed") {
		t.Fatal("expected camera video element")
	}
	if !strings.Contains(rec.Body.String(), "autoBtn") {
		t.Fatal("expected auto scan control")
	}
	if !strings.Contains(rec.Body.String(), "已遇到的词") {
		t.Fatal("expected learned words panel")
	}

	req = httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rec = httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected static status 200, got %d", rec.Code)
	}
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	script := string(body)
	if !strings.Contains(script, "display_text") {
		t.Fatal("expected HUD script content")
	}
	if !strings.Contains(script, "getUserMedia") {
		t.Fatal("expected camera capture logic")
	}
	if !strings.Contains(script, "localStorage") {
		t.Fatal("expected local cache logic")
	}
	if !strings.Contains(script, "toggleAutoScan") {
		t.Fatal("expected auto scan logic")
	}
	if !strings.Contains(script, "speechSynthesis") {
		t.Fatal("expected TTS learning logic")
	}
	if !strings.Contains(script, "learnedWordsKey") {
		t.Fatal("expected learned words history logic")
	}
}

func newTestServer() *Server {
	sceneRepo := infraCache.NewMemorySceneRepository(10, time.Hour)
	recognizer := infraVision.NewMockProvider()
	dictionary := learning.NewStaticDictionary()
	useCase := application.NewRecognizeFrameUseCase(sceneRepo, recognizer, dictionary)
	return NewServer(useCase)
}
