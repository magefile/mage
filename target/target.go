package target

import (
	"os"
	"path/filepath"
	"time"
)

// Path compares the ModTime for src with targets and returns true if any
// target received most recent changes than src.
func Path(src string, targets ...string) (bool, error) {
	stat, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	srcTime := stat.ModTime()
	dt, err := loadTargets(targets)
	if err != nil {
		return false, err
	}
	t := dt.modTime()
	if t.After(srcTime) {
		return true, nil
	}
	return false, nil
}

// Dir reports whether any of the sources have been modified
// more recently than the destination.  If a source or destination is
// a directory, modtimes of files under those directories are compared
// instead.
func Dir(dst string, sources ...string) (bool, error) {
	stat, err := os.Stat(dst)
	if err != nil {
		return false, err
	}
	srcTime := stat.ModTime()
	if stat.IsDir() {
		srcTime = calDirModTimeRecursive(stat)
	}
	dt, err := loadTargets(sources)
	if err != nil {
		return false, err
	}
	t := dt.modTimeDir()
	if t.After(srcTime) {
		return true, nil
	}
	return false, nil
}

func calDirModTimeRecursive(dir os.FileInfo) time.Time {
	t := dir.ModTime()
	filepath.Walk(dir.Name(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.ModTime().After(t) {
			t = info.ModTime()
		}
		return nil
	})
	return t
}

type depTargets struct {
	src    []os.FileInfo
	hasdir bool
	latest time.Time
}

func loadTargets(targets []string) (*depTargets, error) {
	d := &depTargets{}
	for _, v := range targets {
		stat, err := os.Stat(v)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			d.hasdir = true
		}
		d.src = append(d.src, stat)
		if stat.ModTime().After(d.latest) {
			d.latest = stat.ModTime()
		}
	}
	return d, nil
}

func (d *depTargets) modTime() time.Time {
	return d.latest
}

func (d *depTargets) modTimeDir() time.Time {
	if !d.hasdir {
		return d.latest
	}
	for _, i := range d.src {
		t := i.ModTime()
		if i.IsDir() {
			t = calDirModTimeRecursive(i)
		}
		if t.After(d.latest) {
			d.latest = t
		}
	}
	return d.latest
}
