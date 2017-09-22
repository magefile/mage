package parse

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	fns, err := Package("./testdata", []string{"func.go", "command.go"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []Function{
		{
			Name:     "ReturnsError",
			IsError:  true,
			Comment:  "Synopsis for returns error.\nAnd some more text.\n",
			Synopsis: "Synopsis for returns error.",
		},
		{
			Name: "ReturnsVoid",
		},
	}
	diff := cmp.Diff(expected, fns)
	if diff != "" {
		t.Error(diff)
	}
}
