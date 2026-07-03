package vision

import (
	"crypto/sha256"
	"encoding/hex"
)

type MockRecognizer struct{}

func NewMockRecognizer() *MockRecognizer {
	return &MockRecognizer{}
}

func (r *MockRecognizer) Recognize(req RecognitionRequest) (RecognitionResult, error) {
	sceneHash := hashFrame(req.ImageBase64)
	if req.ImageBase64 == "" && req.LastSceneHash != "" {
		sceneHash = req.LastSceneHash
	}

	return RecognitionResult{
		SceneHash: sceneHash,
		Objects: []Object{
			{
				Letter:   "A",
				Name:     "cup",
				Meaning:  "杯子",
				Phonetic: "/kʌp/",
				Sentence: "This is a cup.",
				Box:      Box{X: 120, Y: 80, Width: 90, Height: 110},
				Score:    0.92,
			},
			{
				Letter:   "B",
				Name:     "book",
				Meaning:  "书",
				Phonetic: "/bʊk/",
				Sentence: "This is a book.",
				Box:      Box{X: 260, Y: 140, Width: 150, Height: 80},
				Score:    0.89,
			},
		},
	}, nil
}

func hashFrame(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])[:16]
}
