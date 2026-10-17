package pool

import (
	"sync"
	"testing"
)

var ev = Event{ID: 42, Name: "deploy", Score: 99}

func TestSameOutput(t *testing.T) {
	want := `{"id":42,"name":"deploy","score":99}`
	if RenderNoPool(ev) != want || RenderPooled(ev) != want {
		t.Fatal("renderers disagree")
	}
}

func TestPooledIsConcurrencySafe(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := Event{ID: i, Name: "n", Score: i}
			if RenderPooled(e) != RenderNoPool(e) {
				t.Error("pooled output corrupted under concurrency")
			}
		}()
	}
	wg.Wait()
}

// [start bench]

func BenchmarkRender(b *testing.B) {
	b.Run("impl=nopool", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				RenderNoPool(ev)
			}
		})
	})
	b.Run("impl=pool", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				RenderPooled(ev)
			}
		})
	})
}

// [end bench]
