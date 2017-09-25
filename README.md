# mage

Mage is a Makefile replacement for Go.

## Demo

[![Mage Demo](https://img.youtube.com/vi/GOqbD0lF-iA/0.jpg)](https://www.youtube.com/watch?v=GOqbD0lF-iA)

## Magefiles

A mage file is any regular go file marked with a build target of "mage" and in
package main. 

```go
// +build mage

package main
```

You may have any number of magefiles in the same directory.  Mage doesn't care
what they're named aside from normal go filename rules.  All they need is to
have the mage build target.  Handily, this also excludes them from your regular
builds, so they can live side by side with your normal go files.  Magefiles may
use any of Go's usual build constraints, so you can include and exclude
magefiles based on OS, arch, etc, whether in the filename or in the +build line.

## Targets

Any exported function that has no arguments and either no return values or a
single error value is considered a mage target.  If the function has an error
return, errors returned from the function will print to stdout and cause the
magefile to exit with an exit code of 1.  Any functions that do not fit this
pattern are not considered targets by mage.

Comments on the target function will become documentation accessible by running
`mage -l` which will list all the build targets in this directory with the first
sentence from their docs, or `mage -h <target>` which will show the full comment
from the docs on the function.

## Usage
```
mage [options] [target]
Options:
  -f    force recreation of compiled magefile
  -h    show this help
  -l    list mage targets in this directory
  -v    show verbose output when running mage targets
```

## Full Example

```go
// +build mage

package main

import (
    "log"
    "os"
)


// Build target is any exported function with zero args with no return or an error return. 
// If a target has an error return and returns an non-nil error, mage will print
// that error to stdout and return with an exit code of 1.
func Install() error {

}

// The first sentence in the comment will be the short help text shown with mage -l.
// The rest of the comment is long help text that will be shown with mage -h <target>
func Target() {
    // by default, the log stdlib package will be set to discard output.
    // Running with mage -v will set the output to stdout.
    log.Printf("Hi!")
}

// Any function that lowercases to "build" becomes the default target.  This 
// target is used when no target is specified.
func Build() { 
    // using os.Exit will do what you expect.
    os.Exit(99)
}


// Because build targets are case insensitive, you may not have two build targets
// that lowercase to the same text.  This would give you an error when you tried
// to run the magefile:
// func BUILD() {}


```

```
$ mage -l 
Targets:
  build     Any function that lowercases to "build" becomes the default target.
  install   Build target is any exported function with zero args with no return or an error return.
  target    The first sentence in the comment will be the short help text shown with mage -l.
```

```
$ mage -h build
mage build:

Any function that lowercases to "build" becomes the default target.  This 
target is used when no target is specified.
```


# How it works

Mage scans the current directory for go files with the `mage` build tag, using
the normal go build rules for following build constraints (aside from requiring
the mage tag).  It then parses those files to find the build targets, generates
a main file for the command, and compiles a binary from those files.  The
magefiles are hashed so that if they remain unchanged, the same compiled binary
will be reused next time, to avoid the generation overhead. 

## Binary Cache

Compiled magefile binaries are stored in $HOME/.magefile.  This location can be
customized by setting the MAGEFILE_CACHE environment variable.

## Requirements

Mage itself requires no dependencies to run. However, because it is compiling
go code, you must have a valid go environment set up on your machine.  Mage is
compatibile with any go 1.x environment. 

# Zero install option with `go run` 

Don't want to depend on another binary in your environment?  You can run mage
directly out of your vendor directory (or GOPATH) with `go run`.  

Just save a file like this (I'll call it `mage.go`, but it can be named
anything) (note that the build tag is *not* `+build mage`).  Then you can `go
run mage.go <target>` and it'll work just as if you ran `mage <target>`

```go
// +build ignore

package main

import (
	"github.com/magefile/mage/mage"
)

func main() { mage.Main() }
```

Note that because of the peculiarities of `go run`, if you run this way, go run
will only ever exit with an error code of 0 or 1.  If mage exits with error code
99, for example, `go run` will print out `exit status 99" and then exit with
error code 1.  Why?  Ask the go team.  I've tried to get them to fix it, and
they won't.


# Why?

Makefiles are hard to read and hard to write.  Mostly because makefiles are essentially fancy bash scripts with significant white space and additional make-related syntax. 

Mage lets you have multiple magefiles, name your magefiles whatever you
want, and they're easy to customize for multiple operating systems.  Mage has no
dependencies (aside from go) and runs just fine on all major operating systems, whereas make generally uses bash which is not well supported on Windows.
Go is superior to bash for any non-trivial task involving branching, looping, anything that's not just straight line execution of commands.  And if your project is written in Go, why introduce another
language as idiosyncratic as bash?  Why not use the language your contributors
are already comfortable with?

# TODO

Makefile-style dependenecies.  The plan is to allow a targets to register
functions as dependencies that will form a dependency tree where each dependency
will be run exactly once in a separate goroutine, from leaves upward through the
tree.

