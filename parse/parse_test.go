package parse

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	info, err := Package("./testdata", []string{"func.go", "command.go", "alias.go", "repeating_synopsis.go"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []Function{
		{
			Name:     "ReturnsNilError",
			IsError:  true,
			Comment:  "Synopsis for \"returns\" error.\nAnd some more text.\n",
			Synopsis: `Synopsis for "returns" error.`,
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
		{
			Name:     "RepeatingSynopsis",
			IsError:  true,
			Comment:  "RepeatingSynopsis chops off the repeating function name.\nSome more text.\n",
			Synopsis: "chops off the repeating function name.",
		},
	}

	// DefaultIsError
	if info.DefaultIsError != true {
		t.Fatalf("expected DefaultIsError to be true")
	}

	// DefaultName
	if info.DefaultName != "ReturnsNilError" {
		t.Fatalf("expected DefaultName to be ReturnsNilError")
	}

	if info.Aliases["void"] != "ReturnsVoid" {
		t.Fatalf("expected alias of void to be ReturnsVoid")
	}

	for _, fn := range expected {
		found := false
		for _, infoFn := range info.Funcs {
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
