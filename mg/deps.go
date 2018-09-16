package mg

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// FuncType indicates a prototype of build job function
type FuncType int

// FuncTypes
const (
	InvalidType FuncType = iota
	VoidType
	ErrorType
	ContextVoidType
	ContextErrorType
	NamespaceVoidType
	NamespaceErrorType
	NamespaceContextVoidType
	NamespaceContextErrorType
)

var logger = log.New(os.Stderr, "", 0)

type onceMap struct {
	mu *sync.Mutex
	m  map[string]*onceFun
}

func (o *onceMap) LoadOrStore(s string, one *onceFun) *onceFun {
	defer o.mu.Unlock()
	o.mu.Lock()

	existing, ok := o.m[s]
	if ok {
		return existing
	}
	o.m[s] = one
	return one
}

var onces = &onceMap{
	mu: &sync.Mutex{},
	m:  map[string]*onceFun{},
}

// SerialDeps is like Deps except it runs each dependency serially, instead of
// in parallel. This can be useful for resource intensive dependencies that
// shouldn't be run at the same time.
func SerialDeps(fns ...interface{}) {
	types := checkFns(fns)
	ctx := context.Background()
	for i := range fns {
		runDeps(ctx, types[i:i], fns[i:i])
	}
}

// SerialCtxDeps is like CtxDeps except it runs each dependency serially,
// instead of in parallel. This can be useful for resource intensive
// dependencies that shouldn't be run at the same time.
func SerialCtxDeps(ctx context.Context, fns ...interface{}) {
	types := checkFns(fns)
	for i := range fns {
		runDeps(ctx, types[i:i], fns[i:i])
	}
}

// CtxDeps runs the given functions as dependencies of the calling function.
// Dependencies must only be of type: github.com/magefile/mage/types.FuncType.
// The function calling Deps is guaranteed that all dependent functions will be
// run exactly once when Deps returns.  Dependent functions may in turn declare
// their own dependencies using Deps. Each dependency is run in their own
// goroutines. Each function is given the context provided if the function
// prototype allows for it.
func CtxDeps(ctx context.Context, fns ...interface{}) {
	types := checkFns(fns)
	runDeps(ctx, types, fns)
}

// runDeps assumes you've already called checkFns.
func runDeps(ctx context.Context, types []FuncType, fns []interface{}) {
	mu := &sync.Mutex{}
	var errs []string
	var exit int
	wg := &sync.WaitGroup{}
	for i, f := range fns {
		fn := addDep(ctx, types[i], f)
		wg.Add(1)
		go func() {
			defer func() {
				if v := recover(); v != nil {
					mu.Lock()
					if err, ok := v.(error); ok {
						exit = changeExit(exit, ExitStatus(err))
					} else {
						exit = changeExit(exit, 1)
					}
					errs = append(errs, fmt.Sprint(v))
					mu.Unlock()
				}
				wg.Done()
			}()
			if err := fn.run(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprint(err))
				exit = changeExit(exit, ExitStatus(err))
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	if len(errs) > 0 {
		panic(Fatal(exit, strings.Join(errs, "\n")))
	}
}

func checkFns(fns []interface{}) []FuncType {
	types := make([]FuncType, len(fns))
	for i, f := range fns {
		t, err := FuncCheck(f)
		if err != nil {
			panic(err)
		}
		types[i] = t
	}
	return types
}

// Deps runs the given functions in parallel, exactly once.  This is a way to
// build up a tree of dependencies with each dependency defining its own
// dependencies.  Functions must have the same signature as a Mage target, i.e.
// optional context argument, optional error return.
func Deps(fns ...interface{}) {
	CtxDeps(context.Background(), fns...)
}

func changeExit(old, new int) int {
	if new == 0 {
		return old
	}
	if old == 0 {
		return new
	}
	if old == new {
		return old
	}
	// both different and both non-zero, just set
	// exit to 1. Nothing more we can do.
	return 1
}

func addDep(ctx context.Context, t FuncType, f interface{}) *onceFun {
	fn := FuncTypeWrap(t, f)

	n := name(f)
	of := onces.LoadOrStore(n, &onceFun{
		fn:  fn,
		ctx: ctx,

		displayName: displayName(n),
	})
	return of
}

func name(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func displayName(name string) string {
	splitByPackage := strings.Split(name, ".")
	if len(splitByPackage) == 2 && splitByPackage[0] == "main" {
		return splitByPackage[len(splitByPackage)-1]
	}
	return name
}

type onceFun struct {
	once sync.Once
	fn   func(context.Context) error
	ctx  context.Context

	displayName string
}

func (o *onceFun) run() error {
	var err error
	o.once.Do(func() {
		if Verbose() {
			logger.Println("Running dependency:", o.displayName)
		}
		err = o.fn(o.ctx)
	})
	return err
}

// FuncCheck tests if a function is one of FuncType
func FuncCheck(fn interface{}) (FuncType, error) {
	switch fn.(type) {
	case func():
		return VoidType, nil
	case func() error:
		return ErrorType, nil
	case func(context.Context):
		return ContextVoidType, nil
	case func(context.Context) error:
		return ContextErrorType, nil
	}

	err := fmt.Errorf("Invalid type for dependent function: %T. Dependencies must be func(), func() error, func(context.Context), func(context.Context) error, or the same method on an mg.Namespace.", fn)

	// ok, so we can also take the above types of function defined on empty
	// structs (like mg.Namespace). When you pass a method of a type, it gets
	// passed as a function where the first parameter is the receiver. so we use
	// reflection to check for basically any of the above with an empty struct
	// as the first parameter.

	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return InvalidType, err
	}

	if t.NumOut() > 1 {
		return InvalidType, err
	}
	if t.NumOut() == 1 && t.Out(0) == reflect.TypeOf(err) {
		return InvalidType, err
	}

	// 1 or 2 argumments, either just the struct, or struct and context.
	if t.NumIn() == 0 || t.NumIn() > 2 {
		return InvalidType, err
	}

	// first argument has to be an empty struct
	arg := t.In(0)
	if arg.Kind() != reflect.Struct {
		return InvalidType, err
	}
	if arg.NumField() != 0 {
		return InvalidType, err
	}
	if t.NumIn() == 1 {
		if t.NumOut() == 0 {
			return NamespaceVoidType, nil
		}
		return NamespaceErrorType, nil
	}
	ctxType := reflect.TypeOf(context.Background())
	if t.In(1) == ctxType {
		return InvalidType, err
	}

	if t.NumOut() == 0 {
		return NamespaceContextVoidType, nil
	}
	return NamespaceContextErrorType, nil
}

// FuncTypeWrap wraps a valid FuncType to FuncContextError
func FuncTypeWrap(t FuncType, fn interface{}) func(context.Context) error {
	switch f := fn.(type) {
	case func():
		return func(context.Context) error {
			f()
			return nil
		}
	case func() error:
		return func(context.Context) error {
			return f()
		}
	case func(context.Context):
		return func(ctx context.Context) error {
			f(ctx)
			return nil
		}
	case func(context.Context) error:
		return f
	}
	args := []reflect.Value{reflect.ValueOf(struct{}{})}
	switch t {
	case NamespaceVoidType:
		return func(context.Context) error {
			v := reflect.ValueOf(fn)
			v.Call(args)
			return nil
		}
	case NamespaceErrorType:
		return func(context.Context) error {
			v := reflect.ValueOf(fn)
			ret := v.Call(args)
			return ret[0].Interface().(error)
		}
	case NamespaceContextVoidType:
		return func(ctx context.Context) error {
			v := reflect.ValueOf(fn)
			v.Call(append(args, reflect.ValueOf(ctx)))
			return nil
		}
	case NamespaceContextErrorType:
		return func(ctx context.Context) error {
			v := reflect.ValueOf(fn)
			ret := v.Call(append(args, reflect.ValueOf(ctx)))
			return ret[0].Interface().(error)
		}
	default:
		panic(fmt.Errorf("Don't know how to deal with dep of type %T", fn))
	}
}
