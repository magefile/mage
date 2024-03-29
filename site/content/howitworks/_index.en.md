+++
title="How It Works"
weight = 50
+++

Mage scans the current directory for go files with the `mage` build tag (i.e.
`//go:build mage`), using the normal go build rules for following build
constraints (aside from requiring the mage tag).  It then parses those files to
find the build targets, generates a main file for the command, and compiles a
binary from those files.  The magefiles are hashed so that if they remain
unchanged, the same compiled binary will be reused next time, to avoid the
generation overhead.  As of Mage 1.3.0, the version of Go used to compile the
binary is also used in the hash.

## Binary Cache

Compiled magefile binaries are stored by default in $HOME/.magefile or in
%USERPROFILE%\magefile.  This location can be customized by setting the
MAGEFILE_CACHE environment variable.

## Go Environment

Mage itself requires no dependencies to run. However, because it is compiling go
code, you must have a valid go environment set up on your machine.  Mage is
compatible with any go 1.7+ environment (earlier versions may work but are not
tested).
