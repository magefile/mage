+++
title = "Compiling"
weight = 37
+++

## Mage ignores GOOS and GOARCH for its build

When building the binary for your magefile, mage will ignore the GOOS and GOARCH environment variables and use your current GOOS and GOARCH to ensure the binary that is built can run on your local system.  This way you can set GOOS and GOARCH when you run mage, to have it take effect on the outputs of your magefile, without it also rendering your magefile unrunnable on your local machine.

## Compiling a static binary

It can be useful to compile a static binary which has the mage execution runtime
and the tasks compiled in such that it can be run on another machine without
requiring any dependencies. To do so, pass the output path to the compile flag.
like this:

```plain
$ mage -compile ./static-output
```

The compiled binary uses flags just like the mage binary:

```plain
<cmd_name> [options] [target]

Commands:
  -l    list targets in this binary
  -h    show this help

Options:
  -h    show description of a target
  -t <string>
        timeout in duration parsable format (e.g. 5m30s)
  -v    show verbose output when running targets
```

## Compiling for a different OS -goos and -goarch

If you intend to run the binary on another machine with a different OS platform, you may use the `-goos` and `-goarch` flags to build the compiled binary for the target platform.  Valid values for these flags may be found here: https://golang.org/doc/install/source#environment.  The OS values are obvious (except darwin=MacOS), the GOARCH values most commonly needed will be "amd64" or "386" for 64 for 32 bit versions of common desktop OSes.

Note that if you run `-compile` with `-dir`, the `-compile` target will be *relative to the magefile dir*.