package target

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	dirDelay  = 10 * time.Millisecond
	fileDelay = 100 * time.Millisecond
)

func makeFiles(t *testing.T, dirname string, filenames ...string) {
	dirname = filepath.Join("testdata", dirname)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		if err := os.MkdirAll(dirname, 0777); err != nil {
			t.Fatalf("creating dir %s: %v", dirname, err)
		}
		time.Sleep(dirDelay)
	}

	for _, filename := range filenames {
		path := filepath.Join(dirname, filename)
		if err := ioutil.WriteFile(path, []byte(path), 0600); err != nil {
			t.Fatalf("creating file %s: %v", path, err)
		}
		time.Sleep(fileDelay)
	}
}

func TestPath(t *testing.T) {
	t.Parallel()

	// directory and file relationships:
	//  mod times of files increase according to file number (irrespective of directory)
	//    e.g., 'b/05' is older than 'b/11'
	//  'a' and anything in 'a' is older than anything outside of 'a'
	//  'd' and anything in 'd' is newer than anything outside of 'd'
	//  the directory 'b' is older than 'c', but
	//    'b' has some files both older and newer than anything in 'c'
	makeFiles(t, "a", "01", "02", "03")
	makeFiles(t, "b", "04", "05", "06")
	makeFiles(t, "c", "07", "08", "09")
	makeFiles(t, "b", "10", "11", "12")
	makeFiles(t, "d", "13", "14", "15")
	defer func() {
		if err := os.RemoveAll("testdata"); err != nil {
			t.Fatal(err)
		}
	}()

	table := []struct {
		fn             func(string, ...string) (bool, error)
		desc           string
		target         string
		sources        []string
		targetObsolete bool
		expectedErr    string
	}{
		{
			fn:             File,
			desc:           "target exists and is newer than all sources (files only)",
			target:         "testdata/d/13",
			sources:        []string{"testdata/a/01", "testdata/a/02", "testdata/a/03"},
			targetObsolete: false,
		},
		{
			fn:             File,
			desc:           "target exists but is older than all sources (files only)",
			target:         "testdata/a/01",
			sources:        []string{"testdata/d/13", "testdata/d/14", "testdata/d/15"},
			targetObsolete: true,
		},
		{
			fn:             File,
			desc:           "target does not exist, but all file sources do",
			target:         "testdata/e/16",
			sources:        []string{"testdata/a/01", "testdata/a/02", "testdata/a/03"},
			targetObsolete: true,
		},
		{
			fn:             File,
			desc:           "target exists and is older than sources (files only), but one source does not exist",
			target:         "testdata/a/01",
			sources:        []string{"testdata/b/04", "testdata/z/99", "testdata/b/05"},
			targetObsolete: false,
			expectedErr:    "can't build target testdata/a/01: dependency testdata/z/99 does not exist",
		},
		{
			fn:             File,
			desc:           "target exists and is both older than and newer than some sources (files only)",
			target:         "testdata/b/04",
			sources:        []string{"testdata/a/01", "testdata/c/07"},
			targetObsolete: true,
		},
		{
			fn:             File,
			desc:           "target exists, sources are all directories but are all recursively older",
			target:         "testdata/d/13",
			sources:        []string{"testdata/a", "testdata/b", "testdata/c"},
			targetObsolete: false,
		},
		{
			fn:             File,
			desc:           "target exists, sources are all in one directory; some newer and some older",
			target:         "testdata/c/08",
			sources:        []string{"testdata/b"},
			targetObsolete: true,
		},
		{
			fn:             File,
			desc:           "target exists; sources are files and directories; all are older",
			target:         "testdata/d/15",
			sources:        []string{"testdata/b", "testdata/a/02"},
			targetObsolete: false,
		},
		{
			fn:             File,
			desc:           "target exists; directory sources are older, but file sources are newer",
			target:         "testdata/c/08",
			sources:        []string{"testdata/a", "testdata/d/13", "testdata/d/14"},
			targetObsolete: true,
		},
		{
			fn:             File,
			desc:           "target exists; file sources are older, but directory sources are newer",
			target:         "testdata/c/08",
			sources:        []string{"testdata/a/01", "testdata/a/02", "testdata/d"},
			targetObsolete: true,
		},
		{
			fn:             Dir,
			desc:           "target is a directory, sources are files; all older",
			target:         "testdata/d",
			sources:        []string{"testdata/a/01", "testdata/b/04"},
			targetObsolete: false,
		},
		{
			fn:             Dir,
			desc:           "target is a directory, sources are files; all newer",
			target:         "testdata/a",
			sources:        []string{"testdata/b/04", "testdata/d/13", "testdata/d/14"},
			targetObsolete: true,
		},
		{
			fn:             Dir,
			desc:           "target is a directory, sources are files; some older, some newer",
			target:         "testdata/c",
			sources:        []string{"testdata/a/01", "testdata/d/13"},
			targetObsolete: true,
		},
		{
			fn:             Dir,
			desc:           "target is a directory, sources is a directory with some newer files and some older",
			target:         "testdata/c",
			sources:        []string{"testdata/b"},
			targetObsolete: true,
		},
	}

	for _, c := range table {
		t.Run(c.desc, func(t *testing.T) {
			v, err := c.fn(c.target, c.sources...)

			if err != nil {
				if err.Error() != c.expectedErr {
					if c.expectedErr == "" {
						t.Logf("no error was expected")
					} else {
						t.Logf("expected error: '%s'", c.expectedErr)
					}
					t.Logf("raised error: '%s'", err.Error())
					t.Fail()
				}
			} else {
				if c.expectedErr != "" {
					t.Logf("expected error: '%s'", c.expectedErr)
					t.Logf("no error was raised")
					t.Fail()
				}
			}

			if v != c.targetObsolete {
				t.Fatalf("expected obsolete status '%v'; got '%v'", c.targetObsolete, v)
			}
		})
	}
}
