package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"
)

// [start readwrite]

func Example_readWrite() {
	dir, _ := os.MkdirTemp("", "files-example")
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "notes", "today.txt") // Join uses the OS separator
	os.MkdirAll(filepath.Dir(path), 0o755)

	if err := os.WriteFile(path, []byte("buy milk\n"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	data, _ := os.ReadFile(path)
	fmt.Print(string(data))

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	fmt.Fprintln(f, "call mom")
	f.Close()

	data, _ = os.ReadFile(path)
	fmt.Print(string(data))

	_, err := os.ReadFile(filepath.Join(dir, "missing.txt"))
	fmt.Println(errors.Is(err, fs.ErrNotExist))

	// Output:
	// buy milk
	// buy milk
	// call mom
	// true
}

// [end readwrite]

// [start atomicuse]

func Example_atomicWrite() {
	dir, _ := os.MkdirTemp("", "atomic")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "config.json")

	WriteFileAtomic(path, []byte(`{"v":1}`), 0o644)
	WriteFileAtomic(path, []byte(`{"v":2}`), 0o644)

	data, _ := os.ReadFile(path)
	entries, _ := os.ReadDir(dir)
	fmt.Println(string(data), len(entries)) // no temp files left behind

	// Output:
	// {"v":2} 1
}

// [end atomicuse]

// [start walk]

func Example_walkDir() {
	dir, _ := os.MkdirTemp("", "walk")
	defer os.RemoveAll(dir)
	for _, p := range []string{"a.go", "sub/b.go", "sub/c.txt", "sub/deep/d.go"} {
		full := filepath.Join(dir, p)
		os.MkdirAll(filepath.Dir(full), 0o755)
		os.WriteFile(full, nil, 0o644)
	}

	var goFiles []string
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			rel, _ := filepath.Rel(dir, path)
			goFiles = append(goFiles, filepath.ToSlash(rel))
		}
		return nil
	})
	fmt.Println(goFiles) // WalkDir visits entries in lexical order

	// Output:
	// [a.go sub/b.go sub/deep/d.go]
}

// [end walk]

// [start fsabstraction]

// countLines works on ANY fs.FS: a real directory, embedded files, a
// zip archive or an in-memory test filesystem.
func countLines(fsys fs.FS, pattern string) (int, error) {
	matches, err := fs.Glob(fsys, pattern)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, name := range matches {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return 0, err
		}
		total += strings.Count(string(data), "\n")
	}
	return total, nil
}

func Example_fsInterface() {
	testFS := fstest.MapFS{ // an in-memory fs.FS, perfect for tests
		"a.txt":     {Data: []byte("1\n2\n")},
		"b.txt":     {Data: []byte("3\n")},
		"image.png": {Data: []byte{0x89}},
	}
	n, _ := countLines(testFS, "*.txt")
	fmt.Println(n)

	n, _ = countLines(assetsFS, "assets/*.txt") // the embedded files
	fmt.Println(n)

	// Output:
	// 3
	// 1
}

// [end fsabstraction]

// [start embeduse]

func Example_embed() {
	fmt.Print(greeting)

	entries, _ := assetsFS.ReadDir("assets")
	for _, e := range entries {
		fmt.Println(e.Name())
	}

	// Output:
	// Hello from an embedded file.
	// greeting.txt
}

// [end embeduse]

// [start root]

func Example_osRoot() {
	dir, _ := os.MkdirTemp("", "root")
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "allowed.txt"), []byte("ok"), 0o644)

	// os.Root (Go 1.24+) confines every operation to one directory.
	// Paths that try to escape it fail, even through ../ or symlinks.
	root, err := os.OpenRoot(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer root.Close()

	f, err := root.Open("allowed.txt")
	fmt.Println(err == nil)
	if f != nil {
		f.Close()
	}

	_, err = root.Open("../../etc/passwd")
	fmt.Println(err != nil)

	// Output:
	// true
	// true
}

// [end root]
