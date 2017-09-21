package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/magoofile/magoo/build"
	"github.com/magoofile/magoo/parse"
	"github.com/pkg/errors"
)

func main() {
	log.SetFlags(0)
	ctx := build.Default
	ctx.RequiredTags = []string{"magoo"}
	ctx.BuildTags = []string{"magoo"}
	p, err := ctx.ImportDir(".", 0)
	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			log.Fatal("No files marked with the magoo build tag in this directory.")
		}
		log.Fatal(err)
	}
	fns, err := parse.Package(".", p.GoFiles)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v", fns)
}

func hashStrings(list []string) (string, error) {
	var hashes []string
	for _, s := range list {
		h, err := hash(s)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}
	sort.Strings(hashes)
	h := sha1.Sum([]byte(strings.Join(hashes, "")))
	return fmt.Sprintf("%x", h), nil
}

func hash(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", errors.WithMessage(err, "can't open input file")
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errors.WithMessage(err, "can't write data to hash")
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CacheEnv is the environment variable that users may set to change the
// location where magoo stores its script files.
const CacheEnv = "MAGOO_CACHE"

// confDir returns the default gorram configuration directory.
func confDir(env map[string]string) string {
	d := env[CacheEnv]
	if d != "" {
		return d
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), "magoo")
	default:
		return filepath.Join(os.Getenv("HOME"), ".magoo")
	}
}
