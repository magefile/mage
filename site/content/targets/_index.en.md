+++
title = "Targets"
weight = 10
+++
A target is any exported function that is one of the following types:
```go
func()
func() error 
func(context.Context)
func(context.Context) error
```
A target is effectively a subcommand of mage while running mage in
this directory.  i.e. you can run a target by running `mage <target>`

If the function has an error return, errors returned from the function will
print to stdout and cause the magefile to exit with an exit code of 1.  Any
functions that do not fit this pattern are not considered targets by mage.

Comments on the target function will become documentation accessible by running
`mage -l` which will list all the build targets in this directory with the first
sentence from their docs, or `mage -h <target>` which will show the full comment
from the docs on the function, and a list of aliases if specified.

A target may be designated the default target, which is run when the user runs
`mage` with no target specified. To denote the default, create a `var Default =
<targetname>`  If no default target is specified, running `mage` with no target
will print the list of targets, like `mage -l`.

## Multiple Targets

Multiple targets can be specified as args to Mage, for example `mage foo bar
baz`.  Targets will be run serially, from left to right (so in this case, foo,
then once foo is done, bar, then once bar is done, baz).  Dependencies run using
mg.Deps will still only run once per mage execution, so if each of the targets
depend on the same function, that function will only be run once for all
targets.  If any target panics or returns an error, no later targets will be run.

## Contexts and Cancellation

A default context is passed into any target with a context argument.  This
context will have a timeout if mage was run with -t, and thus will cancel the
running targets and dependencies at that time.  To pass this context to
dependencies, use mg.CtxDeps(ctx, ...) to pass the context from the target to
its dependencies (and pass the context to sub-dependencies).  Dependencies run
with mg.Deps will not get the starting context, and thus will not be cancelled
when the timeout set with -t expires.

mg.CtxDeps will pass along whatever context you give it, so if you want to
modify the original context, or pass in your own, that will work like you expect
it to.

## Aliases

Target aliases can be specified using the following notation:

```go
var Aliases = map[string]interface{} {
  "i":     Install,
  "build": Install,
  "ls":    List,
}
```

The key is an alias and the value is a function identifier.
An alias can be used interchangeably with it's target.

## Namespaces

Namespaces are a way to group related commands, much like subcommands in a
normal application.   To define a namespace in your magefile, simply define an
exported named type of type `mg.Namespace`.  Then, every method on that type which
matches the normal target signature becomes a target under that namespace.

```go
import "github.com/magefile/mage/mg"

type Build mg.Namespace

// Builds the site using hugo.
func (Build) Site() error {
  return nil
}

// Builds the pdf docs.
func (Build) Docs() {}
```

To call a namespaced target, type it as `namespace:target`. For example, the
above would be called by typing

```plain
$ mage build:site
```

Similarly, the help for the target will show how it may be called:

```plain
$ mage -l

build:docs    Builds the pdf docs.
build:site    Builds the site using hugo.
```
