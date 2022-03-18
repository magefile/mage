+++
title = "Release v1.13.0 - Magefiles Directory"
date = 2022-03-18T08:17:52-05:00
type = "post"
author = "Nate Finch"
authorLink = "https://twitter.com/natethefinch"
+++

[Mage v1.13.0](https://github.com/magefile/mage/releases/tag/v1.13.0) is now
released with a few fixes and one big new feature - the magefiles subdirectory.

You can now put your magefiles in a subdirectory of your repo called
`/magefiles`, and Mage will find them automatically, like it always has for
files in the root. Files in `/magefiles` do not need the `mage` build tag to be
discovered by Mage (but if it’s there, it’ll still find them).

This is a big change that will be a significant quality of life improvement for
many developers. Until now, Mage was designed to allow you to keep magefiles in
the root of your repo, tagged with the `mage` build tag, so that your normal go
build tooling would ignore them. However, this had the side effect of meaning
that linting etc would *also* ignore them. You could add the `mage` build tag to
your IDE and/or gopls, but because Mage also requires the files to be in
`package main`, this could result in having files from two different packages in
the root of your repo, which would cause linters and gopls to complain.

Now, you can put those files in a separate directory, and you don’t even need
the build tag any more. No conflicting packages, no need to twiddle with build
tags, and all this is automatic, no settings need to be adjusted, you just need
Mage v1.13+.

Migrating your magefiles from the root of your repo to a subdirectory is
trivial. Just copy them into `/magefiles`.  That’s it. Mage will continue to
work like it always has. If you like, you can remove the build tags, mage
doesn’t care if they’re there or not.

For backwards compatibility reasons, if you have magefiles in the root of your
repo (defined as .go files with the `mage` build tag), and you have a `/magefiles`
subdirectory, Mage will default to the old behavior of only looking at the files
in the root directory. This is to avoid breaking anyone who happened to already
have a `/magefiles` directory. In this instance, Mage v1.13 will print out a
warning when you run it, noting that it’s using the files in the root of the
repo. 

Because this behavior requires some extra time and work even if you *don’t* have
magefiles in the root directory, **in the future** we plan to change the default
so that Mage looks for the `/magefiles` subdirectory first, and ignores any
files in the root unless there is no `/magefiles` directory. We have no plans to
drop support for using files in the root directory, but if you want to continue
using that behavior, you'll eventually need to rename the `/magefiles`
subdirectory (but again, that's planned for the future, not right now). 

Note that as of 1.13, Mage still requires those files to be in `package main`.
We plan to remove that requirement in the future, so that linters won’t complain
about the missing `func main`. 

We hope this helps fix most of the problems some people have had using Mage with
standard Go tooling. Please feel free to post any questions you have in the
`#mage` channel of [gophers slack](https://gophers.slack.com/messages/mage/), or
on the [discussions board](https://github.com/magefile/mage/discussions) on
Mage’s [github repo](https://github.com/magefile/mage). Happy building!