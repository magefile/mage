package mg

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

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

// DepsWithContext runs the given functions as dependencies of the calling function.
// Dependencies must only be of type: github.com/magefile/mage/types.FuncType.  The function calling Deps
// is guaranteed that all dependent functions will be run exactly once when Deps
// returns.  Dependent functions may in turn declare their own dependencies
// using Deps. Each dependency is run in their own goroutines. Each function is
// given the context provided if the function prototype allows for it.
func DepsWithContext(ctx context.Context, fns ...interface{}) {
	for _, f := range fns {
		switch f.(type) {
		case func(), func() error, func(context.Context), func(context.Context) error:
			// ok
		default:
			panic(fmt.Errorf("Invalid type for dependent function: %T. Dependencies must be func(), func() error, func(context.Context) or func(context.Context) error", f))
		}
	}
	mu := &sync.Mutex{}
	var errs []string
	var exit int
	wg := &sync.WaitGroup{}
	for _, f := range fns {
		fn := addDep(ctx, f)
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

// Deps runs the given functions with the default runtime context
func Deps(fns ...interface{}) {
	ctx, _ := GetContext()
	DepsWithContext(ctx, fns...)
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

func addDep(ctx context.Context, f interface{}) *onceFun {
	var fn func(context.Context) error
	switch f := f.(type) {
	case func():
		fn = func(ctx context.Context) error { f(); return nil }
	case func() error:
		fn = func(ctx context.Context) error { return f() }
	case func(context.Context):
		fn = func(ctx context.Context) error { f(ctx); return nil }
	case func(context.Context) error:
		fn = f
	default:
		panic("attempted to add a dep that did not match required type")
	}

	n := name(f)
	of := onces.LoadOrStore(n, &onceFun{
		fn:  fn,
		ctx: ctx,
	})
	return of
}

func name(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

type onceFun struct {
	once sync.Once
	fn   func(context.Context) error
	ctx  context.Context
}

func (o *onceFun) run() error {
	var err error
	o.once.Do(func() {
		err = o.fn(o.ctx)
	})
	return err
}
