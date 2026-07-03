package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"glasses-english-ai/internal/cache"
	"glasses-english-ai/internal/vision"
)

func TestRecognizeReturnsObjects(t *testing.T) {
	server := NewServer(cache.NewSceneCache(10, time.Hour), vision.NewMockRecognizer())
	body := bytes.NewBufferString(`{"device_id":"glass_001","frame_id":"f_1","image_base64":"demo"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/vision/recognize", body)
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var result vision.RecognitionResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Objects) == 0 {
		t.Fatal("expected at least one recognized object")
	}
	if result.Objects[0].Letter != "A" {
		t.Fatalf("expected first object letter A, got %q", result.Objects[0].Letter)
	}
}
