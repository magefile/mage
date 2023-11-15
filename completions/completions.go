package completions

import (
	"fmt"
	"io"
)

type Completion interface {
	GenerateCompletions(w io.Writer) error
}

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
