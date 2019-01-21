package sh

import (
	"errors"
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

// Copy robustly copies the source file to the destination, overwriting the destination if necessary.
func Copy(dst string, src string) error {
	from, err := os.Open(src)
	if err != nil {
		return fmt.Errorf(`can't copy %s: %v`, src, err)
	}
	defer from.Close()
	finfo, err := from.Stat()
	if err != nil {
		return fmt.Errorf(`can't stat %s: %v`, src, err)
	}
	to, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, finfo.Mode())
	if err != nil {
		return fmt.Errorf(`can't copy to %s: %v`, dst, err)
	}
	_, err = io.Copy(to, from)
	if err != nil {
		return fmt.Errorf(`error copying %s to %s: %v`, src, dst, err)
	}
	return nil
}

// CopyDir copies recursively the source directory content to the destination,
// overwriting files in the destination if necessary. Errors will be wrapped
// into a single error.
func CopyDir(dst, src string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sfi.IsDir() {
		return fmt.Errorf("cannot use file %v as source", src)
	}

	srcabs, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	dstabs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	if strings.HasPrefix(dstabs, srcabs) {
		return fmt.Errorf("cannot copy directory %v into itself", dst)
	}

	dfi, err := os.Stat(dst)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dst, sfi.Mode()); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !dfi.IsDir() {
		msg := "cannot overwrite non-directory %v with directory %v"
		return fmt.Errorf(msg, dst, src)
	}

	return copyDir(dst, src)
}

// Cp behaves like the Unix command cp with -rf flags. The possible scenarios
// are:
//
// * (FAIL) No source.
//
// * (FAIL) Non existent source.
//
// * (PASS) File to non existent file.
//
// * (FAIL) File to more than one non existent elements in destination.
//
// * (PASS) File to file.
//
// * (PASS) File to directory.
//
// * (FAIL) Directory to file.
//
// * (PASS) Directory to non existent directory.
//
// * (FAIL) Directory to more than one non existent elements in destination.
//
// * (PASS) Directory to directory.
//
// * (FAIL) Directory to itself.
//
// * (FAIL) Multiple elements to file.
//
// * (PASS) Multiple elements to directory.
//
// * (FAIL) Bad permissions.
func Cp(dst string, src ...string) error {
	dfi, err := os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	n := len(src)

	switch {
	case n == 0:
		return errors.New("missing source")
	case n > 1 && dfi == nil:
		return fmt.Errorf("%v doesn't exists", dst)
	case n > 1 && !dfi.IsDir():
		return fmt.Errorf("%v is not a directory", dst)
	}

	var errs []string
	var fn func(string, string) error

	for _, entry := range src {
		dst := dst

		sfi, err := os.Stat(entry)
		if err != nil {
			return err
		}

		if dfi != nil && dfi.IsDir() {
			dst = filepath.Join(dst, sfi.Name())
		}

		if sfi.IsDir() {
			fn = CopyDir
		} else {
			fn = Copy
		}

		if err := fn(dst, entry); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func copyDir(dst, src string) error {
	var errs []string

	fn := func(srcpath string, fi os.FileInfo, err error) error {
		if src == srcpath {
			return nil
		}

		if err != nil {
			errs = append(errs, err.Error())
			return nil
		}

		dst := filepath.Clean(strings.Replace(srcpath, src, dst, 1))

		if fi.IsDir() {
			if err := os.Mkdir(dst, fi.Mode()); err != nil {
				errs = append(errs, err.Error())
				return nil
			}
		} else {
			if err := Copy(dst, srcpath); err != nil {
				errs = append(errs, err.Error())
				return nil
			}
		}

		return nil
	}

	if err := filepath.Walk(src, fn); err != nil {
		return err
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}
