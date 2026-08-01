package httpapi

import (
	"sync"
	"time"
)

const maxRateLimitBuckets = 10_000

type rateBucket struct {
	tokens float64
	last   time.Time
}

// rateLimiter is deliberately process-local. Production deployments must also
// enforce an edge/distributed limit; this layer remains useful for single-node
// defense in depth and deterministic tests.
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]rateBucket
	now     func() time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]rateBucket),
		now:     time.Now,
	}
}

func (l *rateLimiter) allow(key string, capacity int, refillEvery time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if capacity <= 0 || refillEvery <= 0 {
		return false
	}

	now := l.now()
	bucket, exists := l.buckets[key]
	if !exists {
		if len(l.buckets) >= maxRateLimitBuckets {
			l.prune(now.Add(-time.Hour))
			if len(l.buckets) >= maxRateLimitBuckets {
				return false
			}
		}
		bucket = rateBucket{tokens: float64(capacity), last: now}
	}
	elapsed := now.Sub(bucket.last)
	if elapsed > 0 {
		bucket.tokens += elapsed.Seconds() / refillEvery.Seconds()
		if bucket.tokens > float64(capacity) {
			bucket.tokens = float64(capacity)
		}
		bucket.last = now
	}
	if bucket.tokens < 1 {
		l.buckets[key] = bucket
		return false
	}
	bucket.tokens--
	l.buckets[key] = bucket
	return true
}

func (l *rateLimiter) prune(before time.Time) {
	for key, bucket := range l.buckets {
		if bucket.last.Before(before) {
			delete(l.buckets, key)
		}
	}
}
