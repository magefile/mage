package target

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPathMissingDest(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	src := filepath.Join(dir, "source")
	err = ioutil.WriteFile(src, []byte("hi!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "missing")
	rebuild, err := Path(dst, src)
	if err != nil {
		t.Fatal("Expected no error, but got", err)
	}
	if !rebuild {
		t.Fatal("expected to be told to rebuild, but got false")
	}
}

func TestPathMissingSource(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dst := filepath.Join(dir, "dst")
	err = ioutil.WriteFile(dst, []byte("hi!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "missing")
	_, err = Path(dst, src)
	if !os.IsNotExist(err) {
		t.Fatal("Expected os.IsNotExist(err), but got", err)
	}
}

func TestDirMissingSrc(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	dst := filepath.Join(dir, "dst")
	err = ioutil.WriteFile(dst, []byte("hi!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "missing")
	_, err = Dir(dst, src)
	if !os.IsNotExist(err) {
		t.Fatal("Expected os.IsNotExist(err), but got", err)
	}
}

func TestDirMissingDest(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	src := filepath.Join(dir, "source")
	err = os.Mkdir(src, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(src, "somefile"), []byte("hi!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "missing")
	rebuild, err := Dir(dst, src)
	if err != nil {
		t.Fatal("Expected no error, but got", err)
	}
	if !rebuild {
		t.Fatal("expected to be told to rebuild, but got false")
	}
}

func TestPath(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	err = os.MkdirAll(filepath.Join(dir, filepath.FromSlash("dir/dir2")), 0777)
	if err != nil {
		t.Fatal(err)
	}
	// files are created in order so we know how to expect
	files := []string{
		"file_one",
		"dir/file_two",
		"file_three",
		"dir/dir2/file_four",
	}
	for _, v := range files {
		time.Sleep(10 * time.Millisecond)
		f := filepath.Join(dir, filepath.FromSlash(v))
		err := ioutil.WriteFile(f, []byte(v), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	table := []struct {
		desc    string
		target  string
		sources []string
		expect  bool
	}{
		{
			desc:    "Missing target",
			target:  "missing_file",
			sources: []string{"file_one"},
			expect:  true,
		},
		{
			desc:    "Target is newer",
			target:  "file_three",
			sources: []string{"file_one"},
			expect:  false,
		},
		{
			desc:    "Target is older",
			target:  "file_one",
			sources: []string{"file_three"},
			expect:  true,
		},
		{
			// note that even though file_four is in dir/dir2 ... the modtimes
			// only get propagated up to the parent directory of the folder, not
			// propagated all the way up.
			desc:    "Source is older dir",
			target:  "file_three",
			sources: []string{"dir"},
			expect:  false,
		},
		{
			desc:    "Source is newer dir2",
			target:  "file_three",
			sources: []string{"dir/dir2"},
			expect:  true,
		},
		{
			desc:    "Source is newer dir",
			target:  "file_one",
			sources: []string{"dir"},
			expect:  true,
		},
	}

	for _, c := range table {
		t.Run(c.desc, func(t *testing.T) {
			for i := range c.sources {
				c.sources[i] = filepath.Join(dir, c.sources[i])
			}
			c.target = filepath.Join(dir, c.target)
			v, err := Path(c.target, c.sources...)
			if err != nil {
				t.Fatal(err)
			}
			if v != c.expect {
				t.Errorf("expecting %v got %v", c.expect, v)
			}
		})
	}
}

func TestDir(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	err = os.MkdirAll(filepath.Join(dir, filepath.FromSlash("dir/dir2")), 0777)
	if err != nil {
		t.Fatal(err)
	}
	// files are created in order so we know which one is newer
	files := []string{
		"file_one",
		"dir/file_two",
		"file_three",
		"dir/dir2/file_four",
		"file_five",
	}
	for _, v := range files {
		time.Sleep(10 * time.Millisecond)
		f := filepath.Join(dir, filepath.FromSlash(v))
		err := ioutil.WriteFile(f, []byte(v), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	table := []struct {
		desc    string
		target  string
		sources []string
		expect  bool
	}{
		{
			desc:    "Missing target",
			target:  "missing_file",
			sources: []string{"file_one"},
			expect:  true,
		},
		{
			desc:    "Target is newer",
			target:  "file_three",
			sources: []string{"file_one"},
			expect:  false,
		},
		{
			desc:    "Target is older",
			target:  "file_one",
			sources: []string{"file_three"},
			expect:  true,
		},
		{
			desc:    "Source is older dir",
			target:  "file_five",
			sources: []string{"dir"},
			expect:  false,
		},
		{
			desc:    "Source is newer dir",
			target:  "file_one",
			sources: []string{"dir"},
			expect:  true,
		},
		{
			// This is the tricky one. The modtime on "dir" will be the same
			// as the modtime on dir/file_two, but the modtime on the subdir
			// will be the same as the modtime on dir/dir2/file_four
			// and therefor the should say the source is newer.
			desc:    "Source is newer subdir",
			target:  "file_three",
			sources: []string{"dir"},
			expect:  true,
		},
	}

	for _, c := range table {
		t.Run(c.desc, func(t *testing.T) {
			sources := make([]string, len(c.sources))
			for i := range c.sources {
				sources[i] = filepath.Join(dir, c.sources[i])
			}
			target := filepath.Join(dir, c.target)
			v, err := Dir(target, sources...)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if v != c.expect {
				t.Errorf("expecting %v got %v", c.expect, v)
			}
		})
	}
}
