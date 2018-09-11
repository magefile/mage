+++
title = "On Release Versions"
date = 2018-09-07T09:39:13-04:00
type = "post"
author = "Nate Finch"
authorLink = "twitter.com/natethefinch"
+++

If you've been using Mage for a while (thank you!) then you may have noticed
something odd in the last week.  Mage was previously at v2.3.0... but now the
most recent release is 1.3.0 (released today!).  So, yeah, sorry if this makes
your life difficult, and I know it'll cause problems for some people.  Hopefully
most people have mage vendored and can go about their business until they need
something new, and then it shouldn't be a big deal to just take whatever the
newest release is, even if it "looks" like it's an earlier release.

The reason for this is two things: the major motivating factor is go modules
support.  To enable support for go modules, if you are on version 2 or higher of
your code, then import paths need to be appended with /v2.  i.e. to import
github.com/magefile/mage/mg from a v2 tag, you'd write the import as `import
"github.com/magefile/mage/v2"`.  With a go.mod that declares the module to be
`github.com/magefile/mage/v2`, the go 1.11 tool understands that the /v2 on the
end is not a real subdirectory, but just references the version of the module,
and it'll strip that off.  This understanding was backported to go 1.10.3 and go
1.9.7.  But any other versions of the go tool earlier than 1.11 will expect
there to actually be a v2 directory in your repo for this code.

Since I really want to keep Mage as backwards compatible as possible (CI runs
tests with versions as far back as 1.7)... I didn't want this headache.  I also
didn't want everyone to have to make their import statements say /v2. Mage's
libraries are and always will be backwards compatible.  The tags on the repo
previosuly were not intended to be semantic versions.  In fact, I had intended
to simply make each new release a major version specifically to make it clear
that these aren't semantic versions (just like Chrome is on v70 right now).
However, I wasn't very careful, and some of my tags were in semantic version
format (i.e. v2.2.0) and thus the go 1.11 tool took those to mean that the repo
is on v2.  

What to do?  Mages users right now are still somewhat limited... looking at
github there's maybe a hundred or so repos using Mage (again, thank you!).. but
I hope in the future that this will grow manyfold.  In order to make the many
years and hopefully many orders of magnitude of users' lives easier, I decided
it was best to rip out the v2.x tags and reclaim the title of v1.  This keeps
import paths the same as they were in v1.10 and before. 

From now on, all releases will be in semantic versions of ever increasing v1.x.x.  

Although I know this will cause some issues, I hope that the pain will be brief
and slight.... and the ability to stay backwards compatible with earlier
versions of Go and to make future development easier is worth it in my opinion.
