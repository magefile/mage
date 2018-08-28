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

	"github.com/magefile/mage/types"
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
	checkFns(fns)
	ctx := context.Background()
	for _, f := range fns {
		runDeps(ctx, f)
	}
}

// SerialCtxDeps is like CtxDeps except it runs each dependency serially,
// instead of in parallel. This can be useful for resource intensive
// dependencies that shouldn't be run at the same time.
func SerialCtxDeps(ctx context.Context, fns ...interface{}) {
	checkFns(fns)
	for _, f := range fns {
		runDeps(ctx, f)
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
	checkFns(fns)
	runDeps(ctx, fns...)
}

// runDeps assumes you've already called checkFns.
func runDeps(ctx context.Context, fns ...interface{}) {
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

func checkFns(fns []interface{}) {
	for _, f := range fns {
		if err := types.FuncCheck(f); err != nil {
			panic(err)
		}
	}
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

func addDep(ctx context.Context, f interface{}) *onceFun {
	var fn func(context.Context) error
	if fn = types.FuncTypeWrap(f); fn == nil {
		// should be impossible, since we already checked this
		panic("attempted to add a dep that did not match required type")
	}

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
