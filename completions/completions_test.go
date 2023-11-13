package completions

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestGetCompletions(t *testing.T) {
	testCases := []struct {
		name     string
		shell    string
		expected Completion
		err      bool
	}{
		{
			name:     "zsh",
			shell:    "zsh",
			expected: &Zsh{},
			err:      false,
		},
		{
			name:     "nonexistent",
			shell:    "nonexistent",
			expected: nil,
			err:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completion, err := GetCompletions(tc.shell)
			if err != nil && !tc.err {
				t.Fatalf("expected to get completions, but got error: %v", err)
			}
			if completion != tc.expected {
				t.Fatalf("expected to get completions, but got: %v", completion)
			}
		})
	}
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

func TestGenerateCompletionsFails(t *testing.T) {
	cmpl := &Zsh{}
	writer := &failingWriter{}

	err := cmpl.GenerateCompletions(writer)
	if err == nil {
		t.Fatalf("expected failure to generate completions, but got no error")
	}
}

func TestGenerateCompletions(t *testing.T) {
	zshTestData, err := ioutil.ReadFile("mage.zsh")
	if err != nil {
		t.Fatalf("expected to get test data, but got error: %v", err)
	}

	testCases := []struct {
		name     string
		cmpl     Completion
		expected string
		err      bool
	}{
		{
			name:     "zsh",
			cmpl:     &Zsh{},
			expected: string(zshTestData),
			err:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tc.cmpl.GenerateCompletions(&buf)
			if err != nil {
				t.Fatalf("expected to generate completions, but got error: %v", err)
			}
			actual := buf.String()
			if actual != tc.expected {
				t.Fatalf("expected: %q\ngot: %q", tc.expected, actual)
			}
		})
	}
}
