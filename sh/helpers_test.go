package sh_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magefile/mage/sh"
)


func TestHelpers(t *testing.T) {
	mytmpdir, err := ioutil.TempDir("", "mage")
	if err != nil {
		t.Fatalf("can't create test directory: %v", err)
	}

	defer func() {
		derr := os.RemoveAll(mytmpdir)
		if derr != nil {
			fmt.Printf("error cleaning up after TestHelpers: %v", derr)
		}
	}()

	srcname := filepath.Join(mytmpdir, "test1.txt")
	content := []byte("All work and no play makes Jack a dull boy.")
	err = ioutil.WriteFile(srcname, content, 0644)
	if err != nil {
		t.Fatalf("can't create test file %s: %v", srcname, err)
	}

	destname := filepath.Join(mytmpdir, "test2.txt")

	t.Run("sh/copy", func(t *testing.T) {
		cerr := sh.Copy(destname, srcname)
		if cerr != nil {
			msg := "test file copy from %s to %s failed: %v"
			t.Errorf(msg, srcname, destname, cerr)
		}

		cerr = compareFiles(destname, srcname)
		if cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}
	})

	t.Run("sh/copy/directory", func(t *testing.T) {
		if cerr := sh.Copy(destname, mytmpdir); cerr == nil {
			t.Error("sh.Copy succeeded copying from forbidden source")
		}
	})

	t.Run("sh/copy/forbidden", func(t *testing.T) {
		forbidden := filepath.Join(mytmpdir, "forbidden.txt")
		content := []byte("You are not prepared!")
		err = ioutil.WriteFile(forbidden, content, 0000)
		if err != nil {
			t.Fatalf("can't create test file %s: %v", forbidden, err)
		}

		allowed := filepath.Join(mytmpdir, "allowed.txt")
		if cerr := sh.Copy(allowed, forbidden); cerr == nil {
			t.Error("sh.Copy succeeded copying from forbidden source")
		}

		if cerr := sh.Copy(forbidden, srcname); cerr == nil {
			t.Error("sh.Copy succeeded copying to forbidden destination")
		}
	})

	// While we've got a temporary directory, test how forgiving sh.Rm is
	t.Run("sh/rm/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "file_not_exist.txt")
		rerr := sh.Rm(nef)
		if rerr != nil {
			msg := "sh.Rm complained when removing nonexistent file %s: %v"
			t.Errorf(msg, nef, rerr)
		}
	})

	t.Run("sh/copy/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "file_not_exist.txt")
		nedf := filepath.Join(mytmpdir, "file_not_exist2.txt")
		cerr := sh.Copy(nedf, nef)
		if cerr == nil {
			t.Errorf("sh.Copy succeeded copying nonexistent file %s", nef)
		}
	})

	// We test sh.Rm by clearing up our own test files and directories
	t.Run("sh/rm", func(t *testing.T) {
		rerr := sh.Rm(destname)
		if rerr != nil {
			t.Errorf("failed to remove file %s: %v", destname, rerr)
		}

		rerr = sh.Rm(srcname)
		if rerr != nil {
			t.Errorf("failed to remove file %s: %v", srcname, rerr)
		}

		rerr = sh.Rm(mytmpdir)
		if rerr != nil {
			t.Errorf("failed to remove dir %s: %v", mytmpdir, rerr)
		}

		_, rerr = os.Stat(mytmpdir)
		if rerr == nil {
			t.Errorf("removed dir %s but it's still there?", mytmpdir)
		}
	})

	t.Run("sh/rm/nedir", func(t *testing.T) {
		rerr := sh.Rm(mytmpdir)
		if rerr != nil {
			t.Errorf("sh.Rm complained removing nonexistent dir %s", mytmpdir)
		}
	})
}

// compareDirs checks that two directories are identical for testing purposes.
// That means they have the same contents, and the same permissions. It does
// NOT mean they have the same timestamp, as that is expected to change in
// normal Mage sh helpers operation.
func compareDirs(dst, src string) error {
	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		newDst := filepath.Clean(strings.Replace(path, src, dst, 1))

		if info.IsDir() {
			return compareStats(newDst, path)
		}

		return compareFiles(newDst, path)
	}

	return filepath.Walk(src, fn)
}

// compareFiles checks that two files are identical for testing purposes. That
// means they have the same length, the same contents, and the same
// permissions. It does NOT mean they have the same timestamp, as that is
// expected to change in normal Mage sh.Copy operation.
func compareFiles(dst, src string) error {
	if err := compareStats(dst, src); err != nil {
		return err
	}

	dbytes, err := ioutil.ReadFile(dst)
	if err != nil {
		return fmt.Errorf("can't read %s: %v", dst, err)
	}

	sbytes, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("can't read %s: %v", src, err)
	}

	if !bytes.Equal(dbytes, sbytes) {
		return fmt.Errorf("files %s and %s have different contents", src, dst)
	}

	return nil
}

func compareStats(dst, src string) error {
	dfi, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("can't stat %s: %v", dst, err)
	}

	sfi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("can't stat %s: %v", src, err)
	}

	if dfi.Size() != sfi.Size() {
		msg := "files %s and %s have different sizes: %d vs %d"
		return fmt.Errorf(msg, dst, src, dfi.Size(), sfi.Size())
	}

	if dfi.Mode() != sfi.Mode() {
		msg := "files %s and %s have different permissions: %#4o vs %#4o"
		return fmt.Errorf(msg, dst, src, dfi.Mode(), sfi.Mode())
	}

	return nil
}
