package cache

import (
	"sync"
	"time"

	"glasses-english-ai/internal/vision"
)

type SceneCache struct {
	mu       sync.RWMutex
	maxItems int
	ttl      time.Duration
	items    map[string]cacheItem
}

type cacheItem struct {
	result    vision.RecognitionResult
	createdAt time.Time
}

func NewSceneCache(maxItems int, ttl time.Duration) *SceneCache {
	if maxItems <= 0 {
		maxItems = 1000
	}
	return &SceneCache{
		maxItems: maxItems,
		ttl:      ttl,
		items:    make(map[string]cacheItem),
	}
}

func (c *SceneCache) Get(sceneHash string) (vision.RecognitionResult, bool) {
	c.mu.RLock()
	item, ok := c.items[sceneHash]
	c.mu.RUnlock()

	if !ok {
		return vision.RecognitionResult{}, false
	}
	if time.Since(item.createdAt) > c.ttl {
		c.mu.Lock()
		delete(c.items, sceneHash)
		c.mu.Unlock()
		return vision.RecognitionResult{}, false
	}
	return item.result, true
}

func (c *SceneCache) Set(sceneHash string, result vision.RecognitionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.maxItems {
		c.evictOldest()
	}
	c.items[sceneHash] = cacheItem{result: result, createdAt: time.Now()}
}

func (c *SceneCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range c.items {
		if first || item.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.createdAt
			first = false
		}
	}
	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}
