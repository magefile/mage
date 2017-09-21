package parse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
	"unicode"
)

type Function struct {
	Name    string
	IsError bool
	Comment string
}

// Package parses a package
func Package(path string, files []string) ([]Function, error) {
	fset := token.NewFileSet()

	pkg, err := getPackage(path, files, fset)
	if err != nil {
		return nil, err
	}

	f := ast.MergePackageFiles(
		pkg,
		ast.FilterImportDuplicates|ast.FilterUnassociatedComments,
	)

	info, err := makeInfo(path, fset, f)
	if err != nil {
		return nil, err
	}

	return functions(f, info, fset)
}

// getPackage returns the non-test package at the given path.
func getPackage(path string, files []string, fset *token.FileSet) (*ast.Package, error) {
	fm := make(map[string]bool, len(files))
	for _, f := range files {
		fm[f] = true
	}

	filter := func(f os.FileInfo) bool {
		return fm[f.Name()]
	}

	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for name, pkg := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("no non-test packages found in %s", path)
}

func makeInfo(dir string, fset *token.FileSet, f *ast.File) (types.Info, error) {
	cfg := types.Config{IgnoreFuncBodies: true}

	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	_, err := cfg.Check(dir, fset, []*ast.File{f}, &info)
	return info, err
}

func functions(f *ast.File, info types.Info, fset *token.FileSet) ([]Function, error) {
	fns := exportedFuncs(f, fset)
	fns = errorOrVoid(fns, info)

	cmtMap := ast.NewCommentMap(fset, f, f.Comments)

	functions := make([]Function, len(fns))

	for i, fn := range fns {
		fun := Function{Name: fn.Name.Name}
		fun.Comment = combine(cmtMap[fn])

		// we only support null returns or error returns, so if there's a
		// return, it's an error.
		if fn.Type.Results.NumFields() > 0 {
			fun.IsError = true
		}
		functions[i] = fun
	}
	return functions, nil
}

func combine(comments []*ast.CommentGroup) string {
	s := make([]string, len(comments))
	for i, comment := range comments {
		s[i] = comment.Text()
	}
	return strings.Join(s, " ")
}

// exportedFuncs returns a list of exported non-method functions that return
// nothing or just an error.
func exportedFuncs(f *ast.File, fset *token.FileSet) []*ast.FuncDecl {
	fns := []*ast.FuncDecl{}
	for _, decl := range f.Decls {
		// get all top level functions
		if fn, ok := decl.(*ast.FuncDecl); ok {
			// skip all methods
			if fn.Recv != nil {
				continue
			}
			name := fn.Name.Name
			// look for exported functions only
			if unicode.IsUpper([]rune(name)[0]) {
				fns = append(fns, fn)
			}
		}
	}
	return fns
}

// errorOrVoid filters the list of functions to only those that return only an
// error or have no return value, and have no paramters.
func errorOrVoid(fns []*ast.FuncDecl, info types.Info) []*ast.FuncDecl {
	fds := []*ast.FuncDecl{}

	for _, fn := range fns {
		// look for functions with no params
		if fn.Type.Params.NumFields() > 0 {
			continue
		}

		// look for functions with 0 or 1 return values
		res := fn.Type.Results
		if res.NumFields() > 1 {
			continue
		}
		// 0 return value is ok
		if res.NumFields() == 0 {
			fds = append(fds, fn)
			continue
		}
		// if 1 return value, look for those that return an error
		ret := res.List[0]

		// handle (a, b, c int)
		if len(ret.Names) > 1 {
			continue
		}
		t := info.TypeOf(ret.Type)
		if t != nil && t.String() == "error" {
			fds = append(fds, fn)
		}
	}
	return fds
}
