# Contributing

Of course, contributions are more than welcome. Please read these guidelines for
making the process as painless as possible.

## Discussion

Development discussion should take place on the #mage channel of [gopher
slack](https://gophers.slack.com/).

There is a separate #mage-dev channel that has the github app to post github
activity to the channel, to make it easy to follow.

## Issues

If there's an issue you'd like to work on, please comment on it, so we can
discuss approach, etc. and make sure no one else is currently working on that
issue.

Please always create an issue before sending a PR unless it's an obvious typo
or other trivial change.

## Dependency Management

Currently mage has no dependencies(!).  Let's try to keep it that way.  Since
it's likely that mage will be vendored into a project, adding dependencies to
mage adds dependencies to every project that uses mage.

## Versions

Please try to avoid using features of go and the stdlib that prevent mage from
being buildable with old versions of Go.  Definitely avoid anything that
requires go 1.9.

## Testing

Please write tests for any new features.  Tests must use the normal go testing
package.

Tests must pass the race detector (run `go test -race ./...`).

