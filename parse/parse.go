package parse

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var debug = log.New(ioutil.Discard, "DEBUG: ", log.Ltime|log.Lmicroseconds)

// EnableDebug turns on debug logging.
func EnableDebug() {
	debug.SetOutput(os.Stderr)
}

// PkgInfo contains inforamtion about a package of files according to mage's
// parsing rules.
type PkgInfo struct {
	Description string
	Funcs       []Function
	DefaultFunc Function
	Aliases     map[string]Function
	Imports     map[string]string
}

// Function represented a job function from a mage file
type Function struct {
	PkgAlias  string
	Package   string
	Name      string
	Receiver  string
	IsError   bool
	IsContext bool
	Synopsis  string
	Comment   string
}

// TargetName returns the name of the target as it should appear when used from
// the mage cli.  It is always lowercase.
func (f Function) TargetName() string {
	var names []string
	pkg := f.PkgAlias

	// package alias of "." means use it like a dot import, i.e. don't namespace
	// with the package alias name.
	if pkg == "." {
		pkg = ""
	}
	for _, s := range []string{pkg, f.Receiver, f.Name} {
		if s != "" {
			names = append(names, s)
		}
	}
	return strings.ToLower(strings.Join(names, ":"))
}

// ExecCode returns code for the template switch to run the target.
// It wraps each target call to match the func(context.Context) error that
// runTarget requires.
func (f Function) ExecCode() (string, error) {
	name := f.Name
	if f.Receiver != "" {
		name = f.Receiver + "{}." + name
	}
	if f.Package != "" {
		name = f.Package + "." + name
	}

	if f.IsContext && f.IsError {
		out := `
			wrapFn := func(ctx context.Context) error {
				return %s(ctx)
			}
			err := runTarget(wrapFn)`[1:]
		return fmt.Sprintf(out, name), nil
	}
	if f.IsContext && !f.IsError {
		out := `
			wrapFn := func(ctx context.Context) error {
				%s(ctx)
				return nil
			}
			err := runTarget(wrapFn)`[1:]
		return fmt.Sprintf(out, name), nil
	}
	if !f.IsContext && f.IsError {
		out := `
			wrapFn := func(ctx context.Context) error {
				return %s()
			}
			err := runTarget(wrapFn)`[1:]
		return fmt.Sprintf(out, name), nil
	}
	if !f.IsContext && !f.IsError {
		out := `
			wrapFn := func(ctx context.Context) error {
				%s()
				return nil
			}
			err := runTarget(wrapFn)`[1:]
		return fmt.Sprintf(out, name), nil
	}
	return "", fmt.Errorf("Error formatting ExecCode code for %#v", f)
}

// Package parses a package.  If files is non-empty, it will only parse the files given.
func Package(path string, files []string) (*PkgInfo, error) {
	start := time.Now()
	defer func() {
		debug.Println("time parse Magefiles:", time.Since(start))
	}()
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
			if typ := funcType(f.Decl.Type); typ != invalidType {
				pi.Funcs = append(pi.Funcs, Function{
					Name:      f.Name,
					Receiver:  f.Recv,
					Comment:   toOneLine(f.Doc),
					Synopsis:  sanitizeSynopsis(f),
					IsError:   typ == errorType || typ == contextErrorType,
					IsContext: typ == contextVoidType || typ == contextErrorType,
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
		if typ := funcType(f.Decl.Type); typ != invalidType {
			debug.Printf("found target %v", f.Name)
			pi.Funcs = append(pi.Funcs, Function{
				Name:      f.Name,
				Comment:   toOneLine(f.Doc),
				Synopsis:  sanitizeSynopsis(f),
				IsError:   typ == errorType || typ == contextErrorType,
				IsContext: typ == contextVoidType || typ == contextErrorType,
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
	if err := setImports(p, pi); err != nil {
		return nil, err
	}

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

			var name string
			var receiver string
			switch v := spec.Values[0].(type) {
			case *ast.Ident:
				name = v.Name
			case *ast.SelectorExpr:
				name = v.Sel.Name
				receiver = fmt.Sprintf("%s", v.X)
			default:
				log.Printf("warning: target for Default %s is not a function", spec.Values[0])
			}
			for _, f := range pi.Funcs {
				if f.Name == name && f.Receiver == receiver {
					pi.DefaultFunc = f
					return
				}
			}
			log.Println("warning: default declaration does not reference a mage target")
		}
	}
}

func setImports(p *doc.Package, pi *PkgInfo) error {
	pi.Imports = map[string]string{}
	for _, v := range p.Vars {
		for x, name := range v.Names {
			if name != "MageImports" {
				continue
			}
			spec := v.Decl.Specs[x].(*ast.ValueSpec)
			if len(spec.Values) != 1 {
				return errors.New("MageImports declaration has multiple values")
			}
			v, ok := spec.Values[0].(*ast.CompositeLit)
			if !ok {
				return errors.New("MageImports must be a composite literal map")
			}
			t, ok := v.Type.(*ast.MapType)
			if !ok {
				return errors.New("MageImports is not a map")
			}
			k, ok := t.Key.(*ast.Ident)
			if !ok || k.Name != "string" {
				return errors.New("MageImports must be a map[string]string")
			}
			val, ok := t.Value.(*ast.Ident)
			if !ok || val.Name != "string" {
				return errors.New("MageImports must be a map[string]string")
			}

			for _, e := range v.Elts {
				kv, ok := e.(*ast.KeyValueExpr)
				if !ok {
					return errors.New(`elements of MageImports must be "name" : "importpath"`)
				}
				k, ok := kv.Key.(*ast.BasicLit)
				if !ok {
					return errors.New(`elements of MageImports must be "name" : "importpath"`)
				}
				v, ok := kv.Value.(*ast.BasicLit)
				if !ok {
					return errors.New(`elements of MageImports must be "name" : "importpath"`)
				}
				if !strings.HasPrefix(k.Value, `"`) || !strings.HasSuffix(k.Value, `"`) {
					return errors.New(`elements of MageImports must be "name" : "importpath"`)
				}
				if !strings.HasPrefix(v.Value, `"`) || !strings.HasSuffix(v.Value, `"`) {
					return errors.New(`elements of MageImports must be "name" : "importpath"`)
				}
				alias := strings.Trim(k.Value, `"`)
				path := strings.Trim(v.Value, `"`)
				pi.Imports[alias] = path
			}
		}
	}
	return nil
}

func outputDebug(cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}
	debug.Println("running", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	c.Stderr = errbuf
	c.Stdout = buf
	if err := c.Run(); err != nil {
		debug.Print("error running '", cmd, strings.Join(args, " "), "': ", err, ": ", errbuf)
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
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
			pi.Aliases = map[string]Function{}
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

				var name string
				var receiver string
				switch v := kv.Value.(type) {
				case *ast.Ident:
					name = v.Name
				case *ast.SelectorExpr:
					name = v.Sel.Name
					receiver = fmt.Sprintf("%s", v.X)
				default:
					log.Printf("warning: target for alias %s is not a function", k.Value)
				}
				alias := strings.Trim(k.Value, "\"")
				valid := false
				for _, f := range pi.Funcs {
					if f.Name == name && f.Receiver == receiver {
						pi.Aliases[alias] = f
						valid = true
						break
					}
				}
				if !valid {
					log.Printf("warning: alias declaration (%s) does not reference a mage target", alias)
				}
			}
			return
		}
	}
}

// getPackage returns the non-test package at the given path.
func getPackage(path string, files []string, fset *token.FileSet) (*ast.Package, error) {
	var filter func(f os.FileInfo) bool
	if len(files) > 0 {
		fm := make(map[string]bool, len(files))
		for _, f := range files {
			fm[f] = true
		}

		filter = func(f os.FileInfo) bool {
			return fm[f.Name()]
		}
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

type functype int

const (
	invalidType functype = iota
	voidType
	errorType
	contextVoidType
	contextErrorType
)

func funcType(ft *ast.FuncType) functype {
	if hasContextParam(ft) {
		if hasVoidReturn(ft) {
			return contextVoidType
		}
		if hasErrorReturn(ft) {
			return contextErrorType
		}
	}
	if ft.Params.NumFields() == 0 {
		if hasVoidReturn(ft) {
			return voidType
		}
		if hasErrorReturn(ft) {
			return errorType
		}
	}
	return invalidType
}

func toOneLine(s string) string {
	return strings.TrimSpace(strings.Replace(s, "\n", " ", -1))
}
