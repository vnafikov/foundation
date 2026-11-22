package server

import (
	"sync"
	"time"
)

type messageLimiterShard struct {
	limiters        map[string]messageLimiter
	deleteExpiredAt time.Time
	mu              sync.Mutex
}

// allow расходует общий бюджет получателя.
// Сохраняет бюджет независимо от подключения и отключения соединений.
func (mls *messageLimiterShard) allow(recipient string, interval, burstInterval time.Duration) bool {
	mls.mu.Lock()
	defer mls.mu.Unlock()

	now := time.Now()
	if !now.Before(mls.deleteExpiredAt) {
		mls.deleteExpired(now)
		mls.deleteExpiredAt = now.Add(burstInterval)
	}
	limiter := mls.limiters[recipient]
	if !limiter.allow(now, interval, burstInterval) {
		return false
	}

	if mls.limiters == nil {
		mls.limiters = make(map[string]messageLimiter)
	}
	mls.limiters[recipient] = limiter
	return true
}

// deleteExpired удаляет записи с полностью восстановленным бюджетом.
func (mls *messageLimiterShard) deleteExpired(now time.Time) {
	for recipient, limiter := range mls.limiters {
		if !limiter.refilledAt.After(now) {
			delete(mls.limiters, recipient)
		}
	}
}
