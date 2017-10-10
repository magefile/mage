+++
title = "Dependencies"
weight = 30
+++

Mage supports a makefile-style tree of dependencies using the helper library
[github.com/magefile/mage/mg](https://godoc.org/github.com/magefile/mage/mg). To
declare dependencies, pass any number of dependent functions (either `func()` or
`func() error` - they do not have to be exported) to `mg.Deps()`, and the Deps
function will not return until all declared dependencies have been run (and any
dependencies they have are run).  Dependencies are guaranteed to run exactly
once in a single execution of mage, so if two of your dependencies both depend
on the same function, it is still guaranteed to be run only once, and both funcs
that depend on it will not continue until it has been run. Dependencies are run
in their own goroutines, so they are parellelized as much as possible given the
dependency tree ordering restrictions.

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
    mg.Deps(g)
    fmt.Println("g running")
}

func h() {
    fmt.Println("h running")
}
```

Running `mage build` will produce the following output:

```
h running
g running
f running
Build running
```

Note that since f and g do not depend on each other, and they're running in
their own goroutines, their order is non-deterministic, other than they are
guaranteed to run after h has finished, and before Build continues.
