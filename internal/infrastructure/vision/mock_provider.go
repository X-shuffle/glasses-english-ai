package vision

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"glasses-english-ai/internal/domain"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Recognize(_ context.Context, frame domain.Frame) (domain.SceneRecognition, error) {
	sceneHash := hashFrame(frame.ImageBase64)
	if frame.ImageBase64 == "" && frame.LastSceneHash != "" {
		sceneHash = frame.LastSceneHash
	}

	objects := []domain.VisualObject{
		{English: "cup", Box: domain.BoundingBox{X: 120, Y: 80, Width: 90, Height: 110}, Score: 0.92},
		{English: "book", Box: domain.BoundingBox{X: 260, Y: 140, Width: 150, Height: 80}, Score: 0.89},
		{English: "phone", Box: domain.BoundingBox{X: 440, Y: 190, Width: 80, Height: 130}, Score: 0.84},
	}

	if strings.Contains(strings.ToLower(frame.ImageBase64), "desk") {
		objects = append(objects, domain.VisualObject{
			English: "pen",
			Box:     domain.BoundingBox{X: 540, Y: 120, Width: 120, Height: 40},
			Score:   0.81,
		})
	}

	return domain.SceneRecognition{
		SceneHash: sceneHash,
		Objects:   objects,
	}, nil
}

func hashFrame(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])[:16]
}
