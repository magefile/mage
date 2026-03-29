+++
title = "Mage"
+++

## About

Mage is a make/rake-like build tool using Go.  You write plain-old go functions,
and Mage automatically uses them as Makefile-like runnable targets.

## Installation

### From GitHub source (any OS)

Mage has no dependencies outside the Go standard library, and builds with Go 1.18
and above.

#### Using Go Install

```plain
go install github.com/magefile/mage@latest
mage -init
```
Instead of the @latest tag, you can specify the desired version, for example:

```plain
go install github.com/magefile/mage@v1.16.0
```

### From GitHub releases (any OS)

You may also install a binary release from our
[releases](https://github.com/magefile/mage/releases) page.

### With Homebrew (MacOS)

`brew install mage`

See [mage homebrew formula](https://formulae.brew.sh/formula/mage).

### With MacPorts (MacOS)

`sudo port install mage`

See [port page](https://ports.macports.org/port/mage/)

### With Scoop (Windows)

`scoop install mage`

See [scoop](https://scoop.sh/).

### Using asdf

The [asdf version manager](https://asdf-vm.com/) is a tool for installing release binaries from Github. With asdf installed, the [asdf plugin for mage](https://github.com/mathew-fleisch/asdf-mage) can be used to install any released version of mage.

```shell
asdf plugin add mage
asdf install mage latest
asdf global mage latest
```

## Example Magefile

```go
//go:build mage

package main

//mage:multiline // enable line return retention in doc output.

import (
    "github.com/magefile/mage/sh"
)

// Runs go mod download and then installs the binary.
func Deploy(env string,
    dryrun *bool, // if set to true, only builds the artifacts
) error {
    if err := sh.Run("go", "mod", "download"); err != nil {
        return err
    }
    return sh.Run("go", "install", "./...")
}
```


## Comments as Help Text

Comments on targets (functions) get converted into help text output. The first
sentence becomes the brief comment for the target on `mage -l`. The full comment
is printed out when you run `mage -h <target>`.  Comments on flag arguments are
printed as docs on the flags.

For example:

```go
//go:build mage

// These docs will become the main help text when you run `mage` or `mage -l`.
// This is where you talk about what the thing does.
package main

//mage:multiline // enable line return retention in doc output.

import (
    "github.com/any-go-packge/youwant"
)

// Deploy runs the build and then uploads the artifacts to the server.
// It deploys to the given environment.
func Deploy(ctx context.Context, env string,
    version *string, // git tag for the build, defaults to the next minor build if not set
    dryRun *bool,    // if set to true, just outputs the build artifacts
) error {
    return youwant.ToCallGoCode()
}
```

```plain
$ mage -l
These docs will become the main help text when you run `mage` or `mage -l`.

Targets:
  deploy  runs the build and then uploads the artifacts to the server.

$ mage -h deploy
Deploy runs the build and then uploads the artifacts to the server.
It deploys to the given environment.

Usage:
	mage deploy <env> [<flags>]

Flags:
    -version=<string>  git tag for the build, defaults to the next minor build if not set
    -dryrun=<bool>     if set to true, just outputs the build artifacts
```

## Magefiles directory

If you create your Magefile or files within a directory named `magefiles` And there is no Magefile in your current directory, 
`mage` will default to the directory as the source for your targets while keeping the current directory as working one.

The result is the equivalent of running `mage -d magefiles -w .`

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
  -h        show this help
  -init     create a starting template if no mage files exist
  -l        list mage targets in this directory
  -version  show version info for the mage binary

Options:
  -d <string> 
            directory to read magefiles from (default ".")
  -debug    turn on debug messages
  -f        force recreation of compiled magefile
  -goarch   sets the GOARCH for the binary created by -compile (default: current arch)
  -gocmd <string>
            use the given go binary to compile the output (default: "go")
  -goos     sets the GOOS for the binary created by -compile (default: current OS)
  -h        show description of a target
  -keep     keep intermediate mage files around after running
  -t <string>
            timeout in duration parsable format (e.g. 5m30s)
  -v        show verbose output when running mage targets
  -w <string>
            working directory where magefiles will run (default -d value)
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
