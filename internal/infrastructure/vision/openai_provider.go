package vision

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"glasses-english-ai/internal/domain"
)

type OpenAIProviderConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewOpenAIProvider(cfg OpenAIProviderConfig) *OpenAIProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-5.5"
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	return &OpenAIProvider{
		apiKey:  cfg.APIKey,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		model:   cfg.Model,
		client:  cfg.Client,
	}
}

func (p *OpenAIProvider) Recognize(ctx context.Context, frame domain.Frame) (domain.SceneRecognition, error) {
	if p.apiKey == "" {
		return domain.SceneRecognition{}, errors.New("openai api key is required")
	}
	if frame.ImageBase64 == "" {
		return domain.SceneRecognition{}, errors.New("image_base64 is required for openai vision provider")
	}

	body, err := json.Marshal(p.newRequest(frame.ImageBase64))
	if err != nil {
		return domain.SceneRecognition{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return domain.SceneRecognition{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return domain.SceneRecognition{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.SceneRecognition{}, fmt.Errorf("openai vision provider returned status %d", resp.StatusCode)
	}

	var completion openAIChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return domain.SceneRecognition{}, err
	}
	if len(completion.Choices) == 0 {
		return domain.SceneRecognition{}, errors.New("openai vision provider returned no choices")
	}

	var parsed openAIRecognitionOutput
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &parsed); err != nil {
		return domain.SceneRecognition{}, fmt.Errorf("parse openai recognition output: %w", err)
	}

	objects := make([]domain.VisualObject, 0, len(parsed.Objects))
	for _, object := range parsed.Objects {
		if strings.TrimSpace(object.English) == "" {
			continue
		}
		objects = append(objects, domain.VisualObject{
			English: strings.ToLower(strings.TrimSpace(object.English)),
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
		SceneHash: hashFrame(frame.ImageBase64),
		Objects:   objects,
	}, nil
}

func (p *OpenAIProvider) newRequest(imageDataURL string) openAIChatCompletionRequest {
	return openAIChatCompletionRequest{
		Model: p.model,
		Messages: []openAIMessage{
			{
				Role: "system",
				Content: []openAIContent{
					{
						Type: "text",
						Text: "You identify visible everyday objects for a Chinese learner of English. Return only common concrete objects that are useful to label in a glasses HUD.",
					},
				},
			},
			{
				Role: "user",
				Content: []openAIContent{
					{
						Type: "text",
						Text: "Identify up to 10 visible objects. Use simple singular English nouns. Return bounding boxes in an 800x450 coordinate space.",
					},
					{
						Type: "image_url",
						ImageURL: &openAIImageURL{
							URL: imageDataURL,
						},
					},
				},
			},
		},
		ResponseFormat: openAIResponseFormat{
			Type: "json_schema",
			JSONSchema: openAIJSONSchema{
				Name:   "glasses_vision_objects",
				Strict: true,
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"objects": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"english": map[string]any{"type": "string"},
									"score":   map[string]any{"type": "number"},
									"box": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"x":      map[string]any{"type": "integer"},
											"y":      map[string]any{"type": "integer"},
											"width":  map[string]any{"type": "integer"},
											"height": map[string]any{"type": "integer"},
										},
										"required":             []string{"x", "y", "width", "height"},
										"additionalProperties": false,
									},
								},
								"required":             []string{"english", "score", "box"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"objects"},
					"additionalProperties": false,
				},
			},
		},
		MaxCompletionTokens: 700,
	}
}

type openAIChatCompletionRequest struct {
	Model               string               `json:"model"`
	Messages            []openAIMessage      `json:"messages"`
	ResponseFormat      openAIResponseFormat `json:"response_format"`
	MaxCompletionTokens int                  `json:"max_completion_tokens"`
}

type openAIMessage struct {
	Role    string          `json:"role"`
	Content []openAIContent `json:"content"`
}

type openAIContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *openAIImageURL `json:"image_url,omitempty"`
}

type openAIImageURL struct {
	URL string `json:"url"`
}

type openAIResponseFormat struct {
	Type       string           `json:"type"`
	JSONSchema openAIJSONSchema `json:"json_schema"`
}

type openAIJSONSchema struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type openAIChatCompletionResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIChoiceMessage `json:"message"`
}

type openAIChoiceMessage struct {
	Content string `json:"content"`
}

type openAIRecognitionOutput struct {
	Objects []openAIRecognitionObject `json:"objects"`
}

type openAIRecognitionObject struct {
	English string    `json:"english"`
	Score   float64   `json:"score"`
	Box     openAIBox `json:"box"`
}

type openAIBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
