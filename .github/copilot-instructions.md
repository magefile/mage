# Agent Instructions

## Go Compatibility

This project targets Go 1.18. Do not use language features or standard library
functions/types introduced after Go 1.18.

## Go Dependencies

Do not add any external module dependencies. Only import packages from the Go
standard library or from within this module.

## Go Formatting

After modifying any Go files, run `goimports -w` on the changed files.
