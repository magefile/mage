//+build mage

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/magefile/mage/sh"
)

// Runs "go install" for mage.  This generates the version info the binary.
func Build() error {
	ldf, err := flags()
	if err != nil {
		return err
	}

	return sh.Run("go", "install", "-ldflags="+ldf, "github.com/magefile/mage")
}

// Generates a new release.  Expects the TAG environment variable to be set,
// which will create a new tag with that name.
func Release() (err error) {
	if os.Getenv("TAG") == "" {
		return errors.New("MSG and TAG environment variables are required")
	}
	if err := sh.RunV("git", "tag", "-a", "$TAG"); err != nil {
		return err
	}
	if err := sh.RunV("git", "push", "origin", "$TAG"); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			sh.RunV("git", "tag", "--delete", "$TAG")
			sh.RunV("git", "push", "--delete", "origin", "$TAG")
		}
	}()
	return sh.RunV("goreleaser")
}

// Remove the temporarily generated files from Release.
func Clean() error {
	return sh.Rm("dist")
}

func flags() (string, error) {
	timestamp := time.Now().Format(time.RFC3339)
	hash := hash()
	tag := tag()
	if tag == "" {
		tag = "dev"
	}
	return fmt.Sprintf(`-X "github.com/magefile/mage/mage.timestamp=%s" -X "github.com/magefile/mage/mage.commitHash=%s" -X "github.com/magefile/mage/mage.gitTag=%s"`, timestamp, hash, tag), nil
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	buf := &bytes.Buffer{}
	_, _ = sh.Exec(nil, buf, nil, "git", "describe", "--tags")
	return buf.String()
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}
