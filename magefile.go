//+build mage

package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/magefile/mage/sh"
)

// Runs go install for mage.  This generates the the version
// info the binary.
func Build() error {
	ldf, err := flags()
	if err != nil {
		return err
	}

	return sh.Run("go", "install", "-ldflags="+ldf, "github.com/magefile/mage")
}

// Generates binaries for all supported platforms.  Currently that means a
// combination of windows, linux, and OSX in 32 bit and 64 bit formats.  The
// files will be dumped in the local directory with names according to their
// supported platform.
func BuildAll() error {
	ldf, err := flags()
	if err != nil {
		return err
	}
	var ext string
	for _, OS := range []string{"windows", "darwin", "linux"} {
		if OS == "windows" {
			ext = ".exe"
		} else {
			ext = ""
		}
		for _, ARCH := range []string{"amd64", "386"} {
			log.Printf("running go build for GOOS=%s GOARCH=%s", OS, ARCH)
			env := map[string]string{"GOOS": OS, "GOARCH": ARCH}
			if err := sh.RunWith(env, "go", "build", "-o", "mage_"+OS+"_"+ARCH+ext, "--ldflags="+ldf); err != nil {
				return err
			}
		}
	}
	return err
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
