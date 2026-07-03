package vision

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"glasses-english-ai/internal/domain"
)

type CloudProvider struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

func NewCloudProvider(endpoint, apiKey string, client *http.Client) *CloudProvider {
	if client == nil {
		client = http.DefaultClient
	}
	return &CloudProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   client,
	}
}

func (p *CloudProvider) Recognize(ctx context.Context, frame domain.Frame) (domain.SceneRecognition, error) {
	if p.endpoint == "" {
		return domain.SceneRecognition{}, errors.New("cloud vision endpoint is required")
	}

	payload := cloudRecognitionRequest{
		DeviceID:      frame.DeviceID,
		FrameID:       frame.FrameID,
		ImageBase64:   frame.ImageBase64,
		LastSceneHash: frame.LastSceneHash,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return domain.SceneRecognition{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return domain.SceneRecognition{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return domain.SceneRecognition{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.SceneRecognition{}, fmt.Errorf("cloud vision provider returned status %d", resp.StatusCode)
	}

	var result cloudRecognitionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.SceneRecognition{}, err
	}
	if result.SceneHash == "" {
		result.SceneHash = hashFrame(frame.ImageBase64)
	}

	objects := make([]domain.VisualObject, 0, len(result.Objects))
	for _, object := range result.Objects {
		if object.English == "" {
			object.English = object.Name
		}
		if object.English == "" {
			continue
		}
		objects = append(objects, domain.VisualObject{
			English: object.English,
			Box: domain.BoundingBox{
				X:      object.Box.X,
				Y:      object.Box.Y,
				Width:  object.Box.Width,
				Height: object.Box.Height,
			},
			Score: object.Score,
		})
	}

	return domain.SceneRecognition{
		SceneHash: result.SceneHash,
		Objects:   objects,
	}, nil
}

type cloudRecognitionRequest struct {
	DeviceID      string `json:"device_id"`
	FrameID       string `json:"frame_id"`
	ImageBase64   string `json:"image_base64"`
	LastSceneHash string `json:"last_scene_hash,omitempty"`
}

type cloudRecognitionResponse struct {
	SceneHash string        `json:"scene_hash"`
	Objects   []cloudObject `json:"objects"`
}

type cloudObject struct {
	Name    string   `json:"name,omitempty"`
	English string   `json:"english,omitempty"`
	Box     cloudBox `json:"box"`
	Score   float64  `json:"score"`
}

type cloudBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
