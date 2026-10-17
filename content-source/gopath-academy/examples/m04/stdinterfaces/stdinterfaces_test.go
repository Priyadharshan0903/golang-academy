package stdinterfaces

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// [start stringer]

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
)

// String makes Weekday a fmt.Stringer, so fmt prints names, not numbers.
func (d Weekday) String() string {
	switch d {
	case Sunday:
		return "Sunday"
	case Monday:
		return "Monday"
	case Tuesday:
		return "Tuesday"
	default:
		return fmt.Sprintf("Weekday(%d)", int(d))
	}
}

func Example_stringer() {
	fmt.Println(Monday)
	fmt.Printf("%v and %d\n", Tuesday, Tuesday) // %d still prints the number
	fmt.Println(Weekday(9))

	// Output:
	// Monday
	// Tuesday and 2
	// Weekday(9)
}

// [end stringer]

// [start writer]

// CountingWriter wraps any io.Writer and counts the bytes written.
type CountingWriter struct {
	W     io.Writer
	Bytes int64
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.W.Write(p)
	c.Bytes += int64(n)
	return n, err
}

func Example_writer() {
	cw := &CountingWriter{W: os.Stdout}
	fmt.Fprintf(cw, "hello, %s\n", "gopher") // anything that writes to an io.Writer works
	io.WriteString(cw, "second line\n")
	fmt.Println("bytes:", cw.Bytes)

	// Output:
	// hello, gopher
	// second line
	// bytes: 26
}

// [end writer]

// [start reader]

// upperReader wraps an io.Reader and upper-cases ASCII letters.
type upperReader struct{ r io.Reader }

func (u upperReader) Read(p []byte) (int, error) {
	n, err := u.r.Read(p) // fill p from the inner reader
	for i := range n {
		if 'a' <= p[i] && p[i] <= 'z' {
			p[i] -= 'a' - 'A'
		}
	}
	return n, err // pass io.EOF through unchanged
}

func Example_reader() {
	r := upperReader{strings.NewReader("streams compose nicely")}
	io.Copy(os.Stdout, r) // io.Copy only needs a Reader and a Writer
	fmt.Println()

	// Output:
	// STREAMS COMPOSE NICELY
}

// [end reader]

// [start sort]

// byLength implements sort.Interface: Len, Less and Swap.
type byLength []string

func (s byLength) Len() int           { return len(s) }
func (s byLength) Less(i, j int) bool { return len(s[i]) < len(s[j]) }
func (s byLength) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func Example_sortInterface() {
	words := []string{"banana", "fig", "apple", "kiwi"}
	sort.Stable(byLength(words))
	fmt.Println(words)

	// Output:
	// [fig kiwi apple banana]
}

// [end sort]
