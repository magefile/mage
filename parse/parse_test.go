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
	if !info.DefaultFunc.IsError {
		t.Fatal("expected DefaultIsError to be true")
	}

	// DefaultName
	if info.DefaultFunc.Name != "ReturnsNilError" {
		t.Fatal("expected DefaultName to be ReturnsNilError")
	}

	if info.Aliases["void"].Name != "ReturnsVoid" {
		t.Fatal("expected alias of void to be ReturnsVoid")
	}

	f, ok := info.Aliases["baz"]
	if !ok {
		t.Fatal("missing alias baz")
	}
	if f.Name != "Baz" || f.Receiver != "Build" {
		t.Fatal("expected alias of void to be Build.Baz")
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
			}
			t.Logf("%#v", infoFn)
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

func TestOptionalArgs(t *testing.T) {
	info, err := PrimaryPackage("go", "./testdata", []string{"optargs.go"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []Function{
		{
			Name: "AllOptional",
			Args: []Arg{
				{Name: "a", Type: "string", Optional: true},
				{Name: "b", Type: "int", Optional: true},
			},
		},
		{
			Name: "OptionalBool",
			Args: []Arg{
				{Name: "verbose", Type: "bool", Optional: true},
			},
		},
		{
			Name: "OptionalDuration",
			Args: []Arg{
				{Name: "base", Type: "time.Duration"},
				{Name: "extra", Type: "time.Duration", Optional: true},
			},
		},
		{
			Name: "OptionalFloat64",
			Args: []Arg{
				{Name: "value", Type: "float64"},
				{Name: "factor", Type: "float64", Optional: true},
			},
		},
		{
			Name: "OptionalInt",
			Args: []Arg{
				{Name: "a", Type: "int"},
				{Name: "b", Type: "int", Optional: true},
			},
		},
		{
			Name: "OptionalString",
			Args: []Arg{
				{Name: "name", Type: "string"},
				{Name: "greeting", Type: "string", Optional: true},
			},
		},
	}

	if len(info.Funcs) != len(expected) {
		t.Fatalf("expected %d funcs, got %d", len(expected), len(info.Funcs))
	}

	for _, fn := range expected {
		found := false
		for _, infoFn := range info.Funcs {
			if reflect.DeepEqual(fn, *infoFn) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected:\n%#v\n\nto be in parsed funcs", fn)
			for _, infoFn := range info.Funcs {
				t.Logf("  got: %#v", *infoFn)
			}
		}
	}
}
