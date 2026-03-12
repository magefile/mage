// Package targets defines the build targets for the mage project.
package targets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Install runs "go install" for mage. This generates the version info the binary.
func Install() error {
	name := "mage"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	gocmd := mg.GoCmd()
	// use GOBIN if set in the environment, otherwise fall back to first path
	// in GOPATH environment string
	bin, err := sh.Output(gocmd, "env", "GOBIN")
	if err != nil {
		return fmt.Errorf("can't determine GOBIN: %w", err)
	}
	if bin == "" {
		gopath, err := sh.Output(gocmd, "env", "GOPATH")
		if err != nil {
			return fmt.Errorf("can't determine GOPATH: %w", err)
		}
		paths := strings.Split(gopath, string([]rune{os.PathListSeparator}))
		bin = filepath.Join(paths[0], "bin")
	}
	// specifically don't mkdirall, if you have an invalid gopath in the first
	// place, that's not on us to fix.
	if err := os.Mkdir(bin, 0o700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create %q: %w", bin, err)
	}
	path := filepath.Join(bin, name)

	// we use go build here because if someone built with go get, then `go
	// install` turns into a no-op, and `go install -a` fails on people's
	// machines that have go installed in a non-writeable directory (such as
	// normal OS installs in /usr/bin)
	return sh.RunV(gocmd, "build", "-o", path, "-ldflags="+flags(), "github.com/magefile/mage")
}

var releaseTag = regexp.MustCompile(`^v1\.\d+\.\d+$`)

// Release generates a new release. Expects a version tag in v1.x.x format.
func Release(tag string) (err error) {
	mg.Deps(Tools)

	if !releaseTag.MatchString(tag) {
		return errors.New("TAG environment variable must be in semver v1.x.x format, but was " + tag)
	}

	if err := sh.RunV("git", "tag", "-a", tag, "-m", tag); err != nil {
		return err
	}
	if err := sh.RunV("git", "push", "origin", tag); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = sh.RunV("git", "tag", "--delete", tag)
			_ = sh.RunV("git", "push", "--delete", "origin", tag)
		}
	}()
	return sh.RunV("goreleaser", "release")
}

// Clean removes the temporarily generated files from Release.
func Clean() error {
	return sh.Rm("dist")
}

func flags() string {
	timestamp := time.Now().Format(time.RFC3339)
	hash := hash()
	tag := tag()
	if tag == "" {
		tag = "dev"
	}
	return fmt.Sprintf(`-X "github.com/magefile/mage/mage.timestamp=%s" -X "github.com/magefile/mage/mage.commitHash=%s" -X "github.com/magefile/mage/mage.gitTag=%s"`, timestamp, hash, tag)
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	s, _ := sh.Output("git", "describe", "--tags")
	return s
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}

var goTools = []string{
	"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.2",
	"github.com/goreleaser/goreleaser/v2@v2.14.3",
}

// Tools installs the dev tools used by mage, such as golangci-lint.
func Tools() error {
	for _, tool := range goTools {
		if err := sh.Run("go", "install", tool); err != nil {
			return fmt.Errorf("failed to install %s: %w", tool, err)
		}
	}
	return nil
}

// Lint runs golangci-lint on the codebase.
func Lint() error {
	mg.Deps(Tools)
	return sh.RunV("golangci-lint", "run")
}
