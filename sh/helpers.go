package sh

import (
	"github.com/magefile/mage/shctx"
)

// Rm removes the given file or directory even if non-empty. It will not return
// an error if the target doesn't exist, only if the target cannot be removed.
func Rm(path string) error {
	return shctx.Rm(path)
}

// Copy robustly copies the source file to the destination, overwriting the destination if necessary.
func Copy(dst, src string) error {
	return shctx.Copy(dst, src)
}
