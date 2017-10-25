<h1 align=center>mage</h1>
<p align="center"><img src="https://user-images.githubusercontent.com/3185864/31061203-6f6743dc-a6ec-11e7-9469-b8d667d9bc3f.png"/></p>

## About [![Build Status](https://travis-ci.org/magefile/mage.svg?branch=master)](https://travis-ci.org/magefile/mage)

Mage is a make/rake-like build tool using Go.  You write plain-old go functions,
and mage automatically uses them as Makefile-like runnable targets.


## Installation

Install by running `go run bootstrap.go`, which will build the mage binary with
compile-time data on the version number and commit hash.  If you have installed
using `go get`, your binary will work just fine, but will not have the version
number or commit hash. To fix this, just run `mage build` in the root of the
mage repo (or run `go run bootstrap.go`).

The mage binary will be created in your $GOPATH/bin directory.

You may also install a binary release from our
[releases](https://github.com/magefile/mage/releases) page. 

## Demo

[![Mage Demo](https://img.youtube.com/vi/GOqbD0lF-iA/maxresdefault.jpg)](https://www.youtube.com/watch?v=GOqbD0lF-iA)

## Discussion

Join the `#mage` channel on [gophers slack](https://gophers.slack.com/messages/general/) for discussion of usage, development, etc.

# Documentation

see [magefile.org](https://magefile.org) for full docs

see [godoc.org/github.com/magefile/mage/mage](https://godoc.org/github.com/magefile/mage/mage) for how to use mage as a library.

# Why?

Makefiles are hard to read and hard to write.  Mostly because makefiles are essentially fancy bash scripts with significant white space and additional make-related syntax.

Mage lets you have multiple magefiles, name your magefiles whatever you
want, and they're easy to customize for multiple operating systems.  Mage has no
dependencies (aside from go) and runs just fine on all major operating systems, whereas make generally uses bash which is not well supported on Windows.
Go is superior to bash for any non-trivial task involving branching, looping, anything that's not just straight line execution of commands.  And if your project is written in Go, why introduce another
language as idiosyncratic as bash?  Why not use the language your contributors
are already comfortable with?

# TODO

* File conversion tasks
