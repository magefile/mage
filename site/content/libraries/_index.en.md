+++
title = "Helper Libraries"
weight = 45
+++

There are three helper libraries bundled with mage,
[mg](https://godoc.org/github.com/magefile/mage/mg),
[sh](https://godoc.org/github.com/magefile/mage/sh), and 
[target](https://godoc.org/github.com/magefile/mage/target)  

Package `mg` contains mage-specific helpers, such as Deps for declaring
dependent functions, and functions for returning errors with specific error
codes that mage understands.

Package `sh` contains helpers for running shell-like commands with an API that's
easier on the eyes and more helpful than os/exec, including things like
understanding how to expand environment variables in command args.

Package `target` contains helpers for performing make-like timestamp comparing
of files.  It makes it easy to bail early if this target doesn't need to be run.
