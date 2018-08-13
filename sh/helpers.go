package sh

import (
	"fmt"
	"io"
	"os"
)

// Rm removes the given file or directory even if non-empty. It will not return
// an error if the target doesn't exist, only if the target cannot be removed.
func Rm(path string) error {
	err := os.RemoveAll(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf(`failed to remove %s: %v`, path, err)
}

// Copy robustly copies the source file to the destination, overwriting the destination if necessary.
func Copy(frompath string, topath string) error {
	from, err := os.Open(frompath)
	if err != nil {
		return fmt.Errorf(`can't copy %s: %v`, frompath, err)
	}
	defer from.Close()
	to, err := os.Create(topath)
	if err != nil {
		return fmt.Errorf(`can't copy to %s: %v`, topath, err)
	}
	_, err = io.Copy(to, from)
	if err != nil {
		return fmt.Errorf(`error copying %s to %s: %v`, frompath, topath, err)
	}
	return nil
}
