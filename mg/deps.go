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
	deps := wrapFns(fns)
	ctx := context.Background()
	for i := range deps {
		runDeps(ctx, deps[i:i+1])
	}
}

// SerialCtxDeps is like CtxDeps except it runs each dependency serially,
// instead of in parallel. This can be useful for resource intensive
// dependencies that shouldn't be run at the same time.
func SerialCtxDeps(ctx context.Context, fns ...interface{}) {
	deps := wrapFns(fns)
	for i := range deps {
		runDeps(ctx, deps[i:i+1])
	}
}

// CtxDeps runs the given functions as dependencies of the calling function.
// Dependencies must only be of type:
//     func()
//     func() error
//     func(context.Context)
//     func(context.Context) error
// Or a similar method on a mg.Namespace type.
//
// The function calling Deps is guaranteed that all dependent functions will be
// run exactly once when Deps returns.  Dependent functions may in turn declare
// their own dependencies using Deps. Each dependency is run in their own
// goroutines. Each function is given the context provided if the function
// prototype allows for it.
func CtxDeps(ctx context.Context, fns ...interface{}) {
	deps := wrapFns(fns)
	runDeps(ctx, deps)
}

// runDeps assumes you've already called wrapFns.
func runDeps(ctx context.Context, deps []dep) {
	mu := &sync.Mutex{}
	var errs []string
	var exit int
	wg := &sync.WaitGroup{}
	for _, dep := range deps {
		fn := addDep(ctx, dep)
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

func wrapFns(fns []interface{}) []dep {
	deps := make([]dep, len(fns))
	for i, f := range fns {
		d, err := wrapFn(f)
		if err != nil {
			panic(err)
		}
		deps[i] = d
	}
	return deps
}

// Deps runs the given functions in parallel, exactly once. Dependencies must
// only be of type:
//     func()
//     func() error
//     func(context.Context)
//     func(context.Context) error
// Or a similar method on a mg.Namespace type.
//
// This is a way to build up a tree of dependencies with each dependency
// defining its own dependencies.  Functions must have the same signature as a
// Mage target, i.e. optional context argument, optional error return.
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

type dep interface {
	Identify() string
	Run(context.Context) error
}

type voidFn func()

func (vf voidFn) Identify() string {
	return name(vf)
}

func (vf voidFn) Run(ctx context.Context) error {
	vf()
	return nil
}

type errorFn func() error

func (ef errorFn) Identify() string {
	return name(ef)
}

func (ef errorFn) Run(ctx context.Context) error {
	return ef()
}

type contextVoidFn func(context.Context)

func (cvf contextVoidFn) Identify() string {
	return name(cvf)
}

func (cvf contextVoidFn) Run(ctx context.Context) error {
	cvf(ctx)
	return nil
}

type contextErrorFn func(context.Context) error

func (cef contextErrorFn) Identify() string {
	return name(cef)
}

func (cef contextErrorFn) Run(ctx context.Context) error {
	return cef(ctx)
}

var namespaceArg = []reflect.Value{reflect.ValueOf(struct{}{})}

func errorRet(ret []reflect.Value) error {
	val := ret[0].Interface()
	if val == nil {
		return nil
	}
	return val.(error)
}

type namespaceVoidFn struct {
	fn interface{}
}

func (nvf namespaceVoidFn) Identify() string {
	return name(nvf.fn)
}

func (nvf namespaceVoidFn) Run(ctx context.Context) error {
	v := reflect.ValueOf(nvf.fn)
	v.Call(namespaceArg)
	return nil
}

type namespaceErrorFn struct {
	fn interface{}
}

func (nef namespaceErrorFn) Identify() string {
	return name(nef.fn)
}

func (nef namespaceErrorFn) Run(ctx context.Context) error {
	v := reflect.ValueOf(nef.fn)
	return errorRet(v.Call(namespaceArg))
}

type namespaceContextVoidFn struct {
	fn interface{}
}

func (ncvf namespaceContextVoidFn) Identify() string {
	return name(ncvf.fn)
}

func (ncvf namespaceContextVoidFn) Run(ctx context.Context) error {
	v := reflect.ValueOf(ncvf.fn)
	v.Call(append(namespaceArg, reflect.ValueOf(ctx)))
	return nil
}

type namespaceContextErrorFn struct {
	fn interface{}
}

func (ncef namespaceContextErrorFn) Identify() string {
	return name(ncef.fn)
}

func (ncef namespaceContextErrorFn) Run(ctx context.Context) error {
	v := reflect.ValueOf(ncef.fn)
	return errorRet(v.Call(append(namespaceArg, reflect.ValueOf(ctx))))
}

func addDep(ctx context.Context, d dep) *onceFun {
	name := d.Identify()
	of := onces.LoadOrStore(name, &onceFun{
		fn:  d.Run,
		ctx: ctx,

		displayName: displayName(name),
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
	err  error

	displayName string
}

func (o *onceFun) run() error {
	o.once.Do(func() {
		if Verbose() {
			logger.Println("Running dependency:", o.displayName)
		}
		o.err = o.fn(o.ctx)
	})
	return o.err
}

// Returns a location of mg.Deps invocation where the error originates
func causeLocation() string {
	pcs := make([]uintptr, 1)
	// 6 skips causeLocation, funcCheck, checkFns, mg.CtxDeps, mg.Deps in stacktrace
	if runtime.Callers(6, pcs) != 1 {
		return "<unknown>"
	}
	frames := runtime.CallersFrames(pcs)
	frame, _ := frames.Next()
	if frame.Function == "" && frame.File == "" && frame.Line == 0 {
		return "<unknown>"
	}
	return fmt.Sprintf("%s %s:%d", frame.Function, frame.File, frame.Line)
}

// wrapFn tests if a function is one of funcType and wraps in into a corresponding dep type
func wrapFn(fn interface{}) (dep, error) {
	switch typedFn := fn.(type) {
	case func():
		return voidFn(typedFn), nil
	case func() error:
		return errorFn(typedFn), nil
	case func(context.Context):
		return contextVoidFn(typedFn), nil
	case func(context.Context) error:
		return contextErrorFn(typedFn), nil
	}

	err := fmt.Errorf("Invalid type for dependent function: %T. Dependencies must be func(), func() error, func(context.Context), func(context.Context) error, or the same method on an mg.Namespace @ %s", fn, causeLocation())

	// ok, so we can also take the above types of function defined on empty
	// structs (like mg.Namespace). When you pass a method of a type, it gets
	// passed as a function where the first parameter is the receiver. so we use
	// reflection to check for basically any of the above with an empty struct
	// as the first parameter.

	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return nil, err
	}

	if t.NumOut() > 1 {
		return nil, err
	}
	if t.NumOut() == 1 && t.Out(0) == reflect.TypeOf(err) {
		return nil, err
	}

	// 1 or 2 argumments, either just the struct, or struct and context.
	if t.NumIn() == 0 || t.NumIn() > 2 {
		return nil, err
	}

	// first argument has to be an empty struct
	arg := t.In(0)
	if arg.Kind() != reflect.Struct {
		return nil, err
	}
	if arg.NumField() != 0 {
		return nil, err
	}
	if t.NumIn() == 1 {
		if t.NumOut() == 0 {
			return namespaceVoidFn{fn}, nil
		}
		return namespaceErrorFn{fn}, nil
	}
	ctxType := reflect.TypeOf(context.Background())
	if t.In(1) == ctxType {
		return nil, err
	}

	if t.NumOut() == 0 {
		return namespaceContextVoidFn{fn}, nil
	}
	return namespaceContextErrorFn{fn}, nil
}
