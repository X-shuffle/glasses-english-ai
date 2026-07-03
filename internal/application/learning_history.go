package application

import (
	"context"
	"errors"
	"strings"

	"glasses-english-ai/internal/domain"
)

type LearningHistoryUseCase struct {
	history domain.LearningHistoryRepository
}

func NewLearningHistoryUseCase(history domain.LearningHistoryRepository) *LearningHistoryUseCase {
	return &LearningHistoryUseCase{history: history}
}

func (uc *LearningHistoryUseCase) Record(ctx context.Context, deviceID string, encounters []domain.LearningEncounter) ([]domain.LearnedWord, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil, errors.New("device_id is required")
	}

	cleaned := make([]domain.LearningEncounter, 0, len(encounters))
	for _, encounter := range encounters {
		english := strings.ToLower(strings.TrimSpace(encounter.English))
		if english == "" {
			continue
		}
		cleaned = append(cleaned, domain.LearningEncounter{
			English: english,
			Chinese: strings.TrimSpace(encounter.Chinese),
		})
	}
	if len(cleaned) == 0 {
		return uc.history.ListLearnedWords(ctx, deviceID)
	}
	return uc.history.RecordEncounters(ctx, deviceID, cleaned)
}

func (uc *LearningHistoryUseCase) List(ctx context.Context, deviceID string) ([]domain.LearnedWord, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil, errors.New("device_id is required")
	}
	return uc.history.ListLearnedWords(ctx, deviceID)
}

func (uc *LearningHistoryUseCase) Clear(ctx context.Context, deviceID string) error {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return errors.New("device_id is required")
	}
	return uc.history.ClearLearnedWords(ctx, deviceID)
}
