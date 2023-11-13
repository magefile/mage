package completions

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type Completion interface {
	GenerateCompletions(w io.Writer) error
}

type Zsh struct{}

var completions = map[string]Completion{
	"zsh": &Zsh{},
}

func GetCompletions(shell string) (Completion, error) {
	completions := completions[shell]

	if completions == nil {
		return nil, fmt.Errorf("no completions for shell %q", shell)
	}

	return completions, nil
}

func (z *Zsh) GenerateCompletions(w io.Writer) error {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	filePath := filepath.Join(dir, "mage.zsh")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, string(data))
	if err != nil {
		return err
	}
	return nil
}
