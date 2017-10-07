+++
title = "Helper Libraries"
weight = 45
+++

There are two libraries bundled with mage,
[mg](https://godoc.org/github.com/magefile/mage/mg) and
[sh](https://godoc.org/github.com/magefile/mage/sh).  

Package `mg` contains
mage-specific helpers, such as Deps for declaring dependent functions, and
functions for returning errors with specific error codes that mage understands.

Package `sh` contains helpers for running shell-like commands with an API that's
easier on the eyes and more helpful than os/exec, including things like
understanding how to expand environment variables in command args.

