package cache

import (
	"context"
	"sync"
	"time"

	"glasses-english-ai/internal/domain"
)

type MemorySceneRepository struct {
	mu       sync.RWMutex
	maxItems int
	ttl      time.Duration
	items    map[string]cacheItem
}

type cacheItem struct {
	scene     domain.SceneRecognition
	createdAt time.Time
}

func NewMemorySceneRepository(maxItems int, ttl time.Duration) *MemorySceneRepository {
	if maxItems <= 0 {
		maxItems = 1000
	}
	return &MemorySceneRepository{
		maxItems: maxItems,
		ttl:      ttl,
		items:    make(map[string]cacheItem),
	}
}

func (r *MemorySceneRepository) FindByHash(_ context.Context, sceneHash string) (domain.SceneRecognition, error) {
	r.mu.RLock()
	item, ok := r.items[sceneHash]
	r.mu.RUnlock()

	if !ok {
		return domain.SceneRecognition{}, domain.ErrSceneNotFound
	}
	if time.Since(item.createdAt) > r.ttl {
		r.mu.Lock()
		delete(r.items, sceneHash)
		r.mu.Unlock()
		return domain.SceneRecognition{}, domain.ErrSceneNotFound
	}
	return item.scene, nil
}

func (r *MemorySceneRepository) Save(_ context.Context, scene domain.SceneRecognition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.items) >= r.maxItems {
		r.evictOldest()
	}
	r.items[scene.SceneHash] = cacheItem{scene: scene, createdAt: time.Now()}
	return nil
}

func (r *MemorySceneRepository) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range r.items {
		if first || item.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.createdAt
			first = false
		}
	}
	if oldestKey != "" {
		delete(r.items, oldestKey)
	}
}
