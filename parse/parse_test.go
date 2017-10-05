package parse

import (
	"reflect"
	"testing"
	"strings"
)

func TestParse(t *testing.T) {
	info, err := Package("./testdata", []string{"func.go", "command.go"})
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
		{
			Name:      "TakesContextReturnsError",
			IsError:   true,
			IsContext: true,
		},
		{
			Name:      "TakesContextReturnsVoid",
			IsError:   false,
			IsContext: true,
		},
	}


	// DefaultIsError
	if info.DefaultIsError != true {
		t.Fatalf("expected DefaultIsError to be true")
	}

	// DefaultName
	if strings.Compare(info.DefaultName, "ReturnsError") != 0 {
		t.Fatalf("expected DefaultName to be ReturnsError")
	}

	for fnIndex := range(expected) {
		fn := expected[fnIndex]
		found := false
		for infoFnIndex := range(info.Funcs) {
			infoFn := info.Funcs[infoFnIndex]
			if reflect.DeepEqual(fn, infoFn) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected:\n%#v\n\nto be in:\n%#v", fn, info.Funcs)
		}
	}
}
