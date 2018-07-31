+++
title = "Zero Install Option"
weight = 40
+++


Don't want to depend on another binary in your environment?  You can run mage
directly out of your vendor directory (or GOPATH) with `go run`.

Just save a file like this (I'll call it `mage.go`, but it can be named
anything. Note that the build tag is *not* `+build mage`.  Mage will create its own main file, so we need this one to be excluded from when your magefiles are compiled.  

Now you can `go run mage.go <target>` and it'll work just as if you ran `mage
<target>`

```go
// +build ignore

package main

import (
	"os"
	"github.com/magefile/mage/mage"
)

func main() { os.Exit(mage.Main()) }
```

Note that because of the peculiarities of `go run`, if you run this way, go run
will only ever exit with an error code of 0 or 1.  If mage exits with error code
99, for example, `go run` will print out `exit status 99" and then exit with
error code 1.  Why?  Ask the go team.  I've tried to get them to fix it, and
they won't.

If you are using `dep` for managing Go dependencies, it is necessary to add the mage import path to
a `required` clause in `Gopkg.toml` to prevent it being subsequently removed due to lack of
perceived use thanks for the `+build ignore` tag - for example:

```toml
required = ["github.com/magefile/mage/mage"]
```

## Use Mage as a library

All of mage's functionality is accessible as a compile-in library.  Checkout
[godoc.org/github.com/magefile/mage/mage](https://godoc.org/github.com/magefile/mage/mage)
for full details.

Fair warning, the API of mage/mage may change, so be sure to use vendoring.

## Compiling a static binary

If your tasks are not related to compiling Go code, it can be useful to compile a binary which has
the mage execution runtime and the tasks compiled in such that it can be run on another machine
without requiring installation of dependencies. To do so, pass the output path to the compile flag.
like this:

```
$ mage -compile ./static-output
```
