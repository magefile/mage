package parse

import (
	"go/parser"
	"go/token"
	"go/types"
	"sort"
	"testing"
)

func TestParse(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testfiles/func.go", nil, parser.ParseComments)
	IsNil(err, t)

	makeInfo("testfiles", fset, f)
}

func TestExportedFuncs(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testfiles/func.go", nil, parser.ParseComments)
	IsNil(err, t)

	funcs := exportedFuncs(f, fset)

	names := make([]string, len(funcs))
	for i, fn := range funcs {
		names[i] = fn.Name.Name
	}
	sort.Strings(names)

	Equals([]string{"ReturnsError", "ReturnsString", "ReturnsVoid"}, names, t)
}

func TestErrorOrVoid(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testfiles/func.go", nil, parser.ParseComments)
	IsNil(err, t)

	funcs := exportedFuncs(f, fset)

	info, err := makeInfo("testfiles", fset, f)
	IsNil(err, t)

	funcs = errorOrVoid(funcs, info)

	names := make([]string, len(funcs))
	for i, fn := range funcs {
		names[i] = fn.Name.Name
	}
	sort.Strings(names)

	Equals([]string{"ReturnsError", "ReturnsVoid"}, names, t)
}

func TestComments(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testfiles/command.go", nil, parser.ParseComments)
	IsNil(err, t)

	info, err := makeInfo("testfiles", fset, f)
	IsNil(err, t)

	fns, err := functions(f, info, fset)
	IsNil(err, t)
	Equals(1, len(fns), t)

	fn := fns[0]

	exp := Function{
		Name:    "Command",
		IsError: true,
		Comment: "Command is an example function.\nThis is the second line of comment.\n",
		Params: []Param{
			{
				Name:      "stringArg",
				Type:      types.String,
				IsPointer: false,
				Comment:   "comment for stringArg\n",
			},
			{
				Name:      "stringArg2",
				Type:      types.String,
				IsPointer: false,
				Comment:   "comment for stringArg2\n",
			},
			{
				Name:      "boolArg",
				Type:      types.Bool,
				IsPointer: false,
				Comment:   "comment for boolArg\n",
			},
			{
				Name:      "pStringArg",
				Type:      types.String,
				IsPointer: true,
				Comment:   "comment for pStringArg\n",
			},
		},
	}

	Equals(exp, fn, t)
}
