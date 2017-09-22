# mage

Mage is a Makefile replacement for Go.

## Usage
```
mage [options] [target]
Options:
  -f    force recreation of compiled magefile
  -h    show this help
  -l    list mage targets in this directory
  -v    show verbose output when running mage targets
```

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
magefile to exit with an exit code of 1.  Comments on the target function will
become documentation accessible by running `mage -l` which will list all the
build targets in this directory with the first sentence from their docs, or
`mage -h <target>` which will show the full comment from the docs on the
function.

## Full Example

```go
// +build mage

package main

import "os"

// A build target is any exported function with zero args and either no return
// or an error return.  If a target has an error return and returns an non-nil
// error, mage will print that error to stdout and return with an exit code of 1.
func Install() error {

}

// This comment will be the short help text shown with mage -l
// The rest of the comment is long help text that will be shown with mage -h <target>
func Target() {

}

// any function that lowercases to "build" becomes the default target for when 
// no target is specified.
func Build() { 
    // using os.Exit will do what you expect.
    os.Exit(99)
}


// Because build targets are case insensitive, you may not have two build targets
// that lowercase to the same text.  This would give you an error when you tried
// to run the magefile:
// func BUILD() {}


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

