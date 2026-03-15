package target

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewestModTime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, name := range []string{"a", "b", "c", "d"} {
		out := filepath.Join(dir, name)
		if err := os.WriteFile(out, []byte("hi!"), 0o600); err != nil {
			t.Fatalf("error writing file: %s", err.Error())
		}
	}
	time.Sleep(10 * time.Millisecond)
	outName := filepath.Join(dir, "c")
	outfh, err := os.OpenFile(outName, os.O_APPEND|os.O_WRONLY, 0o600)
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
	dir := t.TempDir()
	for _, name := range []string{"a", "b", "c", "d"} {
		out := filepath.Join(dir, name)
		if err := os.WriteFile(out, []byte("hi!"), 0o600); err != nil {
			t.Fatalf("error writing file: %s", err.Error())
		}
	}
	time.Sleep(10 * time.Millisecond)
	for _, name := range []string{"a", "b", "d"} {
		outName := filepath.Join(dir, name)
		outfh, err := os.OpenFile(outName, os.O_APPEND|os.O_WRONLY, 0o600)
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

func TestDirNewerWithEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// No files inside dir — nothing newer than any target time
	newer, err := DirNewer(time.Now(), dir)
	if err != nil {
		t.Fatal(err)
	}
	// The directory itself was just created, its modtime is very recent
	// so it should be newer than time.Time{} but we pass time.Now()
	// The dir's modtime should be ≤ now, so it should not be newer
	if newer {
		t.Fatal("expected empty dir to not be newer than now")
	}
}

func TestDirNewerMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := DirNewer(time.Now(), filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestNewestModTimeEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// An empty directory still has the directory entry itself
	newest, err := NewestModTime(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Should return the directory's own modtime
	info, _ := os.Stat(dir)
	if !newest.Equal(info.ModTime()) {
		t.Fatalf("expected dir modtime, got %v vs %v", newest, info.ModTime())
	}
}

func TestOldestModTimeEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	oldest, err := OldestModTime(dir)
	if err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(dir)
	if !oldest.Equal(info.ModTime()) {
		t.Fatalf("expected dir modtime, got %v vs %v", oldest, info.ModTime())
	}
}

func TestPathNewerMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := PathNewer(time.Now(), filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestGlobNewerNoMatch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := GlobNewer(time.Now(), filepath.Join(dir, "*.nonexistent"))
	if err == nil {
		t.Fatal("expected error for glob with no matches")
	}
}
