package git

import "github.com/magefile/mage/sh"

// Yes, we ignore errors here. Possible errors include: git not installed, not
// running in a repo, git not executable.  Note that the error text will be sent
// to stderr, so it should be obvious what's wrong.

// Tag returns the git tag for the current branch or "" if none.
func Tag() string {
	tag, _ := sh.Output("git", "describe", "--tags")
	return tag
}

// Hash returns the git hash for the current repo or "" if none.
func Hash() string {
	hash, _ := sh.Output("git", "rev-parse", "HEAD")
	return hash
}

// Branch returns the git branch name for the current repo or "" if none.
func Branch() string {
	branch, _ := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	return branch
}
