// Package files holds the tested examples for lesson 8.5.
package files

import (
	"embed"
	"os"
	"path/filepath"
)

// [start embed]

// The go:embed directive copies files into the binary at build time.
// No file needs to exist next to the program when it runs.

//go:embed assets/greeting.txt
var greeting string

//go:embed assets
var assetsFS embed.FS

// [end embed]

// [start atomic]

// WriteFileAtomic writes data so that readers see either the old file
// or the complete new one, never a half-written file: write to a temp
// file in the same directory, flush it to disk, then rename over the
// target (rename is atomic on POSIX filesystems).
func WriteFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(tmp.Name()) // clean up on any failure
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil { // make sure the bytes reached the disk
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmp.Name(), perm); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// [end atomic]
