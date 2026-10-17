// Package pool holds the examples for lesson 10.6.
package pool

import (
	"bytes"
	"strconv"
	"sync"
)

type Event struct {
	ID    int
	Name  string
	Score int
}

// [start nopool]

// RenderNoPool builds each line in a new buffer.
func RenderNoPool(e Event) string {
	var buf bytes.Buffer
	buf.Grow(64)
	writeEvent(&buf, e)
	return buf.String()
}

// [end nopool]

// [start pool]

// bufPool hands out reusable buffers. New is called when the pool is
// empty. Store pointers (*bytes.Buffer), not values, to avoid an extra
// allocation every time an item is put back.
var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// RenderPooled reuses buffers across calls and across goroutines.
func RenderPooled(e Event) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset() // a pooled buffer still has the previous user's data

	writeEvent(buf, e)
	s := buf.String() // copy out BEFORE returning the buffer to the pool

	if buf.Cap() <= 4096 { // do not keep unusually large buffers alive
		bufPool.Put(buf)
	}
	return s
}

// [end pool]

func writeEvent(buf *bytes.Buffer, e Event) {
	buf.WriteString(`{"id":`)
	buf.WriteString(strconv.Itoa(e.ID))
	buf.WriteString(`,"name":"`)
	buf.WriteString(e.Name)
	buf.WriteString(`","score":`)
	buf.WriteString(strconv.Itoa(e.Score))
	buf.WriteByte('}')
}
