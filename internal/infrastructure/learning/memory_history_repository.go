package learning

import (
	"context"
	"sort"
	"sync"
	"time"

	"glasses-english-ai/internal/domain"
)

type MemoryHistoryRepository struct {
	mu    sync.RWMutex
	words map[string]map[string]domain.LearnedWord
}

func NewMemoryHistoryRepository() *MemoryHistoryRepository {
	return &MemoryHistoryRepository{
		words: make(map[string]map[string]domain.LearnedWord),
	}
}

func (r *MemoryHistoryRepository) RecordEncounters(_ context.Context, deviceID string, encounters []domain.LearningEncounter) ([]domain.LearnedWord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.words[deviceID]; !ok {
		r.words[deviceID] = make(map[string]domain.LearnedWord)
	}

	now := time.Now().UTC()
	for _, encounter := range encounters {
		word := r.words[deviceID][encounter.English]
		word.DeviceID = deviceID
		word.English = encounter.English
		if encounter.Chinese != "" {
			word.Chinese = encounter.Chinese
		}
		word.Count++
		word.LastSeen = now
		r.words[deviceID][encounter.English] = word
	}

	return sortedWords(r.words[deviceID]), nil
}

func (r *MemoryHistoryRepository) ListLearnedWords(_ context.Context, deviceID string) ([]domain.LearnedWord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return sortedWords(r.words[deviceID]), nil
}

func (r *MemoryHistoryRepository) ClearLearnedWords(_ context.Context, deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.words, deviceID)
	return nil
}

func sortedWords(words map[string]domain.LearnedWord) []domain.LearnedWord {
	result := make([]domain.LearnedWord, 0, len(words))
	for _, word := range words {
		result = append(result, word)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].English < result[j].English
		}
		return result[i].Count > result[j].Count
	})
	return result
}
