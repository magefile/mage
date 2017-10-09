package mg

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestTarget(t *testing.T) {
	t.Parallel()
	dirs := []string{"testdata/src_dir", "testdata/target_dir"}
	for _, d := range dirs {
		os.MkdirAll(d, 0777)
		time.Sleep(time.Second)
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
		time.Sleep(time.Second)
	}
	defer func() {
		err := os.RemoveAll("testdata")
		if err != nil {
			t.Fatal(err)
		}
	}()

	table := []struct {
		desc      string
		src       string
		recursive bool
		targets   []string
		expect    bool
	}{
		{"Files only target checks", "testdata/src_file", false,
			[]string{"testdata/target_file"}, true},
		{"Files only target doesn't check", "testdata/target_file",
			false, []string{"testdata/src_file"}, false},
		{"Directories only target checks", "testdata/src_dir", false,
			[]string{"testdata/target_dir"}, true},
		{"Directories only target doesn't  checks", "testdata/target_dir", false,
			[]string{"testdata/src_dir"}, false},
	}

	for _, c := range table {
		v, err := Target(c.src, c.recursive, c.targets...)
		if err != nil {
			t.Fatal(err)
		}
		if v != c.expect {
			t.Errorf("%s : expecting %v got %v", c.desc, c.expect, v)
		}
	}
}
