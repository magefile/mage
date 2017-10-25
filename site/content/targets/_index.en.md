+++
title = "Targets"
weight = 10
+++
A target is any exported function that is one of the following types:
```
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
from the docs on the function.

A target may be designated the default target, which is run when the user runs
`mage` with no target specified. To denote the default, create a `var Default =
<targetname>`  If no default target is specified, running `mage` with no target
will print the list of targets, like `mage -l`.

## Multiple Targets

Multiple targets can be specified as args to Mage, for example `mage foo bar
baz`.  Targets will be run serially, from left to right (so in thise case, foo,
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