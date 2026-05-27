// Package sh provides helpers for running shell commands.
package sh

import (
	"context"
	"io"

	"github.com/magefile/mage/shctx"
)

// RunCmd returns a function that will call Run with the given command. This is
// useful for creating command aliases to make your scripts easier to read, like
// this:
//
//	// in a helper file somewhere
//	var g0 = sh.RunCmd("go")  // go is a keyword :(
//
//	// somewhere in your main code
//		if err := g0("install", "github.com/gohugo/hugo"); err != nil {
//			return err
//	}
//
// Args passed to command get baked in as args to the command when you run it.
// Any args passed in when you run the returned function will be appended to the
// original args.  For example, this is equivalent to the above:
//
//	var goInstall = sh.RunCmd("go", "install") goInstall("github.com/gohugo/hugo")
//
// RunCmd uses Exec underneath, so see those docs for more details.
func RunCmd(cmd string, args ...string) func(args ...string) error {
	return func(args2 ...string) error {
		return Run(cmd, append(args, args2...)...)
	}
}

// OutCmd is like RunCmd except the command returns the output of the
// command.
func OutCmd(cmd string, args ...string) func(args ...string) (string, error) {
	return func(args2 ...string) (string, error) {
		return Output(cmd, append(args, args2...)...)
	}
}

// Run is like RunWith, but doesn't specify any environment variables.
func Run(cmd string, args ...string) error {
	return shctx.Run(context.Background(), cmd, args...)
}

// RunV is like Run, but always sends the command's stdout to os.Stdout.
func RunV(cmd string, args ...string) error {
	return shctx.RunV(context.Background(), cmd, args...)
}

// RunWith runs the given command, directing stderr to this program's stderr and
// printing stdout to stdout if mage was run with -v.  It adds adds env to the
// environment variables for the command being run. Environment variables should
// be in the format name=value.
func RunWith(env map[string]string, cmd string, args ...string) error {
	return shctx.RunWith(context.Background(), env, cmd, args...)
}

// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithV(env map[string]string, cmd string, args ...string) error {
	return shctx.RunWithV(context.Background(), env, cmd, args...)
}

// Output runs the command and returns the text from stdout.
func Output(cmd string, args ...string) (string, error) {
	return shctx.Output(context.Background(), cmd, args...)
}

// OutputWith is like RunWith, but returns what is written to stdout.
func OutputWith(env map[string]string, cmd string, args ...string) (string, error) {
	return shctx.OutputWith(context.Background(), env, cmd, args...)
}

// Exec executes the command, piping its stdout and stderr to the given
// writers. If the command fails, it will return an error that, if returned
// from a target or mg.Deps call, will cause mage to exit with the same code as
// the command failed with. Env is a list of environment variables to set when
// running the command, these override the current environment variables set
// (which are also passed to the command). cmd and args may include references
// to environment variables in $FOO format, in which case these will be
// expanded before the command is run.
//
// Ran reports if the command ran (rather than was not found or not executable).
// Code reports the exit code the command returned if it ran. If err == nil, ran
// is always true and code is always 0.
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, err error) {
	return shctx.Exec(context.Background(), env, stdout, stderr, cmd, args...)
}

// CmdRan examines the error to determine if it was generated as a result of a
// command running via os/exec.Command.  If the error is nil, or the command ran
// (even if it exited with a non-zero exit code), CmdRan reports true.  If the
// error is an unrecognized type, or it is an error from exec.Command that says
// the command failed to run (usually due to the command not existing or not
// being executable), it reports false.
func CmdRan(err error) bool {
	return shctx.CmdRan(err)
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
// 0 if it is nil or 1 if it is a different error.
func ExitStatus(err error) int {
	return shctx.ExitStatus(err)
}
