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

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Install runs "go install" for mage. This generates the version info the binary.
func Install() error {
	fmt.Println("`mage install` is deprecated. Just use `go install` now.")
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
	return sh.RunV(gocmd, "build", "-o", path, "github.com/magefile/mage")
}

var releaseTag = regexp.MustCompile(`^v1\.\d+\.\d+$`)

// Release generates a new release. Expects a version tag in v1.x.x format.
// If dryRun is true, it creates a local tag and runs goreleaser without
// publishing, then deletes the tag. This can be used to verify release artifacts.
func Release(tag string, dryRun *bool) (err error) {
	if err := installTool("goreleaser"); err != nil {
		return err
	}

	if !releaseTag.MatchString(tag) {
		return errors.New("TAG environment variable must be in semver v1.x.x format, but was " + tag)
	}


	if dryRun != nil && *dryRun {
		if err = sh.RunV("git", "tag", "-a", tag, "-m", tag); err != nil {
			return err
		}
		defer func() { _ = sh.RunV("git", "tag", "--delete", tag) }()
		return sh.RunV("goreleaser", "release", "--skip=publish", "--skip=validate", "--clean")
	}

	if err = sh.RunV("git", "tag", "-a", tag, "-m", tag); err != nil {
		return err
	}
	if err = sh.RunV("git", "push", "origin", tag); err != nil { //nolint:gocritic // using = to assign named return for deferred cleanup
		return err
	}
	defer func() {
		if err != nil {
			_ = sh.RunV("git", "tag", "--delete", tag)
			_ = sh.RunV("git", "push", "--delete", "origin", tag)
		}
	}()
	return sh.RunV("goreleaser", "release", "--clean")
}

// Clean removes the temporarily generated files from Release.
func Clean() error {
	return sh.Rm("dist")
}

var goTools = map[string]string{
	"lint":       "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.2",
	"goreleaser": "github.com/goreleaser/goreleaser/v2@v2.14.3",
}

// Tools installs the dev tools used by mage, such as golangci-lint.
func Tools() error {
	for _, tool := range goTools {
		if err := installTool(tool); err != nil {
			return err
		}
	}
	return nil
}

func installTool(tool string) error {
	version, ok := goTools[tool]
	if !ok {
		return fmt.Errorf("unknown tool %q", tool)
	}
	if err := sh.Run("go", "install", version); err != nil {
		return fmt.Errorf("failed to install %s: %w", version, err)
	}
	return nil
}

// Lint runs golangci-lint on the codebase.
func Lint() error {
	err := installTool("lint")
	if err != nil {
		return err
	}

	return sh.RunV("golangci-lint", "run")
}
