+++
title = "Targets"
weight = 10
+++
Any exported function that is either `func()` or `func() error` is considered a
mage target.  A target is effectively a subcommand of mage while running mage in
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

Currently only a single target may be run at a single time.  Attempting to run
multiple targets from a single invocation of mage will result in an error.  This
may change in the future.
