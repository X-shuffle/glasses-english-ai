package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"glasses-english-ai/internal/domain"
)

func TestCloudProviderRecognizesObjects(t *testing.T) {
	var authHeader string
	var receivedImage string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")

		var req cloudRecognitionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		receivedImage = req.ImageBase64

		_ = json.NewEncoder(w).Encode(cloudRecognitionResponse{
			SceneHash: "scene_cloud_1",
			Objects: []cloudObject{
				{
					English: "chair",
					Box:     cloudBox{X: 10, Y: 20, Width: 80, Height: 90},
					Score:   0.91,
				},
			},
		})
	}))
	defer server.Close()

	provider := NewCloudProvider(server.URL, "test-token", server.Client())
	scene, err := provider.Recognize(context.Background(), domain.Frame{
		DeviceID:    "glass_001",
		FrameID:     "frame_1",
		ImageBase64: "data:image/jpeg;base64,dGVzdA==",
	})
	if err != nil {
		t.Fatal(err)
	}

	if authHeader != "Bearer test-token" {
		t.Fatalf("expected authorization header, got %q", authHeader)
	}
	if receivedImage == "" {
		t.Fatal("expected cloud provider to receive image")
	}
	if scene.SceneHash != "scene_cloud_1" {
		t.Fatalf("expected cloud scene hash, got %q", scene.SceneHash)
	}
	if len(scene.Objects) != 1 || scene.Objects[0].English != "chair" {
		t.Fatalf("expected chair object, got %#v", scene.Objects)
	}
}

func TestNewProviderFallsBackToMockWhenCloudEndpointMissing(t *testing.T) {
	provider := NewProvider(ProviderConfig{Provider: "cloud"})

	if _, ok := provider.(*MockProvider); !ok {
		t.Fatalf("expected mock provider fallback, got %T", provider)
	}
}
