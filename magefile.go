//+build mage

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/magefile/mage/git"
	"github.com/magefile/mage/sh"
)

// Runs go install for mage.  This generates the the version
// info the binary.
func Build() error {
	ldf, err := flags()
	if err != nil {
		return err
	}

	log.Print("running go install")
	return sh.Run("go", "install", "--ldflags="+ldf, "github.com/magefile/mage")
}

func flags() (string, error) {
	timestamp := time.Now().Format(time.RFC3339)
	hash := git.Hash()
	version := git.Tag()
	if version == "" {
		version = "dev"
	}
	return fmt.Sprintf(`-X "github.com/magefile/mage/mage.timestamp=%s" -X "github.com/magefile/mage/mage.commitHash=%s" -X "github.com/magefile/mage/mage.version=%s"`, timestamp, hash, version), nil
}
