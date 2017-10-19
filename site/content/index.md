+++
title = "Mage"
+++

<p align="center"><img width=300 src="/images/gary.svg"/></p>

<p align="center">Mage is a make/rake-like build tool using Go.</p>

## Installation

Mage has no dependencies outside the Go standard library, and builds with Go 1.7
and above (possibly even lower versions, but they're not regularly tested). To
install, just use `go get`:

`go get github.com/magefile/mage`

## Demo

{{< youtube GOqbD0lF-iA >}}

## Discussion

Join the `#mage` channel on [gophers slack](https://gophers.slack.com/messages/general/) for discussion of usage, development, etc.


## Plugins

There are no plugins.  You don't need plugins.  It's just Go code.  You can
import whatever libraries you want.  Every library in the go ecosystem is a mage
plugin.  Every tool you use with Go can be used with Magefiles.

## Usage
```
mage [options] [target]
Options:
  -f    force recreation of compiled magefile
  -h    show this help
  -init
        create a starting template if no mage files exist
  -keep
        keep intermediate mage files around after running
  -l    list mage targets in this directory
  -t string
    	  timeout in duration parsable format (e.g. 5m30s)
  -v    show verbose output when running mage targets
  -version
        show version info for the mage binary
```

## Environment Variables

You may set MAGE_VERBOSE=1 to always enable verbose logging in your magefiles,
without having to remember to pass -v every time.

## Why?

Makefiles are hard to read and hard to write.  Mostly because makefiles are essentially fancy bash scripts with significant white space and additional make-related syntax.

Mage lets you have multiple magefiles, name your magefiles whatever you
want, and they're easy to customize for multiple operating systems.  Mage has no
dependencies (aside from go) and runs just fine on all major operating systems, whereas make generally uses bash which is not well supported on Windows.
Go is superior to bash for any non-trivial task involving branching, looping, anything that's not just straight line execution of commands.  And if your project is written in Go, why introduce another
language as idiosyncratic as bash?  Why not use the language your contributors
are already comfortable with?

## Code

[https://github.com/magefile/mage](https://github.com/magefile/mage)

## Projects that build with Mage

[![Hugo](/images/hugo.png)](https://github.com/gohugoio/hugo) [![Gnorm](/images/gnorm.png)](https://github.com/gnormal/gnorm)
