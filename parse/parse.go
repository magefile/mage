package parse

import (
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"

	mgTypes "github.com/magefile/mage/types"
)

var debug = log.New(ioutil.Discard, "DEBUG: ", 0)

// EnableDebug turns on debug logging.
func EnableDebug() {
	debug.SetOutput(os.Stderr)
}

// PkgInfo contains inforamtion about a package of files according to mage's
// parsing rules.
type PkgInfo struct {
	Description      string
	Funcs            []Function
	DefaultIsError   bool
	DefaultIsContext bool
	DefaultName      string
	DefaultFunc      Function
	Aliases          map[string]string
}

// Function represented a job function from a mage file
type Function struct {
	Name      string
	Receiver  string
	IsError   bool
	IsContext bool
	Synopsis  string
	Comment   string
}

// TemplateName returns the invocation name, supporting namespaced functions.
func (f Function) TemplateName() string {
	if f.Receiver != "" {
		return strings.ToLower(fmt.Sprintf("%s:%s", f.Receiver, f.Name))
	}

	return f.Name
}

// TemplateString returns code for the template switch to run the target.
// It wraps each target call to match the func(context.Context) error that
// runTarget requires.
func (f Function) TemplateString() string {
	name := f.Name
	if f.Receiver != "" {
		name = fmt.Sprintf("%s{}.%s", f.Receiver, f.Name)
	}
	if f.IsContext && f.IsError {
		out := `wrapFn := func(ctx context.Context) error {
				return %s(ctx)
			}
			err := runTarget(wrapFn)`
		return fmt.Sprintf(out, name)
	}
	if f.IsContext && !f.IsError {
		out := `wrapFn := func(ctx context.Context) error {
				%s(ctx)
				return nil
			}
			err := runTarget(wrapFn)`
		return fmt.Sprintf(out, name)
	}
	if !f.IsContext && f.IsError {
		out := `wrapFn := func(ctx context.Context) error {
				return %s()
			}
			err := runTarget(wrapFn)`
		return fmt.Sprintf(out, name)
	}
	if !f.IsContext && !f.IsError {
		out := `wrapFn := func(ctx context.Context) error {
				%s()
				return nil
			}
			err := runTarget(wrapFn)`
		return fmt.Sprintf(out, name)
	}
	return `fmt.Printf("Error formatting job code\n")
	os.Exit(1)`
}

// Package parses a package
func Package(path string, files []string) (*PkgInfo, error) {
	fset := token.NewFileSet()
	pkg, err := getPackage(path, files, fset)
	if err != nil {
		return nil, err
	}

	p := doc.New(pkg, "./", 0)
	pi := &PkgInfo{
		Description: toOneLine(p.Doc),
	}

typeloop:
	for _, t := range p.Types {
		for _, s := range t.Decl.Specs {
			if id, ok := s.(*ast.TypeSpec); ok {
				if sel, ok := id.Type.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if ident.Name != "mg" || sel.Sel.Name != "Namespace" {
							continue typeloop
						}
					}
				}
				break
			}
		}
		for _, f := range t.Methods {
			if !ast.IsExported(f.Name) {
				continue
			}
			if typ := voidOrError(f.Decl.Type); typ != mgTypes.InvalidType {
				pi.Funcs = append(pi.Funcs, Function{
					Name:      f.Name,
					Receiver:  f.Recv,
					Comment:   toOneLine(f.Doc),
					Synopsis:  sanitizeSynopsis(f),
					IsError:   typ == mgTypes.ErrorType || typ == mgTypes.ContextErrorType,
					IsContext: typ == mgTypes.ContextVoidType || typ == mgTypes.ContextErrorType,
				})
			}
		}
	}
	for _, f := range p.Funcs {
		if f.Recv != "" {
			debug.Printf("skipping method %s.%s", f.Recv, f.Name)
			// skip methods
			continue
		}
		if !ast.IsExported(f.Name) {
			debug.Printf("skipping non-exported function %s", f.Name)
			// skip non-exported functions
			continue
		}
		if typ := voidOrError(f.Decl.Type); typ != mgTypes.InvalidType {
			debug.Printf("found target %v", f.Name)
			pi.Funcs = append(pi.Funcs, Function{
				Name:      f.Name,
				Comment:   toOneLine(f.Doc),
				Synopsis:  sanitizeSynopsis(f),
				IsError:   typ == mgTypes.ErrorType || typ == mgTypes.ContextErrorType,
				IsContext: typ == mgTypes.ContextVoidType || typ == mgTypes.ContextErrorType,
			})
		} else {
			debug.Printf("skipping function with invalid signature func %s(%v)(%v)", f.Name, fieldNames(f.Decl.Type.Params), fieldNames(f.Decl.Type.Results))
		}
	}

	hasDupes, names := checkDupes(pi)
	if hasDupes {
		msg := "Build targets must be case insensitive, thus the following targets conflict:\n"
		for _, v := range names {
			if len(v) > 1 {
				msg += "  " + strings.Join(v, ", ") + "\n"
			}
		}
		return nil, errors.New(msg)
	}

	setDefault(p, pi)
	setAliases(p, pi)

	return pi, nil
}

func fieldNames(flist *ast.FieldList) string {
	if flist == nil {
		return ""
	}
	list := flist.List
	if len(list) == 0 {
		return ""
	}
	args := make([]string, 0, len(list))
	for _, f := range list {
		names := make([]string, 0, len(f.Names))
		for _, n := range f.Names {
			if n.Name != "" {
				names = append(names, n.Name)
			}
		}
		nms := strings.Join(names, ", ")
		if nms != "" {
			nms += " "
		}
		args = append(args, nms+fmt.Sprint(f.Type))
	}
	return strings.Join(args, ", ")
}

// checkDupes checks a package for duplicate target names.
func checkDupes(info *PkgInfo) (hasDupes bool, names map[string][]string) {
	names = map[string][]string{}
	lowers := map[string]bool{}
	for _, f := range info.Funcs {
		low := strings.ToLower(f.Name)
		if f.Receiver != "" {
			low = strings.ToLower(f.Receiver) + ":" + low
		}
		if lowers[low] {
			hasDupes = true
		}
		lowers[low] = true
		names[low] = append(names[low], f.Name)
	}
	return hasDupes, names
}

// sanitizeSynopsis sanitizes function Doc to create a summary.
func sanitizeSynopsis(f *doc.Func) string {
	synopsis := doc.Synopsis(f.Doc)

	// If the synopsis begins with the function name, remove it. This is done to
	// not repeat the text.
	// From:
	// clean	Clean removes the temporarily generated files
	// To:
	// clean 	removes the temporarily generated files
	if syns := strings.Split(synopsis, " "); strings.EqualFold(f.Name, syns[0]) {
		return strings.Join(syns[1:], " ")
	}

	return synopsis
}

func setDefault(p *doc.Package, pi *PkgInfo) {
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
					pi.DefaultIsContext = f.IsContext
					pi.DefaultFunc = f
					return
				}
			}
			log.Println("warning: default declaration does not reference a mage target")
		}
	}
}

func setAliases(p *doc.Package, pi *PkgInfo) {
	for _, v := range p.Vars {
		for x, name := range v.Names {
			if name != "Aliases" {
				continue
			}
			spec, ok := v.Decl.Specs[x].(*ast.ValueSpec)
			if !ok {
				log.Println("warning: aliases declaration is not a value")
				return
			}
			if len(spec.Values) != 1 {
				log.Println("warning: aliases declaration has multiple values")
			}
			comp, ok := spec.Values[0].(*ast.CompositeLit)
			if !ok {
				log.Println("warning: aliases declaration is not a map")
				return
			}
			pi.Aliases = make(map[string]string)
			for _, elem := range comp.Elts {
				kv, ok := elem.(*ast.KeyValueExpr)
				if !ok {
					log.Println("warning: alias declaration is not a map element")
					return
				}
				k, ok := kv.Key.(*ast.BasicLit)
				if !ok || k.Kind != token.STRING {
					log.Println("warning: alias is not a string")
					return
				}
				v, ok := kv.Value.(*ast.Ident)
				if !ok {
					log.Println("warning: alias target is not a function")
					return
				}
				alias := strings.Trim(k.Value, "\"")
				valid := false
				for _, f := range pi.Funcs {
					valid = valid || f.Name == v.Name
				}
				if !valid {
					log.Printf("warning: alias declaration (%s) does not reference a mage target", alias)
				}
				pi.Aliases[alias] = v.Name
			}
			return
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
		return nil, fmt.Errorf("failed to parse directory: %v", err)
	}

	for name, pkg := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("no non-test packages found in %s", path)
}

func hasContextParam(ft *ast.FuncType) bool {
	if ft.Params.NumFields() != 1 {
		return false
	}
	ret := ft.Params.List[0]
	sel, ok := ret.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	if pkg.Name != "context" {
		return false
	}
	return sel.Sel.Name == "Context"
}

func hasVoidReturn(ft *ast.FuncType) bool {
	res := ft.Results
	return res.NumFields() == 0
}

func hasErrorReturn(ft *ast.FuncType) bool {
	res := ft.Results
	if res.NumFields() != 1 {
		return false
	}
	ret := res.List[0]
	if len(ret.Names) > 1 {
		return false
	}
	return fmt.Sprint(ret.Type) == "error"
}

func voidOrError(ft *ast.FuncType) mgTypes.FuncType {
	if hasContextParam(ft) {
		if hasVoidReturn(ft) {
			return mgTypes.ContextVoidType
		}
		if hasErrorReturn(ft) {
			return mgTypes.ContextErrorType
		}
	}
	if ft.Params.NumFields() == 0 {
		if hasVoidReturn(ft) {
			return mgTypes.VoidType
		}
		if hasErrorReturn(ft) {
			return mgTypes.ErrorType
		}
	}
	return mgTypes.InvalidType
}

func toOneLine(s string) string {
	return strings.TrimSpace(strings.Replace(s, "\n", " ", -1))
}
