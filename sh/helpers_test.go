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

// compareFiles checks that two files are identical for testing purposes. That means they have the same length,
// the same contents, and the same permissions. It does NOT mean they have the same timestamp, as that is expected
// to change in normal Mage sh.Copy operation.
func compareFiles(file1 string, file2 string) error {
	s1, err := os.Stat(file1)
	if err != nil {
		return fmt.Errorf("can't stat %s: %v", file1, err)
	}
	s2, err := os.Stat(file2)
	if err != nil {
		return fmt.Errorf("can't stat %s: %v", file2, err)
	}
	if s1.Size() != s2.Size() {
		return fmt.Errorf("files %s and %s have different sizes: %d vs %d", file1, file2, s1.Size(), s2.Size())
	}
	if s1.Mode() != s2.Mode() {
		return fmt.Errorf("files %s and %s have different permissions: %#4o vs %#4o", file1, file2, s1.Mode(), s2.Mode())
	}
	f1bytes, err := ioutil.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("can't read %s: %v", file1, err)
	}
	f2bytes, err := ioutil.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("can't read %s: %v", file2, err)
	}
	if !bytes.Equal(f1bytes, f2bytes) {
		return fmt.Errorf("files %s and %s have different contents", file1, file2)
	}
	return nil
}

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
	err = ioutil.WriteFile(srcname, []byte("All work and no play makes Jack a dull boy."), 0644)
	if err != nil {
		t.Fatalf("can't create test file %s: %v", srcname, err)
	}
	destname := filepath.Join(mytmpdir, "test2.txt")

	t.Run("sh/copy", func(t *testing.T) {
		cerr := sh.Copy(destname, srcname)
		if cerr != nil {
			t.Errorf("test file copy from %s to %s failed: %v", srcname, destname, cerr)
		}
		cerr = compareFiles(srcname, destname)
		if cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}
	})

	t.Run("sh/copydir", func(t *testing.T) {
		cpddir := filepath.Join(mytmpdir, "copydir-dir")
		if cerr := os.Mkdir(cpddir, 0744); cerr != nil {
			t.Fatalf("can't create CopyDir test directory: %v", cerr)
		}

		// Directory to non existent directory.

		neDir := filepath.Join(mytmpdir, "copydir-non-existent-directory")
		if cerr := sh.CopyDir(neDir, cpddir); cerr != nil {
			msg := "test directory copy from %s to %s failed: %v"
			t.Errorf(msg, cpddir, neDir, cerr)
		}

		if cerr := compareDirs(cpddir, neDir); cerr != nil {
			t.Errorf("test directory copy verification failed: %v", cerr)
		}

		// Directory to directory.

		if cerr := sh.CopyDir(neDir, cpddir); cerr != nil {
			msg := "test directory copy from %s to %s failed: %v"
			t.Errorf(msg, cpddir, neDir, cerr)
		}

		ndir := filepath.Join(neDir, "copydir-dir")
		if cerr := compareDirs(cpddir, ndir); cerr != nil {
			t.Errorf("test directory copy verification failed: %v", cerr)
		}
	})

	t.Run("sh/copydir/ne", func(t *testing.T) {
		ned := filepath.Join(mytmpdir, "copydir-ne-dir")
		if cerr := sh.CopyDir(destname, ned); cerr == nil {
			t.Errorf("sh.CopyDir succeeded copying nonexistent directory %s", ned)
		}
	})

	t.Run("sh/copydir/files", func(t *testing.T) {
		if cerr := sh.CopyDir(mytmpdir, srcname); cerr == nil {
			t.Error("sh.CopyDir succeeded using file as source")
		}

		otherd := filepath.Join(mytmpdir, "copydir-files-dir")
		if cerr := os.Mkdir(otherd, 0744); cerr != nil {
			t.Fatalf("can't create other test directory: %v", cerr)
		}

		if cerr := sh.CopyDir(destname, otherd); cerr == nil {
			t.Error("sh.CopyDir succeeded using file as destination")
		}
	})

	t.Run("sh/copydir/d2itself", func(t *testing.T) {
		ii := filepath.Join(mytmpdir, "copydir-d2itself-dir")
		if cerr := sh.CopyDir(ii, mytmpdir); cerr == nil {
			t.Error("sh.CopyDir succeeded copying folders into itself")
		}
	})

	t.Run("sh/copydir/lne", func(t *testing.T) {
		otherd := filepath.Join(mytmpdir, "copydir-lne-dir")
		if cerr := os.Mkdir(otherd, 0744); cerr != nil {
			t.Fatalf("can't create other test directory: %v", cerr)
		}

		ned := filepath.Join(mytmpdir, "long", "non", "existent", "directory")
		if cerr := sh.CopyDir(ned, otherd); cerr == nil {
			msg := "sh.CopyDir succeeded copying to a long nonexistent directory %s"
			t.Errorf(msg, ned)
		}
	})

	t.Run("sh/cp", func(t *testing.T) {
		cpdir := filepath.Join(mytmpdir, "cp-dir")
		if cerr := os.Mkdir(cpdir, 0744); cerr != nil {
			t.Fatalf("can't create cp test directory: %v", cerr)
		}

		// File to non existent file.

		cpdestname := filepath.Join(cpdir, "cp-test.txt")
		if cerr := sh.Cp(cpdestname, srcname); cerr != nil {
			msg := "test file copy from %s to %s failed: %v"
			t.Errorf(msg, srcname, cpdestname, cerr)
		}

		if cerr := compareFiles(srcname, cpdestname); cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}

		// File to file.

		otherf := filepath.Join(mytmpdir, "other.txt")
		if cerr := ioutil.WriteFile(otherf, []byte("Content"), 0644); err != nil {
			t.Fatalf("can't create other test file %s: %v", otherf, cerr)
		}

		if cerr := sh.Cp(cpdestname, otherf); cerr != nil {
			msg := "test file copy from %s to %s failed: %v"
			t.Errorf(msg, otherf, cpdestname, cerr)
		}

		if cerr := compareFiles(otherf, cpdestname); cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}

		// File to directory.

		if cerr := sh.Cp(cpdir, srcname); cerr != nil {
			msg := "test file copy from %s to %s/ failed: %v"
			t.Errorf(msg, srcname, cpdir, cerr)
		}

		cpdestname = filepath.Join(cpdir, filepath.Base(srcname))
		if cerr := compareFiles(srcname, cpdestname); cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}

		// Directory to non existent directory.

		neDir := filepath.Join(mytmpdir, "cp-non-existent-directory")
		if cerr := sh.Cp(neDir, cpdir); cerr != nil {
			msg := "test directory copy from %s to %s failed: %v"
			t.Errorf(msg, cpdir, neDir, cerr)
		}

		if cerr := compareDirs(cpdir, neDir); cerr != nil {
			t.Errorf("test directory copy verification failed: %v", cerr)
		}

		// Directory to directory.

		if cerr := sh.Cp(cpdir, neDir); cerr != nil {
			msg := "test directory copy from %s to %s failed: %v"
			t.Errorf(msg, neDir, cpdir, cerr)
		}

		ndir := filepath.Join(cpdir, "cp-non-existent-directory")
		if cerr := compareDirs(neDir, ndir); cerr != nil {
			t.Errorf("test directory copy verification failed: %v", cerr)
		}

		// Multiple elements to directory.

		mdir := filepath.Join(mytmpdir, "cp-multiple-elements")
		if cerr := os.Mkdir(mdir, 0744); cerr != nil {
			t.Fatalf("can't create multiple type test directory: %v", cerr)
		}

		if cerr := sh.Cp(mdir, cpdir, srcname, destname); cerr != nil {
			msg := "test multiple copy to %s failed: %v"
			t.Errorf(msg, mdir, cerr)
		}

		ndir = filepath.Join(mdir, "cp-dir")
		if cerr := compareDirs(cpdir, ndir); cerr != nil {
			t.Errorf("test directory copy verification failed: %v", cerr)
		}

		nfile := filepath.Join(mdir, "test1.txt")
		if cerr := compareFiles(srcname, nfile); cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}

		nfile = filepath.Join(mdir, "test2.txt")
		if cerr := compareFiles(destname, nfile); cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}
	})

	t.Run("sh/cp/ns", func(t *testing.T) {
		if cerr := sh.Cp(destname); cerr == nil {
			t.Error("sh.Cp succeeded without sources")
		}
	})

	t.Run("sh/cp/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "cp_file_not_exist.txt")
		if cerr := sh.Cp(destname, nef); cerr == nil {
			t.Errorf("sh.Cp succeeded copying nonexistent file %s", nef)
		}
	})

	t.Run("sh/cp/f2lne", func(t *testing.T) {
		ned := filepath.Join(mytmpdir, "long", "non", "existent", "directory")
		if cerr := sh.Cp(ned, srcname); cerr == nil {
			msg := "sh.Cp succeeded copying to a long nonexistent directory %s"
			t.Errorf(msg, ned)
		}
	})

	t.Run("sh/cp/d2f", func(t *testing.T) {
		otherd := filepath.Join(mytmpdir, "cp-d2f-dir")
		if cerr := os.Mkdir(otherd, 0744); cerr != nil {
			t.Fatalf("can't create other test directory: %v", cerr)
		}

		if cerr := sh.Cp(destname, otherd); cerr == nil {
			t.Error("sh.Cp replaces files with folders")
		}
	})

	t.Run("sh/cp/d2lne", func(t *testing.T) {
		otherd := filepath.Join(mytmpdir, "cp-d2lne-dir")
		if cerr := os.Mkdir(otherd, 0744); cerr != nil {
			t.Fatalf("can't create other test directory: %v", cerr)
		}

		ned := filepath.Join(mytmpdir, "long", "non", "existent", "directory")
		if cerr := sh.Cp(ned, otherd); cerr == nil {
			msg := "sh.Cp succeeded copying to a long nonexistent directory %s"
			t.Errorf(msg, ned)
		}
	})

	t.Run("sh/cp/dne", func(t *testing.T) {
		ned := filepath.Join(mytmpdir, "cp_dir_not_exist")

		if cerr := sh.Cp(ned, srcname, srcname); cerr == nil {
			t.Errorf("sh.Cp succeeded copying to nonexistent destination %s", ned)
		}
	})

	t.Run("sh/cp/m2f", func(t *testing.T) {
		if cerr := sh.Cp(destname, srcname, srcname); cerr == nil {
			t.Error("sh.Cp succeeded copying multiple elements to file")
		}
	})

	t.Run("sh/cp/forbidden", func(t *testing.T) {
		fd, err := ioutil.TempDir("", "mage-forbidden")
		if err != nil {
			t.Fatalf("can't create test directory: %v", err)
		}

		forbidden := filepath.Join(fd, "forbidden")
		if cerr := os.Mkdir(forbidden, 0000); cerr != nil {
			t.Fatalf("can't create a forbidden directory: %v", cerr)
		}

		if cerr := sh.Cp(forbidden, mytmpdir); cerr == nil {
			t.Error("sh.Cp succeeded copying elements to forbidden folder")
		}
	})

	// While we've got a temporary directory, test how forgiving sh.Rm is
	t.Run("sh/rm/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "file_not_exist.txt")
		rerr := sh.Rm(nef)
		if rerr != nil {
			t.Errorf("sh.Rm complained when removing nonexistent file %s: %v", nef, rerr)
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

func compareDirs(src, dst string) error {
	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		dstfile := filepath.Clean(strings.Replace(path, src, dst, 1))
		return compareFiles(path, dstfile)
	}

	return filepath.Walk(src, fn)
}
