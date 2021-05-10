package parse

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/magefile/mage/internal"
)

func init() {
	internal.SetDebug(log.New(os.Stdout, "", 0))
}

func TestParse(t *testing.T) {
	info, err := PrimaryPackage("go", "./testdata", []string{"func.go", "command.go", "alias.go", "repeating_synopsis.go", "subcommands.go"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []Function{
		{
			Name:     "ReturnsNilError",
			IsError:  true,
			Comment:  "Synopsis for \"returns\" error. And some more text.",
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
			Comment:  "RepeatingSynopsis chops off the repeating function name. Some more text.",
			Synopsis: "chops off the repeating function name.",
		},
		{
			Name:     "Foobar",
			Receiver: "Build",
			IsError:  true,
		},
		{
			Name:     "Baz",
			Receiver: "Build",
			IsError:  false,
		},
	}

	if info.DefaultFunc == nil {
		t.Fatal("expected default func to exist, but was nil")
	}

	// DefaultIsError
	if info.DefaultFunc.IsError != true {
		t.Fatalf("expected DefaultIsError to be true")
	}

	// DefaultName
	if info.DefaultFunc.Name != "ReturnsNilError" {
		t.Fatalf("expected DefaultName to be ReturnsNilError")
	}

	if info.DeinitFunc == nil {
		t.Fatal("expected deinit func to exist, but was nil")
	}

	// DeinitIsError
	if info.DeinitFunc.IsError != true {
		t.Fatalf("expected DeinitIsError to be true")
	}

	// DeinitName
	if info.DeinitFunc.Name != "Shutdown" {
		t.Fatalf("expected DeinitName to be Shutdown")
	}

	if info.Aliases["void"].Name != "ReturnsVoid" {
		t.Fatalf("expected alias of void to be ReturnsVoid")
	}

	f, ok := info.Aliases["baz"]
	if !ok {
		t.Fatal("missing alias baz")
	}
	if f.Name != "Baz" || f.Receiver != "Build" {
		t.Fatalf("expected alias of void to be Build.Baz")
	}

	if len(info.Aliases) != 2 {
		t.Fatalf("expected to only have two aliases, but have %#v", info.Aliases)
	}

	for _, fn := range expected {
		found := false
		for _, infoFn := range info.Funcs {
			if reflect.DeepEqual(fn, *infoFn) {
				found = true
				break
			} else {
				t.Logf("%#v", infoFn)
			}
		}
		if !found {
			t.Fatalf("expected:\n%#v\n\nto be in:\n%#v", fn, info.Funcs)
		}
	}
}

func TestGetImportSelf(t *testing.T) {
	imp, err := getImport("go", "github.com/magefile/mage/parse/testdata/importself", "")
	if err != nil {
		t.Fatal(err)
	}
	if imp.Info.AstPkg.Name != "importself" {
		t.Fatalf("expected package importself, got %v", imp.Info.AstPkg.Name)
	}
}
