package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"glasses-english-ai/internal/domain"
)

func TestOpenAIProviderRecognizesObjects(t *testing.T) {
	var authHeader string
	var path string
	var model string
	var imageURL string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		path = r.URL.Path

		var req openAIChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		model = req.Model
		imageURL = req.Messages[1].Content[1].ImageURL.URL

		_ = json.NewEncoder(w).Encode(openAIChatCompletionResponse{
			Choices: []openAIChoice{
				{
					Message: openAIChoiceMessage{
						Content: `{"objects":[{"english":"Chair","score":0.94,"box":{"x":100,"y":90,"width":140,"height":180}}]}`,
					},
				},
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-test",
		Client:  server.Client(),
	})
	scene, err := provider.Recognize(context.Background(), domain.Frame{
		DeviceID:    "glass_001",
		FrameID:     "frame_1",
		ImageBase64: "data:image/jpeg;base64,dGVzdA==",
	})
	if err != nil {
		t.Fatal(err)
	}

	if authHeader != "Bearer test-key" {
		t.Fatalf("expected auth header, got %q", authHeader)
	}
	if path != "/chat/completions" {
		t.Fatalf("expected chat completions path, got %q", path)
	}
	if model != "gpt-test" {
		t.Fatalf("expected configured model, got %q", model)
	}
	if !strings.HasPrefix(imageURL, "data:image/jpeg;base64,") {
		t.Fatalf("expected data URL image, got %q", imageURL)
	}
	if len(scene.Objects) != 1 {
		t.Fatalf("expected one object, got %d", len(scene.Objects))
	}
	if scene.Objects[0].English != "chair" {
		t.Fatalf("expected normalized english chair, got %q", scene.Objects[0].English)
	}
}

func TestNewProviderFallsBackToMockWhenOpenAIKeyMissing(t *testing.T) {
	provider := NewProvider(ProviderConfig{Provider: "openai"})

	if _, ok := provider.(*MockProvider); !ok {
		t.Fatalf("expected mock provider fallback, got %T", provider)
	}
}
