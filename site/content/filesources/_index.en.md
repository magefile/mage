+++
title = "File Sources and Destinations"
weight = 35
+++

Mage supports make-like comparisons of file sources and file targets.  Using the
[target](https://godoc.org/github.com/magefile/mage/target) library, you can
easily compare the last modified times of a target file or directory with the
last modified time of the file or directories required to build that target.

`target.Path` compares the last modified time of a target file or directory with
the last modified time of one or more files or directories.  If any of the
sources are newer than the destination, the function will return true.  Note
that Path does not recurse into directories. If you give it a directory, the
only last modified time it'll check is that of the directory itself.

`target.Dir` is like `target.Path` except that it recursively checks files and
directories under any directories specified, comparing timestamps.