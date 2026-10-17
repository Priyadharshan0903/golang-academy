package iostreams

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

// [start copy]

func Example_copy() {
	src := strings.NewReader("streams of bytes\n")
	n, err := io.Copy(os.Stdout, src) // any Reader to any Writer
	fmt.Println(n, err)

	all, _ := io.ReadAll(strings.NewReader("small input")) // only for data you know is small
	fmt.Println(len(all))

	// Output:
	// streams of bytes
	// 17 <nil>
	// 11
}

// [end copy]

// [start tee]

func Example_teeAndMultiWriter() {
	// Hash a stream while copying it, in one pass.
	var saved bytes.Buffer
	h := sha256.New()
	src := io.TeeReader(strings.NewReader("hello"), h) // everything read is also written to h

	io.Copy(&saved, src)
	fmt.Println(saved.String())
	fmt.Printf("%x\n", h.Sum(nil)[:4])

	// Write once, deliver to several writers.
	var a, b bytes.Buffer
	w := io.MultiWriter(&a, &b)
	fmt.Fprint(w, "same bytes")
	fmt.Println(a.String() == b.String())

	// Output:
	// hello
	// 2cf24dba
	// true
}

// [end tee]

// [start limit]

func Example_limitAndMulti() {
	r := io.LimitReader(strings.NewReader("0123456789"), 4) // stop after 4 bytes
	b, _ := io.ReadAll(r)
	fmt.Println(string(b))

	joined := io.MultiReader(strings.NewReader("head-"), strings.NewReader("tail"))
	b, _ = io.ReadAll(joined)
	fmt.Println(string(b))

	// Output:
	// 0123
	// head-tail
}

// [end limit]

// [start scanner]

func Example_scanner() {
	input := "first line\nsecond line\n\nlast line"
	sc := bufio.NewScanner(strings.NewReader(input))
	lines := 0
	for sc.Scan() {
		lines++
	}
	if err := sc.Err(); err != nil { // always check after the loop
		fmt.Println("scan error:", err)
	}
	fmt.Println("lines:", lines)

	words := bufio.NewScanner(strings.NewReader("count  these\twords\n please"))
	words.Split(bufio.ScanWords) // tokens are words instead of lines
	n := 0
	for words.Scan() {
		n++
	}
	fmt.Println("words:", n)

	// Output:
	// lines: 4
	// words: 4
}

// [end scanner]

// [start bufwriter]

func Example_bufferedWriter() {
	var out bytes.Buffer
	w := bufio.NewWriter(&out)
	fmt.Fprint(w, "buffered")

	fmt.Printf("before Flush: %q\n", out.String()) // still in w's buffer
	if err := w.Flush(); err != nil {              // forgetting Flush loses data
		fmt.Println(err)
	}
	fmt.Printf("after Flush: %q\n", out.String())

	// Output:
	// before Flush: ""
	// after Flush: "buffered"
}

// [end bufwriter]

// [start pipe]

func Example_pipe() {
	pr, pw := io.Pipe() // synchronous in-memory pipe: Writer side feeds Reader side

	go func() {
		defer pw.Close() // the reader sees io.EOF after Close
		for i := range 3 {
			fmt.Fprintf(pw, "chunk %d\n", i)
		}
	}()

	sc := bufio.NewScanner(pr)
	for sc.Scan() {
		fmt.Println("got", sc.Text())
	}

	// Output:
	// got chunk 0
	// got chunk 1
	// got chunk 2
}

// [end pipe]
