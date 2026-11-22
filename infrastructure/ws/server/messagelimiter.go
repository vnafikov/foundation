package server

import "time"

type messageLimiter struct {
	refilledAt time.Time
}

// allow расходует бюджет, восстанавливаемый на одно сообщение за interval.
func (ml *messageLimiter) allow(now time.Time, interval, burstInterval time.Duration) bool {
	refilledAt := ml.refilledAt
	if refilledAt.Before(now) {
		refilledAt = now
	}
	refilledAt = refilledAt.Add(interval)
	if refilledAt.Sub(now) > burstInterval {
		return false
	}

	ml.refilledAt = refilledAt
	return true
}
