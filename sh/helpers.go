package sh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// Copy robustly copies the source file to the destination, overwriting the
// destination if necessary.
func Copy(dst, src string) error {
	finfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("can't stat %s: %v", src, err)
	}

	if err = copyFile(dst, src, finfo); err != nil {
		return fmt.Errorf("can't copy %s to %s: %v", src, dst, err)
	}

	return nil
}

// CopyDir copies the source directory content to given destination.
// Destination  must exist.
func CopyDir(dst, src string) error {
	fn := func(srcpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if srcpath == src {
			return filepath.SkipDir
		}

		dst := filepath.Clean(strings.Replace(srcpath, src, dst, 1))

		if fi.IsDir() {
			mkerr := os.Mkdir(dst, fi.Mode())
			if os.IsExist(mkerr) {
				return os.Chmod(dst, fi.Mode())
			}

			return mkerr
		}

		return copyFile(dst, srcpath, fi)
	}

	if err := filepath.Walk(src, fn); err != nil {
		return fmt.Errorf("can't copy %s to %s: %v", src, dst, err)
	}

	return nil
}

// CopyDirAll copies the source directory content to given destination. If
// destination doesn't exist, it will be created
func CopyDirAll(dst, src string) error {
	finfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("can't stat %s: %v", src, err)
	}

	if err := os.MkdirAll(dst, finfo.Mode()); err != nil {
		return fmt.Errorf("can't create %s: %v", dst, err)
	}

	return CopyDir(dst, src)
}

// CopyAny copies the source elements to the given destination. Destination
// must be a directory.
func CopyAny(dst string, src ...string) error {
	dfi, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("can't stat %s: %v", dst, err)
	}

	if !dfi.IsDir() {
		return fmt.Errorf("can't use %s as destination", dst)
	}

	for _, el := range src {
		sfi, elerr := os.Stat(el)
		if elerr != nil {
			return fmt.Errorf("can't stat %s: %v", el, elerr)
		}

		dst := filepath.Join(dst, sfi.Name())

		if sfi.IsDir() {
			if elerr = CopyDirAll(dst, el); elerr != nil {
				return elerr
			}

			continue
		}

		if elerr = copyFile(dst, el, sfi); elerr != nil {
			return elerr
		}
	}

	return nil
}

func copyFile(dst, src string, sfi os.FileInfo) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}

	defer from.Close()

	to, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sfi.Mode())

	if err != nil {
		return err
	}

	defer to.Close()

	if _, err = io.Copy(to, from); err != nil {
		return err
	}

	return nil
}
