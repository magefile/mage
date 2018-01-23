package target

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestPath(t *testing.T) {
	t.Parallel()
	dirs := []string{"testdata/src_dir", "testdata/target_dir"}
	for _, d := range dirs {
		os.MkdirAll(d, 0777)
		time.Sleep(10 * time.Millisecond)
	}
	files := []string{
		"testdata/src_dir/some_file",
		"testdata/src_dir/some_",
		"testdata/src_file",
		"testdata/target_dir/some_file",
		"testdata/target_file",
	}
	for _, v := range files {
		err := ioutil.WriteFile(v, []byte(v), 0600)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	defer func() {
		err := os.RemoveAll("testdata")
		if err != nil {
			t.Fatal(err)
		}
	}()

	table := []struct {
		desc    string
		src     string
		targets []string
		expect  bool
	}{
		{"Files only target checks", "testdata/src_file",
			[]string{"testdata/target_file"}, true},
		{"Files only target doesn't check", "testdata/target_file",
			[]string{"testdata/src_file"}, false},
		{"Directories only target checks", "testdata/src_dir",
			[]string{"testdata/target_dir"}, true},
		{"Directories only target doesn't  checks", "testdata/target_dir",
			[]string{"testdata/src_dir"}, false},
	}

	for _, c := range table {
		v, err := Path(c.src, c.targets...)
		if err != nil {
			t.Fatal(err)
		}
		if v != c.expect {
			t.Errorf("%s : expecting %v got %v", c.desc, c.expect, v)
		}
	}
}
