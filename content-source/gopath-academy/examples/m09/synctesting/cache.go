//go:build go1.25

package synctesting

import (
	"sync"
	"time"
)

// [start code]

// TTLCache stores values that expire after ttl. A background goroutine
// removes expired entries every sweep interval.
type TTLCache struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[string]item
	stop  chan struct{}
	done  chan struct{}
}

type item struct {
	val     string
	expires time.Time
}

func NewTTLCache(ttl, sweep time.Duration) *TTLCache {
	c := &TTLCache{ttl: ttl, items: map[string]item{}, stop: make(chan struct{}), done: make(chan struct{})}
	go c.sweeper(sweep)
	return c
}

func (c *TTLCache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = item{v, time.Now().Add(c.ttl)}
}

func (c *TTLCache) Get(k string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[k]
	if !ok || time.Now().After(it.expires) {
		return "", false
	}
	return it.val, true
}

func (c *TTLCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func (c *TTLCache) Close() { close(c.stop); <-c.done }

func (c *TTLCache) sweeper(every time.Duration) {
	defer close(c.done)
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-c.stop:
			return
		case now := <-t.C:
			c.mu.Lock()
			for k, it := range c.items {
				if now.After(it.expires) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// [end code]
