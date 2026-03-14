package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu         sync.Mutex
	Visitors   map[string]*Visitor
	RatePerSec int
	Capacity   int
}

type Visitor struct {
	LastSeen time.Time
	Tokens   float64
}

func NewRateLimiter(ratePerSec, capacity int) *RateLimiter {
	return &RateLimiter{
		Visitors:   make(map[string]*Visitor),
		RatePerSec: ratePerSec,
		Capacity:   capacity,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	visitor, exists := rl.Visitors[ip]

	if len(rl.Visitors) > 10000 {
		rl.cleanupOldVisitors(now)
	}

	if !exists {
		rl.Visitors[ip] = &Visitor{
			LastSeen: now,
			Tokens:   float64(rl.Capacity - 1),
		}
		return true
	}

	elapsed := now.Sub(visitor.LastSeen).Seconds()
	newTokens := visitor.Tokens + (elapsed * float64(rl.RatePerSec))

	if newTokens > float64(rl.Capacity) {
		newTokens = float64(rl.Capacity)
	}

	visitor.LastSeen = now

	if newTokens >= 1.0 {
		visitor.Tokens = newTokens - 1.0
		return true
	}

	visitor.Tokens = newTokens
	return false
}

func (rl *RateLimiter) cleanupOldVisitors(now time.Time) {
	for ip, visitor := range rl.Visitors {
		if now.Sub(visitor.LastSeen) > 10*time.Minute {
			delete(rl.Visitors, ip)
		}
	}
}
