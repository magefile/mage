package shctx

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestOutCmd(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	cmd := OutCmd(os.Args[0], "-printArgs", "foo", "bar")
	out, err := cmd(ctx, "baz", "bat")
	if err != nil {
		t.Fatal(err)
	}
	expected := "[foo bar baz bat]"
	if out != expected {
		t.Fatalf("expected %q but got %q", expected, out)
	}
}

func TestExitCode(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	ran, err := Exec(ctx, nil, nil, nil, os.Args[0], "-helper", "-exit", "99")
	if err == nil {
		t.Fatal("unexpected nil error from run")
	}
	if !ran {
		t.Error("ran returned as false, but should have been true")
	}
	code := ExitStatus(err)
	if code != 99 {
		t.Fatalf("expected exit status 99, but got %v", code)
	}
}

func TestEnv(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	env := "SOME_REALLY_LONG_MAGEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(
		ctx,
		map[string]string{env: "foobar"},
		out,
		nil,
		os.Args[0], "-printVar", env,
	)
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Error("expected ran to be true but was false.")
	}
	if out.String() != "foobar\n" {
		t.Errorf("expected foobar, got %q", out)
	}
}

func TestNotRun(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	ran, err := Exec(ctx, nil, nil, nil, "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	t.Setenv("MAGE_FOOBAR", "baz")
	s, err := Output(ctx, "echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}
}

func TestContextTimeout(t *testing.T) {
	deadline := time.Now().Add(100 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	err := Run(ctx, os.Args[0], "-sleep", "1")

	// Check that the command was aborted roughly at the right time. This way it's likely it was
	// caused by the context cancellation, and not some other cause.
	timeSinceDeadline := time.Since(deadline)
	switch {
	case timeSinceDeadline < -10*time.Millisecond:
		t.Fatalf("command exited unexpectedly early (%v before the deadline)", -timeSinceDeadline)
	case timeSinceDeadline > 100*time.Millisecond:
		t.Fatalf("command exited unexpectedly late (%v after the deadline)", timeSinceDeadline)
	}

	// Check the exit status, the command shouldn't have run succesfully as it was cancelled.
	var exitError *exec.ExitError
	switch {
	case errors.As(err, &exitError):
		// Expected case.
	case errors.Is(err, context.DeadlineExceeded):
		// This shouldn't happen, as the timeout of the context should only happen after the
		// subprocess is already running. But if it's returned, it's technically not wrong.
	case err == nil:
		t.Fatal("unexpected nil error from run")
	default:
		wrapped := errors.Unwrap(err)
		t.Fatalf("unexpected error: %v of type %T wrapping %v %T", err, err, wrapped, wrapped)
	}
}

func TestCmdRanNilErr(t *testing.T) {
	if !CmdRan(nil) {
		t.Fatal("CmdRan(nil) should return true")
	}
}

func TestCmdRanNotFound(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	_, err := Exec(ctx, nil, nil, nil, "thiswontwork")
	if CmdRan(err) {
		t.Fatal("CmdRan should return false for not-found command")
	}
}

func TestExitStatusNil(t *testing.T) {
	code := ExitStatus(nil)
	if code != 0 {
		t.Fatalf("expected 0 for nil error, got %d", code)
	}
}

func TestExitStatusNonExecError(t *testing.T) {
	code := ExitStatus(errors.New("generic error"))
	if code != 1 {
		t.Fatalf("expected 1 for generic error, got %d", code)
	}
}

func TestExitStatusFromExec(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	_, err := Exec(ctx, nil, nil, nil, os.Args[0], "-helper", "-exit", "42")
	code := ExitStatus(err)
	if code != 42 {
		t.Fatalf("expected exit status 42, got %d", code)
	}
}

func TestRunCmd(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	echoHello := RunCmd("echo", "hello")
	err := echoHello(ctx, "world")
	// RunWith directs output based on verbose, so just check no error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutputWith(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()

	out, err := OutputWith(
		ctx,
		map[string]string{"MY_TEST_VAR": "xyz"},
		os.Args[0], "-printVar", "MY_TEST_VAR",
	)
	if err != nil {
		t.Fatal(err)
	}
	if out != "xyz" {
		t.Fatalf("expected 'xyz', got %q", out)
	}
}

// testCtx returns a context and cancel function for general testing.
// The timeout is long enough for any command to run on a typical system.
func testCtx() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
