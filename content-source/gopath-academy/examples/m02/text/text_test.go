package text

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// [start bytes]

func Example_bytesVsRunes() {
	s := "héllo, 世界"

	fmt.Println(len(s))                    // bytes, not characters
	fmt.Println(utf8.RuneCountInString(s)) // characters (runes)
	fmt.Println(s[1])                      // indexing gives a byte
	fmt.Println(s[0:1])                    // slicing works on bytes too

	// Output:
	// 14
	// 9
	// 195
	// h
}

// [end bytes]

// [start range]

func Example_rangeOverString() {
	for i, r := range "aé世" {
		fmt.Printf("byte offset %d: %c (U+%04X, %d bytes)\n", i, r, r, utf8.RuneLen(r))
	}

	// Output:
	// byte offset 0: a (U+0061, 1 bytes)
	// byte offset 1: é (U+00E9, 2 bytes)
	// byte offset 3: 世 (U+4E16, 3 bytes)
}

// [end range]

// [start convert]

func Example_conversions() {
	s := "Go!"
	b := []byte(s) // copies the bytes
	b[0] = 'N'
	fmt.Println(s, string(b)) // s is unchanged: strings are immutable

	r := []rune("世界")
	fmt.Println(len(r), string(r[1]))

	// Output:
	// Go! No!
	// 2 界
}

// [end convert]

// [start builder]

func Example_builder() {
	var sb strings.Builder
	for i := range 3 {
		fmt.Fprintf(&sb, "line %d\n", i)
	}
	fmt.Print(sb.String())

	fmt.Println(strings.Join([]string{"a", "b", "c"}, "-"))
	fmt.Println(strings.Fields("  split   on spaces "))
	fmt.Println(strings.EqualFold("Go", "GO"))

	// Output:
	// line 0
	// line 1
	// line 2
	// a-b-c
	// [split on spaces]
	// true
}

// [end builder]

// [start raw]

func Example_rawStrings() {
	path := `C:\Users\ana` // backslashes are literal
	multi := `line one
line two`
	fmt.Println(path)
	fmt.Println(multi)

	// Output:
	// C:\Users\ana
	// line one
	// line two
}

// [end raw]
