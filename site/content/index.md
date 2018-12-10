+++
title = "Mage"
+++

## About

Mage is a make/rake-like build tool using Go.  You write plain-old go functions,
and Mage automatically uses them as Makefile-like runnable targets.


## Installation

Mage has no dependencies outside the Go standard library, and builds with Go 1.7
and above (possibly even lower versions, but they're not regularly tested). 

Install mage by running 

```plain
go get -u -d github.com/magefile/mage
cd $GOPATH/src/github.com/magefile/mage
go run bootstrap.go
```

This will download the code into your GOPATH, and then run the bootstrap script
to build mage with version infomation embedded in it.  A normal `go get`
(without -d) will build the binary correctly, but no version info will be
embedded.  If you've done this, no worries, just go to
$GOPATH/src/github.com/magefile/mage and run `mage install` or `go run
bootstrap.go` and a new binary will be created with the correct version
information.

The mage binary will be created in your $GOPATH/bin directory.

You may also install a binary release from our
[releases](https://github.com/magefile/mage/releases) page. 

## Example Magefile

```go
//+build mage

package main

// Runs dep ensure and then installs the binary.
func Build() error {
    if err := sh.Run("dep", "ensure"); err != nil {
        return err
    }
    return sh.Run("go", "install", "./...")
}
```

Run the above `Build` target by simply running `mage build` in the same directory as the magefile.

## Demo

{{< youtube Hoga60EF_1U >}}

## Discussion

Join the `#mage` channel on [gophers slack](https://gophers.slack.com/messages/general/) for discussion of usage, development, etc.


## Plugins

There are no plugins.  You don't need plugins.  It's just Go code.  You can
import whatever libraries you want.  Every library in the go ecosystem is a mage
plugin.  Every tool you use with Go can be used with Magefiles.

## Usage
```plain
mage [options] [target]

Mage is a make-like command runner.  See https://magefile.org for full docs.

Commands:
  -clean    clean out old generated binaries from CACHE_DIR
  -compile <string>
            output a static binary to the given path
  -init     create a starting template if no mage files exist
  -l        list mage targets in this directory
  -h        show this help
  -version  show version info for the mage binary

Options:
  -d <string> 
            run magefiles in the given directory (default ".")
  -debug    turn on debug messages
  -h        show description of a target
  -f        force recreation of compiled magefile
  -keep     keep intermediate mage files around after running
  -gocmd <string>
		    use the given go binary to compile the output (default: "go")
  -goos     sets the GOOS for the binary created by -compile (default: current OS)
  -goarch   sets the GOARCH for the binary created by -compile (default: current arch)
  -t <string>
            timeout in duration parsable format (e.g. 5m30s)
  -v        show verbose output when running mage targets
  ```

## Why?

Makefiles are hard to read and hard to write.  Mostly because makefiles are essentially fancy bash
scripts with significant white space and additional make-related syntax.

Mage lets you have multiple magefiles, name your magefiles whatever you want, and they're easy to
customize for multiple operating systems.  Mage has no dependencies (aside from go) and runs just
fine on all major operating systems, whereas make generally uses bash which is not well supported on
Windows.  Go is superior to bash for any non-trivial task involving branching, looping, anything
that's not just straight line execution of commands.  And if your project is written in Go, why
introduce another language as idiosyncratic as bash?  Why not use the language your contributors are
already comfortable with?

## Code

[https://github.com/magefile/mage](https://github.com/magefile/mage)

## Projects that build with Mage

[![Hugo](/images/hugo.png)](https://github.com/gohugoio/hugo) [![Gnorm](/images/gnorm.png)](https://github.com/gnormal/gnorm)
