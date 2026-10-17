package syncprims

import (
	"fmt"
	"sort"
	"sync"
)

// [start mutex]

// Inventory is safe for concurrent use. The mutex guards items;
// every method that touches items locks it first.
type Inventory struct {
	mu    sync.Mutex
	items map[string]int
}

func NewInventory() *Inventory { return &Inventory{items: map[string]int{}} }

func (inv *Inventory) Add(name string, n int) {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.items[name] += n
}

func (inv *Inventory) Count(name string) int {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	return inv.items[name]
}

func Example_mutex() {
	inv := NewInventory()
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			inv.Add("widget", 1)
		}()
	}
	wg.Wait()
	fmt.Println(inv.Count("widget"))

	// Output:
	// 100
}

// [end mutex]

// [start rwmutex]

// Cache allows many concurrent readers, or one writer.
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(k string) (string, bool) {
	c.mu.RLock() // shared lock: readers do not block each other
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *Cache) Set(k, v string) {
	c.mu.Lock() // exclusive lock: waits for all readers to finish
	defer c.mu.Unlock()
	if c.data == nil {
		c.data = map[string]string{}
	}
	c.data[k] = v
}

func Example_rwMutex() {
	var c Cache // zero value is usable: Set creates the map lazily
	c.Set("region", "ap-south-1")

	var wg sync.WaitGroup
	hits := make([]bool, 50)
	for i := range hits {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, hits[i] = c.Get("region")
		}()
	}
	wg.Wait()
	fmt.Println(hits[0], hits[49])

	// Output:
	// true true
}

// [end rwmutex]

// [start once]

type Config struct{ Env string }

var (
	loadCount int
	getConfig = sync.OnceValue(func() *Config { // Go 1.21+
		loadCount++ // runs exactly once, even with concurrent callers
		return &Config{Env: "production"}
	})
)

func Example_onceValue() {
	var wg sync.WaitGroup
	envs := make([]string, 20)
	for i := range envs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			envs[i] = getConfig().Env
		}()
	}
	wg.Wait()
	sort.Strings(envs)
	fmt.Println(envs[0], loadCount)

	// Output:
	// production 1
}

// [end once]

// [start waitgroup]

func Example_collectWithMutex() {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		squares []int
	)
	for i := 1; i <= 5; i++ {
		wg.Add(1) // Add BEFORE starting the goroutine, never inside it
		go func() {
			defer wg.Done()
			sq := i * i
			mu.Lock()
			squares = append(squares, sq)
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Ints(squares) // goroutines finish in any order
	fmt.Println(squares)

	// Output:
	// [1 4 9 16 25]
}

// [end waitgroup]
