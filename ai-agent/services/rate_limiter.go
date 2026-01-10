package services

import (
	"context"
	"sync"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
)

const (
	UploadLimitPerHour  = 10
	AIGenerationPerDay  = 20
	CleanupWindow       = 1 * time.Hour
)

// RateLimiter tracks rate limits for users
type RateLimiter struct {
	uploadLimits    map[string]*UserLimit
	aiGenLimits     map[string]*UserLimit
	mu              sync.RWMutex
	cleanupInterval time.Duration
}

// UserLimit tracks user's requests
type UserLimit struct {
	Count      int
	WindowStart time.Time
	WindowDuration time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		uploadLimits:    make(map[string]*UserLimit),
		aiGenLimits:     make(map[string]*UserLimit),
		cleanupInterval: CleanupWindow,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// CheckUploadLimit checks if user can upload
func (rl *RateLimiter) CheckUploadLimit(ctx context.Context, userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.uploadLimits[userID]
	if !exists {
		rl.uploadLimits[userID] = &UserLimit{
			Count:          1,
			WindowStart:    time.Now(),
			WindowDuration: 1 * time.Hour,
		}
		return true
	}

	// Check if window expired
	if time.Since(limit.WindowStart) > limit.WindowDuration {
		limit.Count = 1
		limit.WindowStart = time.Now()
		return true
	}

	// Check limit
	if limit.Count >= UploadLimitPerHour {
		commons.Log(ctx, "[RateLimiter.CheckUploadLimit::Exceeded [user_id=%s [count=%d", userID, limit.Count)
		return false
	}

	limit.Count++
	return true
}

// CheckAIGenerationLimit checks if user can generate quiz
func (rl *RateLimiter) CheckAIGenerationLimit(ctx context.Context, userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.aiGenLimits[userID]
	if !exists {
		rl.aiGenLimits[userID] = &UserLimit{
			Count:          1,
			WindowStart:    time.Now(),
			WindowDuration: 24 * time.Hour,
		}
		return true
	}

	// Check if window expired
	if time.Since(limit.WindowStart) > limit.WindowDuration {
		limit.Count = 1
		limit.WindowStart = time.Now()
		return true
	}

	// Check limit
	if limit.Count >= AIGenerationPerDay {
		commons.Log(ctx, "[RateLimiter.CheckAIGenerationLimit::Exceeded [user_id=%s [count=%d", userID, limit.Count)
		return false
	}

	limit.Count++
	return true
}

// GetUploadUsage returns user's upload usage
func (rl *RateLimiter) GetUploadUsage(userID string) (int, int) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.uploadLimits[userID]
	if !exists || time.Since(limit.WindowStart) > limit.WindowDuration {
		return 0, UploadLimitPerHour
	}

	return limit.Count, UploadLimitPerHour
}

// GetAIGenUsage returns user's AI generation usage
func (rl *RateLimiter) GetAIGenUsage(userID string) (int, int) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.aiGenLimits[userID]
	if !exists || time.Since(limit.WindowStart) > limit.WindowDuration {
		return 0, AIGenerationPerDay
	}

	return limit.Count, AIGenerationPerDay
}

// cleanupLoop periodically removes expired entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes expired limits
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean upload limits
	for userID, limit := range rl.uploadLimits {
		if now.Sub(limit.WindowStart) > limit.WindowDuration*2 {
			delete(rl.uploadLimits, userID)
		}
	}

	// Clean AI gen limits
	for userID, limit := range rl.aiGenLimits {
		if now.Sub(limit.WindowStart) > limit.WindowDuration*2 {
			delete(rl.aiGenLimits, userID)
		}
	}
}
