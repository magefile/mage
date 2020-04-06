+++
title = "Importing Targets"
weight = 11
+++

Mage allows you to import targets from another package into the current
magefile.  This is very useful when you have common targets that you'd like to
be able to reuse across many repos.

## Imported Package Restrictions

To import targets from another package, that package must *not* be a main
package.  i.e. it cannot be `package main`.  This is in contrast to a normal
mage package that must be `package main`.  The reason is that the go tool won't
let you import a main package.

In addition, all package files will be imported, so long as they don't have a 
build tag.  If you try to import a package consisting only of files with build 
tags (e.g. `//+build mage`), it will cause an error since mage doesn't set any 
build tags when importing packages.  Any exported function, in imported 
packages, that matches Mage's allowed formats will be picked up as a target.

Aliases and defaults in imported packages will be ignored. 

Other than these differences, you can write targets in those packages just
like a normal magefile.

## Two Ways to Import

Importing targets from a package simply requires adding a `// mage:import`
comment on an import statement in your magefile.  If there is a name after this
tag, the targets will be imported into what is effectively like a namespace.

```go
import (
    // mage:import
    _ "example.com/me/foobar" 
    // mage:import build
    "example.com/me/builder"
)
```

The first mage import above will add the targets from the foobar package to the
root namespace of your current magefile.  Thus, if there's a `func Deploy()` in
foobar, your magefile will now have a `deploy` target.

The second import above will create a `build` namespace for the targets in the
`builder` package.  Thus, if there's a `func All()` in builder, your magefile
will now have a `build:all` target.

Note that you can use imported package targets as dependencies with `mg.Deps`,
just like any other function works in `mg.Deps`, so you could have
`mg.Deps(builder.All)`.

If you don't need to actually use the package in your root magefile, simply make
the import an underscore import like the first import above.


