package ratelimiter_test

import (
	"testing"
	"time"
	"warehouse-control/internal/http-server/ratelimiter"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(5, 5)

	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow("127.0.0.1"), "Request %d should be allowed", i+1)
	}

	assert.False(t, rl.Allow("127.0.0.1"), "Sixth request should be denied")

	time.Sleep(1100 * time.Millisecond)
	assert.True(t, rl.Allow("127.0.0.1"), "Request after wait should be allowed")
	assert.True(t, rl.Allow("127.0.0.1"), "Second request after wait should be allowed")
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(2, 2)

	assert.True(t, rl.Allow("127.0.0.1"))
	assert.True(t, rl.Allow("127.0.0.1"))
	assert.False(t, rl.Allow("127.0.0.1"))

	assert.True(t, rl.Allow("192.168.1.1"))
	assert.True(t, rl.Allow("192.168.1.1"))
	assert.False(t, rl.Allow("192.168.1.1"))
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(1, 10)

	for i := 0; i < 15000; i++ {
		ip := "192.168.1." + string('0'+rune(i%10))
		rl.Visitors[ip] = &ratelimiter.Visitor{
			LastSeen: time.Now().Add(-11 * time.Minute),
			Tokens:   0,
		}
	}

	rl.Allow("127.0.0.1")
	assert.Less(t, len(rl.Visitors), 10000, "Should clean up old visitors")
}
