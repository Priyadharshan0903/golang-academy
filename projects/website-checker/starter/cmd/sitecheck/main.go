package main

import (
	"bufio"
	"context"
	"encoding/json"
	"example.com/sitecheck/checker"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop() // os.Exit does not run defers; release signal ownership explicitly.
	os.Exit(code)
}

func run(parent context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("sitecheck", flag.ContinueOnError)
	flags.SetOutput(stderr)
	workers := flags.Int("workers", 8, "maximum concurrent requests (1-8)")
	perRequest := flags.Duration("timeout", 2*time.Second, "deadline per request")
	total := flags.Duration("total", 15*time.Second, "deadline for the network checking stage")
	file := flags.String("file", "", "read URLs from a file, or - for stdin")
	asJSON := flags.Bool("json", false, "write a JSON result array")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *total <= 0 {
		fmt.Fprintln(stderr, "total deadline must be positive")
		return 2
	}
	engine, err := checker.New(checker.Config{Workers: *workers, RequestTimeout: *perRequest})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	urls := flags.Args()
	if *file != "" {
		var input io.Reader = stdin
		if *file != "-" {
			opened, err := os.Open(*file)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 2
			}
			defer opened.Close()
			input = opened
		}
		added, err := readURLs(input)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		urls = append(urls, added...)
	}
	if len(urls) == 0 {
		fmt.Fprintln(stderr, "provide URLs as arguments or with -file")
		return 2
	}
	ctx, cancel := context.WithTimeout(parent, *total)
	defer cancel()
	results, runError := engine.Check(ctx, urls)
	var outputError error
	if *asJSON {
		outputError = json.NewEncoder(stdout).Encode(results)
	} else {
		for _, result := range results {
			status := fmt.Sprint(result.StatusCode)
			if result.Error != "" {
				status = "ERROR"
			}
			if _, err := fmt.Fprintf(stdout, "%s\t%s\n", status, result.URL); err != nil {
				outputError = err
				break
			}
			if result.Error != "" {
				fmt.Fprintf(stderr, "%s: %s\n", result.URL, result.Error)
			}
		}
	}
	if outputError != nil {
		fmt.Fprintln(stderr, "write report:", outputError)
		return 1
	}
	if runError != nil {
		fmt.Fprintln(stderr, "check:", runError)
		return 1
	}
	for _, result := range results {
		if result.Error != "" || result.StatusCode < 200 || result.StatusCode >= 300 {
			return 1
		}
	}
	return 0
}

func readURLs(input io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 1024), 64*1024)
	urls := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(urls) >= checker.MaxURLs {
			return nil, fmt.Errorf("input exceeds %d URLs", checker.MaxURLs)
		}
		urls = append(urls, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read URL list: %w", err)
	}
	return urls, nil
}
