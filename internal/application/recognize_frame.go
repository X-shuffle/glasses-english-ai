package application

import (
	"context"

	"glasses-english-ai/internal/domain"
)

type RecognizeFrameCommand struct {
	DeviceID      string
	FrameID       string
	ImageBase64   string
	LastSceneHash string
	OfflineOK     bool
}

type RecognizeFrameUseCase struct {
	scenes     domain.SceneRepository
	recognizer domain.RecognitionProvider
	learning   domain.LearningContentProvider
}

func NewRecognizeFrameUseCase(
	scenes domain.SceneRepository,
	recognizer domain.RecognitionProvider,
	learning domain.LearningContentProvider,
) *RecognizeFrameUseCase {
	return &RecognizeFrameUseCase{
		scenes:     scenes,
		recognizer: recognizer,
		learning:   learning,
	}
}

func (uc *RecognizeFrameUseCase) Execute(ctx context.Context, cmd RecognizeFrameCommand) (domain.SceneRecognition, error) {
	frame := domain.Frame{
		DeviceID:      cmd.DeviceID,
		FrameID:       cmd.FrameID,
		ImageBase64:   cmd.ImageBase64,
		LastSceneHash: cmd.LastSceneHash,
		OfflineOK:     cmd.OfflineOK,
	}
	if err := frame.Validate(); err != nil {
		return domain.SceneRecognition{}, err
	}

	if frame.LastSceneHash != "" {
		scene, err := uc.scenes.FindByHash(ctx, frame.LastSceneHash)
		if err == nil {
			return scene.WithCacheFlag(true), nil
		}
	}

	scene, err := uc.recognizer.Recognize(ctx, frame)
	if err != nil {
		if frame.OfflineOK && frame.LastSceneHash != "" {
			cached, cacheErr := uc.scenes.FindByHash(ctx, frame.LastSceneHash)
			if cacheErr == nil {
				return cached.WithCacheFlag(true), nil
			}
		}
		return domain.SceneRecognition{}, err
	}

	scene = scene.AssignDisplayLabels(ctx, uc.learning)
	if scene.SceneHash != "" {
		if err := uc.scenes.Save(ctx, scene); err != nil {
			return domain.SceneRecognition{}, err
		}
	}
	return scene, nil
}
