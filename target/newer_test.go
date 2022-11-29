package target

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewestModTime(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp dir: %s", err.Error())
	}
	defer os.RemoveAll(dir)
	for _, name := range []string{"a", "b", "c", "d"} {
		out := filepath.Join(dir, name)
		if err := ioutil.WriteFile(out, []byte("hi!"), 0644); err != nil {
			t.Fatalf("error writing file: %s", err.Error())
		}
	}
	time.Sleep(10 * time.Millisecond)
	outName := filepath.Join(dir, "c")
	outfh, err := os.OpenFile(outName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("error opening file to append: %s", err.Error())
	}
	if _, err := outfh.WriteString("\nbye!\n"); err != nil {
		t.Fatalf("error appending to file: %s", err.Error())
	}
	if err := outfh.Close(); err != nil {
		t.Fatalf("error closing file: %s", err.Error())
	}

	afi, err := os.Stat(filepath.Join(dir, "a"))
	if err != nil {
		t.Fatalf("error stating unmodified file: %s", err.Error())
	}

	cfi, err := os.Stat(outName)
	if err != nil {
		t.Fatalf("error stating modified file: %s", err.Error())
	}
	if afi.ModTime().Equal(cfi.ModTime()) {
		t.Fatal("modified and unmodified file mtimes equal")
	}

	newest, err := NewestModTime(dir)
	if err != nil {
		t.Fatalf("error finding newest mod time: %s", err.Error())
	}
	if !newest.Equal(cfi.ModTime()) {
		t.Fatal("expected newest mod time to match c")
	}
}

func TestOldestModTime(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp dir: %s", err.Error())
	}
	defer os.RemoveAll(dir)
	for _, name := range []string{"a", "b", "c", "d"} {
		out := filepath.Join(dir, name)
		if err := ioutil.WriteFile(out, []byte("hi!"), 0644); err != nil {
			t.Fatalf("error writing file: %s", err.Error())
		}
	}
	time.Sleep(10 * time.Millisecond)
	for _, name := range []string{"a", "b", "d"} {
		outName := filepath.Join(dir, name)
		outfh, err := os.OpenFile(outName, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("error opening file to append: %s", err.Error())
		}
		if _, err := outfh.WriteString("\nbye!\n"); err != nil {
			t.Fatalf("error appending to file: %s", err.Error())
		}
		if err := outfh.Close(); err != nil {
			t.Fatalf("error closing file: %s", err.Error())
		}
	}

	afi, err := os.Stat(filepath.Join(dir, "a"))
	if err != nil {
		t.Fatalf("error stating unmodified file: %s", err.Error())
	}

	outName := filepath.Join(dir, "c")
	cfi, err := os.Stat(outName)
	if err != nil {
		t.Fatalf("error stating modified file: %s", err.Error())
	}
	if afi.ModTime().Equal(cfi.ModTime()) {
		t.Fatal("modified and unmodified file mtimes equal")
	}

	newest, err := OldestModTime(dir)
	if err != nil {
		t.Fatalf("error finding oldest mod time: %s", err.Error())
	}
	if !newest.Equal(cfi.ModTime()) {
		t.Fatal("expected newest mod time to match c")
	}
}
