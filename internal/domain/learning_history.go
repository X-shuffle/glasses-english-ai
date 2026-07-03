package domain

import (
	"context"
	"time"
)

type LearningEncounter struct {
	English string
	Chinese string
}

type LearnedWord struct {
	DeviceID string
	English  string
	Chinese  string
	Count    int
	LastSeen time.Time
}

type LearningHistoryRepository interface {
	RecordEncounters(ctx context.Context, deviceID string, encounters []LearningEncounter) ([]LearnedWord, error)
	ListLearnedWords(ctx context.Context, deviceID string) ([]LearnedWord, error)
	ClearLearnedWords(ctx context.Context, deviceID string) error
}
