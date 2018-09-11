+++
title = "Release v1.3.0 - GoCmd"
date = 2018-09-07T21:19:49-04:00
type = "post"
author = "Nate Finch"
authorLink = "twitter.com/natethefinch"
aliases = ["/blog/2018/09/release-1.3.0/"]
+++

As Brad Fitzpatrick would say, [Mage
1.3.0](https://github.com/magefile/mage/releases/tag/v1.3.0) is the best release
of Mage ever! This comes quick on the heels of gophercon, so I have been
motivated to move some interesting new features through.  1.2.4 added real
support for go modules (though it needs some more polish).  This release adds a
feature that I think will become more useful as more projects use Mage.  

As you may or may not know (it hasn't been well advertised), it's now extremely
easy to run multiple go versions side by side.  You can `go get
golang.org/dl/go1.xx` for any version of go from 1.8 and up.  It'll drop a
binary on your system that you can then run `go1.xx download` and it'll download
and set up a new go environment for that verison of go.  From then on, you can
`go1.xx build` etc just like you do with the "go" binary.   

This makes it super easy to work on different projects with different required
versions of go support.  I used to work at Canonical, and we were restricted to
the versions of Go that shipped in the latest Ubuntu release, because our code
was going to shiop with that release, too, and it needed to buid with the
built-in go compiler.  This meant that we were almost always behind the latest
go version.  Of course, go is backwards compatible, so you can build 1.6
compatible code with go 1.11.... but you can also accidentally introduce
dependencies that don't exist and/or don't work in old versions.

Oh, we're supposed to be talking about Mage?  Yeah, so there's a new flag for
Mage in 1.3.0: `-gocmd` which lets you specify the binary to use to compile the
magefile binary.  Thus, if the code your magefile imports needs an older version
of Go, you can specify it there, e.g. `mage -gocmd go1.8.3 build`.  This
literally just calls `go1.8.3 build` to compile the magefile binary, rather than
`go build`. This means that you need that binary on your path (or specify the
full path), and it means that if you pass in something wacky (like grep or
something) that wacky things may happen.

One important change is that the version of go used to compile the magefile
binary is now part of the hash that determines whether we need to recompile the
binary.  So, if you run the exact same magefile with two different versions of
go, it'll create two different files in your magefile cache. (yes, there's a
binary cache, see [https://magefile.org/howitworks](https://magefile.org/howitworks/)).

Also in this version, we now print out the version of Go used to compile Mage
itself when you run `mage -version`.  This is important because mage compiled
with < 1.11 does not play well with go modules.  There's also some more
debugging output showing versions of go being used when you run with -debug...
again, this can help figure out when you're not running what you think you're
running.

Hope this is useful to some people.  Please feel free to drop by the repo and
[make an issue](https://github.com/magefile/mage/issues) if you have an idea of
something that could make Mage even better.
