//go:build !go1.18
// +build !go1.18

package mage

// set by ldflags when you "mage build"
var (
	commitHash = "<not set>"
	timestamp  = "<not set>"
	gitTag     = ""
)

func getTag() string {
	return gitTag
}
