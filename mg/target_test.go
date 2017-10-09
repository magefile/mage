package mg

import "testing"

func TestTarget(t *testing.T) {
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
