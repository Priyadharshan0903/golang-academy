//go:build go1.26

package slogging

import (
	"bytes"
	"fmt"
	"log/slog"
)

// [start multi]

func Example_multiHandler() {
	var human, machine bytes.Buffer
	drop := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}

	// NewMultiHandler (Go 1.26+) sends each record to several handlers.
	log := slog.New(slog.NewMultiHandler(
		slog.NewTextHandler(&human, &slog.HandlerOptions{ReplaceAttr: drop}),
		slog.NewJSONHandler(&machine, &slog.HandlerOptions{ReplaceAttr: drop}),
	))
	log.Info("deployed", "version", "1.4.0")

	fmt.Print(human.String())
	fmt.Print(machine.String())

	// Output:
	// level=INFO msg=deployed version=1.4.0
	// {"level":"INFO","msg":"deployed","version":"1.4.0"}
}

// [end multi]
