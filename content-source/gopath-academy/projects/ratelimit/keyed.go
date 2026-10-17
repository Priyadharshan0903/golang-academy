package ratelimit

import (
	"sync"
	"time"
)

// Keyed keeps one Bucket per key, for example per client IP or API key.
// It is safe for concurrent use.
type Keyed struct {
	mu      sync.Mutex
	rate    float64
	burst   int
	buckets map[string]*entry
	now     func() time.Time
}

type entry struct {
	bucket   *Bucket
	lastSeen time.Time
}

// NewKeyed returns a Keyed limiter whose buckets all use rate and burst.
func NewKeyed(rate float64, burst int) *Keyed {
	return &Keyed{rate: rate, burst: burst, buckets: make(map[string]*entry), now: time.Now}
}

// Allow reports whether one event for key may happen now.
func (k *Keyed) Allow(key string) bool {
	return k.bucket(key).Allow()
}

func (k *Keyed) bucket(key string) *Bucket {
	k.mu.Lock()
	defer k.mu.Unlock()
	e, ok := k.buckets[key]
	if !ok {
		b := New(k.rate, k.burst)
		b.now = k.now
		b.last = k.now()
		e = &entry{bucket: b}
		k.buckets[key] = e
	}
	e.lastSeen = k.now()
	return e.bucket
}

// Prune removes buckets not used for longer than idle, and returns how
// many it removed. Call it periodically, or memory grows with every
// key ever seen.
func (k *Keyed) Prune(idle time.Duration) int {
	k.mu.Lock()
	defer k.mu.Unlock()
	cutoff := k.now().Add(-idle)
	removed := 0
	for key, e := range k.buckets {
		if e.lastSeen.Before(cutoff) {
			delete(k.buckets, key)
			removed++
		}
	}
	return removed
}

// Len returns the number of keys currently tracked.
func (k *Keyed) Len() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.buckets)
}
