package mage

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVariadicTargetDiscovery(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/variadic",
		Stderr: stderr,
		Stdout: stdout,
		List:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected code 0, got %d; stderr: %s", code, stderr)
	}
	want := `Targets:
  collect           prints its fixed prefix and all remaining arguments.
  fixed             prints one fixed argument.
  shared:collect    prints imported variadic arguments.
  tools:collect     prints namespace arguments.
  types             prints converted fixed arguments followed by variadic arguments.
  variadic*         prints all remaining arguments.

* default target
`
	if got := stdout.String(); got != want {
		t.Fatalf("expected output %q, got %q", want, got)
	}
}

func TestUnsupportedVariadicTargetsStayHidden(t *testing.T) {
	for _, target := range []string{"optionalandvariadic", "variadicint"} {
		t.Run(target, func(t *testing.T) {
			stderr := &bytes.Buffer{}
			stdout := &bytes.Buffer{}
			inv := Invocation{
				Dir:    "./testdata/variadic",
				Stderr: stderr,
				Stdout: stdout,
				Help:   true,
				Args:   []string{target},
			}
			code := Invoke(inv)
			if code != 2 {
				t.Fatalf("expected code 2, got %d; stdout: %s; stderr: %s", code, stdout, stderr)
			}
			want := "Unknown target: \"" + target + "\"\n"
			if got := stderr.String(); got != want {
				t.Fatalf("expected error %q, got %q", want, got)
			}
		})
	}
}

func TestVariadicArgsThroughParseAndRun(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	code := ParseAndRun(stdout, stderr, nil, []string{
		"-d", "testdata/variadic",
		"fixed", "before",
		"collect", "prefix", "-flag", "--", "fixed", "after",
	})
	if code != 0 {
		t.Fatalf("expected code 0, got %d; stderr: %s", code, stderr)
	}
	want := "fixed:before\ncollect:prefix:[\"-flag\" \"--\" \"fixed\" \"after\"]\n"
	if got := stdout.String(); got != want {
		t.Fatalf("expected output %q, got %q", want, got)
	}
}

func TestVariadicHelpMatchesCompiledBinary(t *testing.T) {
	dir := "./testdata/variadic"
	name := filepath.Join(t.TempDir(), "mage_variadic_help_test")
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	stderr := &bytes.Buffer{}
	code := Invoke(Invocation{
		Dir:        dir,
		Stdout:     os.Stdout,
		Stderr:     stderr,
		CompileOut: name,
	})
	if code != 0 {
		t.Fatalf("compile failed with code %d: %s", code, stderr)
	}

	stdout := &bytes.Buffer{}
	stderr.Reset()
	code = Invoke(Invocation{
		Dir:    dir,
		Stdout: stdout,
		Stderr: stderr,
		Help:   true,
		Args:   []string{"collect"},
	})
	if code != 0 {
		t.Fatalf("mage help failed with code %d: %s", code, stderr)
	}
	mageOutput := stdout.String()

	stdout.Reset()
	stderr.Reset()
	cmd := exec.CommandContext(context.Background(), name, "-h", "collect")
	cmd.Env = os.Environ()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("compiled binary help failed: %v; stderr: %s", err, stderr)
	}
	compiledOutput := stdout.String()

	binaryBase := filepath.Base(name)
	normalizedMage := strings.ReplaceAll(mageOutput, "\tmage ", "\tBINARY ")
	normalizedCompiled := strings.ReplaceAll(compiledOutput, "\t"+binaryBase+" ", "\tBINARY ")
	if normalizedMage != normalizedCompiled {
		t.Fatalf("help output mismatch:\nmage: %q\ncompiled: %q", mageOutput, compiledOutput)
	}
	if !strings.Contains(normalizedMage, "BINARY collect <prefix> [<args>...]") {
		t.Fatalf("variadic usage missing from help: %q", mageOutput)
	}
}

func TestVariadicTargetDuplicateDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name string
		dir  string
		want string
	}{
		{
			name: "alias",
			dir:  "./testdata/variadic_dupe_alias",
			want: `alias "BUILD" duplicates existing target(s): <current>.Build`,
		},
		{
			name: "imported target",
			dir:  "./testdata/variadic_dupe_import",
			want: `"build" target has multiple definitions`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stderr := &bytes.Buffer{}
			code := Invoke(Invocation{
				Dir:    tc.dir,
				Stdout: &bytes.Buffer{},
				Stderr: stderr,
				List:   true,
			})
			if code != 1 {
				t.Fatalf("expected code 1, got %d; stderr: %s", code, stderr)
			}
			if got := stderr.String(); !strings.Contains(got, tc.want) {
				t.Fatalf("expected error containing %q, got %q", tc.want, got)
			}
		})
	}
}
