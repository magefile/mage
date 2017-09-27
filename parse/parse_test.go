package parse

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	info, err := Package("./testdata", []string{"func.go", "command.go"})
	if err != nil {
		t.Fatal(err)
	}

	expected := PkgInfo{
		Funcs: []Function{
			{
				Name:     "ReturnsError",
				IsError:  true,
				Comment:  "Synopsis for returns error.\nAnd some more text.\n",
				Synopsis: "Synopsis for returns error.",
			},
			{
				Name: "ReturnsVoid",
			},
		},
		DefaultIsError: true,
		DefaultName:    "ReturnsError",
	}
	diff := cmp.Diff(expected, *info)
	if diff != "" {
		t.Error(diff)
	}
}
