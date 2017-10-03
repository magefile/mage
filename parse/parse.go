package parse

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"strings"
)

type PkgInfo struct {
	Funcs          []Function
	DefaultIsError bool
	DefaultName    string
}

type Function struct {
	Name     string
	IsError  bool
	Synopsis string
	Comment  string
}

// Package parses a package
func Package(path string, files []string) (*PkgInfo, error) {
	fset := token.NewFileSet()

	pkg, err := getPackage(path, files, fset)
	if err != nil {
		return nil, err
	}

	info, err := makeInfo(path, fset, pkg.Files)
	if err != nil {
		return nil, err
	}

	pi := &PkgInfo{}

	p := doc.New(pkg, "./", 0)
	for _, f := range p.Funcs {
		if f.Recv != "" {
			// skip methods
			continue
		}
		if !ast.IsExported(f.Name) {
			// skip non-exported functions
			continue
		}
		if typ := voidOrError(f.Decl.Type, info); typ != InvalidType {
			pi.Funcs = append(pi.Funcs, Function{
				Name:     f.Name,
				Comment:  f.Doc,
				Synopsis: doc.Synopsis(f.Doc),
				IsError:  voidOrError(f.Decl.Type, info) == ErrorType,
			})
		}
	}

	setDefault(p, pi, info)

	return pi, nil
}

func setDefault(p *doc.Package, pi *PkgInfo, info types.Info) {
	for _, v := range p.Vars {
		for x, name := range v.Names {
			if name != "Default" {
				continue
			}
			spec := v.Decl.Specs[x].(*ast.ValueSpec)
			if len(spec.Values) != 1 {
				log.Println("warning: default declaration has multiple values")
			}
			id, ok := spec.Values[0].(*ast.Ident)
			if !ok {
				log.Println("warning: default declaration is not a function name")
			}
			for _, f := range pi.Funcs {
				if f.Name == id.Name {
					pi.DefaultName = f.Name
					pi.DefaultIsError = f.IsError
					return
				}
			}
			log.Println("warning: default declaration does not reference a mage target")
		}
	}
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

func makeInfo(dir string, fset *token.FileSet, files map[string]*ast.File) (types.Info, error) {
	cfg := types.Config{
		Importer: getImporter(fset),
	}

	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	fs := make([]*ast.File, 0, len(files))
	for _, v := range files {
		fs = append(fs, v)
	}

	_, err := cfg.Check(dir, fset, fs, &info)
	return info, err
}

// errorOrVoid filters the list of functions to only those that return only an
// error or have no return value, and have no paramters.
func errorOrVoid(fns []*ast.FuncDecl, info types.Info) []*ast.FuncDecl {
	fds := []*ast.FuncDecl{}

	for _, fn := range fns {
		if voidOrError(fn.Type, info) != InvalidType {
			fds = append(fds, fn)
		}
	}
	return fds
}

// FuncType is the type of a function that mage understands.
type FuncType int

// FuncTypes
const (
	InvalidType FuncType = iota
	VoidType
	ErrorType
)

func voidOrError(ft *ast.FuncType, info types.Info) FuncType {
	// look for functions with no params
	if ft.Params.NumFields() > 0 {
		return InvalidType
	}

	// look for functions with 0 or 1 return values
	res := ft.Results
	if res.NumFields() > 1 {
		return InvalidType
	}
	// 0 return value is ok
	if res.NumFields() == 0 {
		return VoidType
	}
	// if 1 return value, look for those that return an error
	ret := res.List[0]

	// handle (a, b, c int)
	if len(ret.Names) > 1 {
		return InvalidType
	}
	t := info.TypeOf(ret.Type)
	if t != nil && t.String() == "error" {
		return ErrorType
	}
	return InvalidType
}
