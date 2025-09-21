package rpc

import (
	"sync"
	"time"
)

type tokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{buckets: make(map[string]*tokenBucket)}
}

func (l *RateLimiter) Allow(key string, capacity int, refillPerMinute int) bool {
	if capacity <= 0 || refillPerMinute <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	bucket, ok := l.buckets[key]
	refillPerSec := float64(refillPerMinute) / 60.0
	if !ok {
		bucket = &tokenBucket{
			capacity:   float64(capacity),
			tokens:     float64(capacity) - 1, // consume one immediately
			refillRate: refillPerSec,
			lastRefill: now,
		}
		l.buckets[key] = bucket
		return true
	}

	// Refill tokens based on elapsed time
	delta := now.Sub(bucket.lastRefill).Seconds()
	if delta > 0 {
		bucket.tokens += delta * bucket.refillRate
		if bucket.tokens > bucket.capacity {
			bucket.tokens = bucket.capacity
		}
		bucket.lastRefill = now
	}

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}
	return false
}

var defaultRateLimiter = NewRateLimiter()
