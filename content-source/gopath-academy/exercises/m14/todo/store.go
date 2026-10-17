// Package todo practices writing a command-line tool with subcommands
// and the flag package (lesson 14.1). The tool keeps a to-do list in a
// JSON file.
package todo

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

// Item is one entry of the list.
type Item struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// load reads the list from path. A missing file is an empty list.
func load(path string) ([]Item, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var items []Item
	err = json.Unmarshal(b, &items)
	return items, err
}

// save writes the list to path, atomically: it writes a temporary file
// and renames it, so a crash never leaves a half-written list.
func save(path string, items []Item) error {
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
