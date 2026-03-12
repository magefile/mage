//go:build !windows

package mage

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSignals(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := "./testdata/signals"
	compileDir := t.TempDir()
	name := filepath.Join(compileDir, "mage_out")
	inv := Invocation{
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: name,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}

	run := func(stdout, stderr *bytes.Buffer, filename string, target string, signals ...syscall.Signal) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.CommandContext(context.Background(), filename, target)
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, target, err, stdout, stderr)
		}
		pid := cmd.Process.Pid
		go func() {
			time.Sleep(time.Millisecond * 500)
			for _, s := range signals {
				_ = syscall.Kill(pid, s)
				time.Sleep(time.Millisecond * 50)
			}
		}()
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, target, err, stdout, stderr)
		}
		return nil
	}

	if err := run(stdout, stderr, name, "exitsAfterSighup", syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	want := "received sighup\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, "exitsAfterSigint", syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "exiting...done\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}
	got = stderr.String()
	want = "cancelling mage targets, waiting up to 5 seconds for cleanup...\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, "exitsAfterCancel", syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "exiting...done\ndeferred cleanup\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}
	got = stderr.String()
	want = "cancelling mage targets, waiting up to 5 seconds for cleanup...\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, "ignoresSignals", syscall.SIGINT, syscall.SIGINT); err == nil {
		t.Fatal("expected an error because of force kill")
	}
	got = stderr.String()
	want = "cancelling mage targets, waiting up to 5 seconds for cleanup...\nexiting mage\nError: exit forced\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, "ignoresSignals", syscall.SIGINT); err == nil {
		t.Fatal("expected an error because of force kill")
	}
	got = stderr.String()
	want = "cancelling mage targets, waiting up to 5 seconds for cleanup...\nError: cleanup timeout exceeded\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}
}
