package sh

import (
	"io"
	"os"

	"github.com/magefile/mage/internal"
)

// Rm removes the given file or directory even if non-empty. It will not return
// an error if the target doesn't exist, only if the target cannot be removed.
func Rm(path string) error {
	err := os.RemoveAll(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return internal.WrapErrorf(err, `failed to remove %s: %v`, path, err)
}

// Copy robustly copies the source file to the destination, overwriting the destination if necessary.
func Copy(dst string, src string) error {
	from, err := os.Open(src)
	if err != nil {
		return internal.WrapErrorf(err, `can't copy %s: %v`, src, err)
	}
	defer from.Close()
	finfo, err := from.Stat()
	if err != nil {
		return internal.WrapErrorf(err, `can't stat %s: %v`, src, err)
	}
	to, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, finfo.Mode())
	if err != nil {
		return internal.WrapErrorf(err, `can't copy to %s: %v`, dst, err)
	}
	defer to.Close()
	_, err = io.Copy(to, from)
	if err != nil {
		return internal.WrapErrorf(err, `error copying %s to %s: %v`, src, dst, err)
	}
	return nil
}
