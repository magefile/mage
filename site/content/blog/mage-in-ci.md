+++
title = "Mage in CI"
date = 2018-09-06T21:41:23-04:00
type = "post"
author = "Nate Finch"
authorLink = "twitter.com/natethefinch"
+++

So, there's a bootstrap problem with Mage.  What if you want to use Mage for
your build, but you need to get it during your CI build?  Mage is best built...
with Mage.  The problem is that if you don't use the custom `mage install` then
Mage isn't compiled with ldflags that set all the usseful info in `mage
-version`.  So what do you do?

Luckily, Mage has a [zero install](/zeroInstall) option.  You don't need Mage to
build Mage, there's a bootstrap file that will let you use `go run` to install
Mage with all the great version info.  The bootstrap.go file in the root of the
repo hooks into mage's libraries and acts like Mage itself, so you can pass it
mage targets in the current directory.

This will download the mage source and install it (if you're using a gopath):

```plain
go get -d github.com/magefile/mage
go run $GOPATH/src/github.com/magefile/mage/bootstrap.go install
```

If you're using go modules, you can do it with good old git clone:

```plain
git clone git@github.com:magefile/mage
cd mage
go run bootstrap.go install
```

Note that in the second case, the binary will be copied to where go env GOPATH
points (for now, PRs welcome).