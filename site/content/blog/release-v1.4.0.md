+++
title = "Release v1.4.0 - Go Modules"
date = 2018-09-11T09:17:52-04:00
type = "post"
author = "Nate Finch"
authorLink = "twitter.com/natethefinch"
+++

[Mage v1.4.0](https://github.com/magefile/mage/releases/tag/v1.4.0) is released
with proper go modules support.

I finally figured out the wacky problem that Mage was having with Go Modules.
The main problem seems to be that it's the working directory where `go` is run
that determines whether or not the go tool decides to use modules or gopath.
This is backwards of what I would expect, which would be that the location of
the code determines whether to use modules or not (i.e. if it's outside a
gopath, use modules).  

I filed [a bug about this](https://github.com/golang/go/issues/27612) on the go
project.  It's possible this is intended, but confusing behavior.

However, the good news is that now that I figured out the problem, I can fix it.
I of course also wrote a test around it.  It's easier to reproduce in the clean
environment of CI than in my own messy machine, which is a nice reversal of how
most of these bugs go.

I don't know if this is the last of the go module bugs, but for now it seems to
be a pretty important milestone passed, with tests to keep it working in the
future.

This version also changes the implementation of how we detect magefiles and how
we parse them, greatly simplifying the code.  Since Mage wants to filter files
to only those that include the `mage` build tag, I had originally hacked the
go/build package to include a way to make tags required. Normally if you build
with additional tags, the build includes files with those tags *and* those with
no tags... I wanted to exclude those with no tags.  However, it occurred to me
that this could be done much more simply by just running `go list` twice - once
with no tags and once with the `mage` tag, and then see what additional files
were added with the second run.  By definition, those files are ones that have
the mage tag.

I also simplified how we're processing types.  Go 1.11's modules change how type
parsing works, and the old go/types package isn't up to the challenge. There's a
new package at [golang.org/x/tools/go/packages](golang.org/x/tools/go/packages)
that is supposed to replace go/types for this.  I started looking into using
go/packages, but then realized my situation is simple enough that I can actually
just use go/ast and not even need to dive into types very much (the only two
types Mage knows about right now is `error` and `context.Context`).

The end result was a PR that removed 11,000 lines of code, most of which was 3
copies of standard library packages that no longer needed to be used & included
- two hacked copies of go/build (one for go1.10 and lower, and one for go1.11),
as well as a copy of the go/build source importer, which wasn't in the standard
library in go1.8 and below.

Please take a look and definitely file an issue if you have any problems with
module support or the new parsing code.

