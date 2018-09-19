+++
title = "Dependencies"
weight = 30
+++

Mage supports a makefile-style tree of dependencies using the helper library
[github.com/magefile/mage/mg](https://godoc.org/github.com/magefile/mage/mg). To
declare dependencies, pass any number of dependent functions of the following
types:
```go
func()
func() error
func(ctx context.Context)
func(ctx context.Context) error
```
(they may be targets, but do not need to be and do not have to be exported) to
`mg.Deps()`, and the Deps function will not return until all declared
dependencies have been run (and any dependencies they have are run).
Dependencies are guaranteed to run exactly once in a single execution of mage,
so if two of your dependencies both depend on the same function, it is still
guaranteed to be run only once, and both funcs that depend on it will not
continue until it has been run. 

## Parallelism

If run with `mg.Deps` or `mg.CtxDeps`, dependencies are run in their own
goroutines, so they are parellelized as much as possible given the dependency
tree ordering restrictions.  If run with `mg.SerialDeps` or `mg.SerialCtxDeps`,
the dependencies are run serially, though each dependency or sub-dependency will
still only ever be run once. 

## Contexts and Cancellation

Dependencies that have a context.Context argument will be passed a context,
either a default context if passed into `mg.Deps` or `mg.SerialDeps`, or the one
passed into `mg.CtxDeps` or `mg.SerialCtxDeps`.  The default context, which is
also passed into [targets](/targets) with a context argument, will be cancelled
when and if the timeout specified on the command line is hit.

### Example Dependencies

```go
func Build() {
    mg.Deps(f, g)
    fmt.Println("Build running")
}

func f() {
    mg.Deps(h)
    fmt.Println("f running")
}

func g() {
    mg.Deps(h)
    fmt.Println("g running")
}

func h() {
    fmt.Println("h running")
}
```

Running `mage build` will produce the following output:

```plain
h running
g running
f running
Build running
```

Note that since f and g do not depend on each other, and they're running in
their own goroutines, their order is non-deterministic, other than they are
guaranteed to run after h has finished, and before Build continues.
