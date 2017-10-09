package mg

import (
	"os"
	"path/filepath"
	"time"
)

// Target checks whether the target files/directories have been more recently
// modified than src.
//
// When recursive is set to true, if src is a directory it will be walked
// recursively to deermine modified time.
func Target(src string, recursive bool, targets ...string) (bool, error) {
	if len(targets) == 0 {
		return true, nil
	}
	stat, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	srcTime := stat.ModTime()
	if stat.IsDir() && recursive {
		srcTime = calDirModTimeRecursive(stat)
	}
	dt, err := loadTargets(targets)
	if err != nil {
		return false, err
	}
	t := dt.modTime(false)
	if t.After(srcTime) {
		return true, nil
	}
	if dt.hasdir {
		return dt.modTime(true).After(srcTime), nil
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

func (d *depTargets) modTime(recursive bool) time.Time {
	if !recursive || !d.hasdir {
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
