package domain

import (
	"context"
	"errors"
	"fmt"
)

var ErrSceneNotFound = errors.New("scene not found")

type Frame struct {
	DeviceID      string
	FrameID       string
	ImageBase64   string
	LastSceneHash string
	OfflineOK     bool
}

func (f Frame) Validate() error {
	if f.DeviceID == "" {
		return errors.New("device_id is required")
	}
	if f.FrameID == "" {
		return errors.New("frame_id is required")
	}
	if f.ImageBase64 == "" && f.LastSceneHash == "" {
		return errors.New("image_base64 or last_scene_hash is required")
	}
	return nil
}

type SceneRecognition struct {
	SceneHash string
	FromCache bool
	Objects   []VisualObject
}

func (s SceneRecognition) WithCacheFlag(fromCache bool) SceneRecognition {
	s.FromCache = fromCache
	return s
}

func (s SceneRecognition) AssignDisplayLabels(ctx context.Context, cards LearningContentProvider) SceneRecognition {
	for i := range s.Objects {
		s.Objects[i].Letter = letterForIndex(i)
		card, err := cards.FindCard(ctx, s.Objects[i].English)
		if err == nil {
			s.Objects[i].Learning = card
			s.Objects[i].Chinese = card.Chinese
			s.Objects[i].Phonetic = card.Phonetic
			s.Objects[i].Sentence = card.ExampleSentence
		}
		s.Objects[i].DisplayText = fmt.Sprintf("%s %s / %s", s.Objects[i].Letter, s.Objects[i].English, s.Objects[i].Chinese)
		s.Objects[i].SpeakText = fmt.Sprintf("%s. %s. %s", s.Objects[i].Letter, s.Objects[i].English, s.Objects[i].Sentence)
	}
	return s
}

type VisualObject struct {
	Letter      string
	English     string
	Chinese     string
	Phonetic    string
	Sentence    string
	DisplayText string
	SpeakText   string
	Box         BoundingBox
	Score       float64
	Learning    LearningCard
}

type BoundingBox struct {
	X      int
	Y      int
	Width  int
	Height int
}

type LearningCard struct {
	English         string
	Chinese         string
	Phonetic        string
	ExampleSentence string
	ExampleMeaning  string
}

type SceneRepository interface {
	FindByHash(ctx context.Context, sceneHash string) (SceneRecognition, error)
	Save(ctx context.Context, scene SceneRecognition) error
}

type RecognitionProvider interface {
	Recognize(ctx context.Context, frame Frame) (SceneRecognition, error)
}

type LearningContentProvider interface {
	FindCard(ctx context.Context, english string) (LearningCard, error)
}

func letterForIndex(index int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if index < len(alphabet) {
		return string(alphabet[index])
	}
	return fmt.Sprintf("Z%d", index-len(alphabet)+1)
}
