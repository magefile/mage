package target

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var Path = File

// Path reports if any of the sources have been modified more recently
// than the destination.  Path does not descend into directories, it literally
// just checks the modtime of each thing you pass to it.
func File(target string, sources ...string) (bool, error) {
	sourcesModTime, err := checkDependencies(sources)
	if err != nil {
		return false, fmt.Errorf("can't build target %s: %v", target, err)
	}

	targetInfo, err := os.Stat(target)

	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("can't build target %s: %v", target, err)
	}
	if err == nil && targetInfo.IsDir() {
		return false, fmt.Errorf("target %s is a directory", target)
	}

	return os.IsNotExist(err) || sourcesModTime.After(targetInfo.ModTime()), nil
}

// Dir reports whether any of the sources have been modified
// more recently than the destination.  If a source or destination is
// a directory, modtimes of files under those directories are compared
// instead.
func Dir(target string, sources ...string) (bool, error) {
	sourcesModTime, err := checkDependencies(sources)
	if err != nil {
		return false, fmt.Errorf("can't build target %s: %v", target, err)
	}

	targetInfo, err := os.Stat(target)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("can't build target %s: %v", target, err)
	}
	if !targetInfo.IsDir() {
		return false, fmt.Errorf("target %s is not a directory", target)
	}

	if os.IsNotExist(err) {
		return true, nil
	}

	targetModTime, err := recursiveModTime(target, targetInfo)
	if err != nil {
		return false, fmt.Errorf("can't build target %s: %v", target, err)
	}
	return sourcesModTime.After(targetModTime), nil
}

func checkDependencies(sources []string) (time.Time, error) {
	var (
		mostRecentModTime time.Time
	)
	for _, source := range sources {
		fileInfo, err := os.Stat(source)
		if os.IsNotExist(err) {
			return time.Time{}, fmt.Errorf("dependency %s does not exist", source)
		}

		var sourceModTime time.Time
		if fileInfo.IsDir() {
			if sourceModTime, err = recursiveModTime(source, fileInfo); err != nil {
				return time.Time{}, fmt.Errorf("an error occurred getting the modification time for %s", source)
			}
		} else {
			sourceModTime = fileInfo.ModTime()
		}

		if sourceModTime.After(mostRecentModTime) {
			mostRecentModTime = sourceModTime
		}
	}
	return mostRecentModTime, nil
}

func recursiveModTime(rootPath string, rootInfo os.FileInfo) (time.Time, error) {
	mostRecentModTime := rootInfo.ModTime()
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.ModTime().After(mostRecentModTime) {
			mostRecentModTime = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return mostRecentModTime, nil
}
