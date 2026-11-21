package tasks

import (
	"fmt"
	"os"
	"path/filepath"
)

func WorkingDir() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(filepath.Base(dir))
	return nil
}
