package parse

import (
	"reflect"
	"testing"
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

	// TODO: Don't use DeepEqual, you lazy git!
	if !reflect.DeepEqual(expected, *info) {
		t.Fatalf("expected:\n%#v\n\ngot:\n%#v", expected, *info)
	}
}
