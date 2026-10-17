package formatting

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

type point struct{ X, Y int }

// [start verbs]

func Example_verbs() {
	p := point{1, 2}
	fmt.Printf("%v | %+v | %#v | %T\n", p, p, p, p)

	fmt.Printf("%d | %5d | %-5d| %05d | %x | %b\n", 42, 42, 42, 42, 255, 5)
	fmt.Printf("%f | %.2f | %8.3f | %e | %g\n", 3.14159, 3.14159, 3.14159, 1234.5678, 0.000012)
	fmt.Printf("%s | %q | %10s| %-10s|\n", "go", "go", "right", "left")
	fmt.Printf("%t | %c | %U | %x\n", true, 'G', 'G', "hi")
	fmt.Printf("%*d | %v%%\n", 6, 7, 50) // width from an argument; %% is a literal percent

	// Output:
	// {1 2} | {X:1 Y:2} | formatting.point{X:1, Y:2} | formatting.point
	// 42 |    42 | 42   | 00042 | ff | 101
	// 3.141590 | 3.14 |    3.142 | 1.234568e+03 | 1.2e-05
	// go | "go" |      right| left      |
	// true | G | U+0047 | 6869
	//      7 | 50%
}

// [end verbs]

// [start strconv]

func Example_strconv() {
	fmt.Println(strconv.Quote("tab\there"), strconv.FormatInt(255, 2))

	f, _ := strconv.ParseFloat("2.5e3", 64)
	b, _ := strconv.ParseBool("true")
	fmt.Println(f, b)

	// Append functions build output without allocating intermediate strings.
	buf := []byte("id=")
	buf = strconv.AppendInt(buf, 42, 10)
	buf = append(buf, ';')
	fmt.Println(string(buf))

	// Output:
	// "tab\there" 11111111
	// 2500 true
	// id=42;
}

// [end strconv]

// [start strings]

func Example_stringsHelpers() {
	r := strings.NewReplacer("<", "&lt;", ">", "&gt;") // many replacements in one pass
	fmt.Println(r.Replace("<b>bold</b>"))

	key, value, found := strings.Cut("timeout=30s", "=")
	fmt.Println(key, value, found)

	fmt.Println(strings.EqualFold("Straße", "STRASSE"), strings.EqualFold("Go", "GO"))
	fmt.Println(strings.Map(func(r rune) rune {
		if r == 'a' {
			return -1 // drop the rune
		}
		return r
	}, "banana"))

	// Output:
	// &lt;b&gt;bold&lt;/b&gt;
	// timeout 30s true
	// false true
	// bnn
}

// [end strings]

// [start tabwriter]

func Example_tabwriter() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tAGE")
	fmt.Fprintln(w, "api-7f9c\tRunning\t3d")
	fmt.Fprintln(w, "worker-1\tCrashLoopBackOff\t12m")
	w.Flush()

	// Output:
	// NAME      STATUS            AGE
	// api-7f9c  Running           3d
	// worker-1  CrashLoopBackOff  12m
}

// [end tabwriter]

// [start append]

func Example_appendf() {
	// fmt.Appendf (Go 1.19+) formats into an existing byte slice.
	line := fmt.Appendf(nil, "%s=%d", "retries", 3)
	line = fmt.Append(line, " ok")
	fmt.Println(string(line))

	// Output:
	// retries=3 ok
}

// [end append]
