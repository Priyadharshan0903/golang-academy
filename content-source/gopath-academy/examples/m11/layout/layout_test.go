//go:build amd64 || arm64

package layout

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

// [start example]

func Example_sizes() {
	var w Wasteful
	var c Compact
	fmt.Println("Wasteful:", unsafe.Sizeof(w), "bytes; ID at offset", unsafe.Offsetof(w.ID), "Count at", unsafe.Offsetof(w.Count))
	fmt.Println("Compact:", unsafe.Sizeof(c), "bytes; ID at offset", unsafe.Offsetof(c.ID), "Count at", unsafe.Offsetof(c.Count))
	fmt.Println("alignment of int64, int32, bool:", unsafe.Alignof(w.ID), unsafe.Alignof(w.Count), unsafe.Alignof(w.Active))
	// Output:
	// Wasteful: 32 bytes; ID at offset 8 Count at 20
	// Compact: 16 bytes; ID at offset 0 Count at 8
	// alignment of int64, int32, bool: 8 4 1
}

func Example_headers() {
	var (
		s  string
		sl []int
		a  any
		m  map[string]int
		ch chan int
		f  func()
		p  *int
	)
	fmt.Println("string:", unsafe.Sizeof(s))    // pointer + length
	fmt.Println("slice:", unsafe.Sizeof(sl))    // pointer + length + capacity
	fmt.Println("interface:", unsafe.Sizeof(a)) // type word + data word
	fmt.Println("map, chan, func, pointer:", unsafe.Sizeof(m), unsafe.Sizeof(ch), unsafe.Sizeof(f), unsafe.Sizeof(p))
	fmt.Println("empty struct:", unsafe.Sizeof(struct{}{}))
	// Output:
	// string: 16
	// slice: 24
	// interface: 16
	// map, chan, func, pointer: 8 8 8 8
	// empty struct: 0
}

// [end example]

// [start unsafe]

func Example_unsafeString() {
	b := []byte("hello")
	s := unsafe.String(&b[0], len(b)) // no copy: s shares b's memory

	b[0] = 'j' // breaks the rule that strings never change
	fmt.Println(s)
	// Output: jello
}

// [end unsafe]

// [start bench]

func BenchmarkCounters(b *testing.B) {
	const n = 1_000_000
	run := func(b *testing.B, a, c interface{ Add(int64) int64 }) {
		for b.Loop() {
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				for range n {
					a.Add(1)
				}
			}()
			go func() {
				defer wg.Done()
				for range n {
					c.Add(1)
				}
			}()
			wg.Wait()
		}
	}
	b.Run("layout=adjacent", func(b *testing.B) { var x Counters; run(b, &x.A, &x.B) })
	b.Run("layout=padded", func(b *testing.B) { var x PaddedCounters; run(b, &x.A, &x.B) })
}

// [end bench]
