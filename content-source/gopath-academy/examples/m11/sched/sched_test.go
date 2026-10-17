package sched

import (
	"crypto/sha256"
	"testing"
)

func TestPreemption(t *testing.T) {
	d := PreemptionDemo()
	t.Logf("main goroutine ran again after %v", d)
}

func TestSchedSnapshot(t *testing.T) {
	s := SchedSnapshot()
	if s["/sched/gomaxprocs:threads"] < 1 || s["/sched/goroutines:goroutines"] < 1 {
		t.Fatalf("unexpected metrics: %v", s)
	}
	for k, v := range s {
		t.Logf("%s = %d", k, v)
	}
}

func makeBlocks() [][]byte {
	blocks := make([][]byte, 64)
	for i := range blocks {
		blocks[i] = make([]byte, 64<<10)
		blocks[i][0] = byte(i)
	}
	return blocks
}

func TestHashAll(t *testing.T) {
	blocks := makeBlocks()
	got := HashAll(blocks, 8)
	for i, b := range blocks {
		if got[i] != sha256.Sum256(b) {
			t.Fatalf("block %d: wrong hash", i)
		}
	}
}

// [start bench]

// Run with: go test -run xxx -bench HashAll -cpu 1,2,4 ./examples/m11/sched
// The -cpu flag reruns the benchmark with each GOMAXPROCS value.
func BenchmarkHashAll(b *testing.B) {
	blocks := makeBlocks()
	b.SetBytes(int64(len(blocks) * len(blocks[0])))
	for b.Loop() {
		HashAll(blocks, 8)
	}
}

// [end bench]
