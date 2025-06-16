package sh

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/magefile/mage/mg"
)

// RunCmd returns a function that will call Run with the given command. This is
// useful for creating command aliases to make your scripts easier to read, like
// this:
//
//  // in a helper file somewhere
//  var g0 = sh.RunCmd("go")  // go is a keyword :(
//
//  // somewhere in your main code
//	if err := g0("install", "github.com/gohugo/hugo"); err != nil {
//		return err
//  }
//
// Args passed to command get baked in as args to the command when you run it.
// Any args passed in when you run the returned function will be appended to the
// original args.  For example, this is equivalent to the above:
//
//  var goInstall = sh.RunCmd("go", "install") goInstall("github.com/gohugo/hugo")
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
	return RunWith(nil, cmd, args...)
}

// RunV is like Run, but always sends the command's stdout to os.Stdout.
func RunV(cmd string, args ...string) error {
	_, err := Exec(nil, os.Stdout, os.Stderr, cmd, args...)
	return err
}

// RunWith runs the given command, directing stderr to this program's stderr and
// printing stdout to stdout if mage was run with -v.  It adds adds env to the
// environment variables for the command being run. Environment variables should
// be in the format name=value.
func RunWith(env map[string]string, cmd string, args ...string) error {
	var output io.Writer
	if mg.Verbose() {
		output = os.Stdout
	}
	_, err := Exec(env, output, os.Stderr, cmd, args...)
	return err
}

// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithV(env map[string]string, cmd string, args ...string) error {
	_, err := Exec(env, os.Stdout, os.Stderr, cmd, args...)
	return err
}

// Output runs the command and returns the text from stdout.
func Output(cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := Exec(nil, buf, os.Stderr, cmd, args...)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// OutputWith is like RunWith, but returns what is written to stdout.
func OutputWith(env map[string]string, cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := Exec(env, buf, os.Stderr, cmd, args...)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// Exec executes the command, piping its stdout and stderr to the given
// writers. If the command fails, it will return an error that, if returned
// from a target or mg.Deps call, will cause mage to exit with the same code as
// the command failed with. Env is a list of environment variables to set when
// running the command, these override the current environment variables set
// (which are also passed to the command). cmd and args may include references
// to environment variables in $FOO or ${FOO} format, in which case these will
// be expanded before the command is run; use a backslash before the dollar
// sign to avoid this behavior.
//
// Ran reports if the command ran (rather than was not found or not executable).
// Code reports the exit code the command returned if it ran. If err == nil, ran
// is always true and code is always 0.
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, err error) {
	expand := func(s string) string {
		s2, ok := env[s]
		if ok {
			return s2
		}
		return os.Getenv(s)
	}
	cmd = Expand(cmd, expand)
	for i := range args {
		args[i] = Expand(args[i], expand)
	}
	ran, code, err := run(env, stdout, stderr, cmd, args...)
	if err == nil {
		return true, nil
	}
	if ran {
		return ran, mg.Fatalf(code, `running "%s %s" failed with exit code %d`, cmd, strings.Join(args, " "), code)
	}
	return ran, fmt.Errorf(`failed to run "%s %s: %v"`, cmd, strings.Join(args, " "), err)
}

func run(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, code int, err error) {
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
	}
	c.Stderr = stderr
	c.Stdout = stdout
	c.Stdin = os.Stdin

	var quoted []string 
	for i := range args {
		quoted = append(quoted, fmt.Sprintf("%q", args[i]));
	}
	// To protect against logging from doing exec in global variables
	if mg.Verbose() {
		log.Println("exec:", cmd, strings.Join(quoted, " "))
	}
	err = c.Run()
	return CmdRan(err), ExitStatus(err), err
}
// CmdRan examines the error to determine if it was generated as a result of a
// command running via os/exec.Command.  If the error is nil, or the command ran
// (even if it exited with a non-zero exit code), CmdRan reports true.  If the
// error is an unrecognized type, or it is an error from exec.Command that says
// the command failed to run (usually due to the command not existing or not
// being executable), it reports false.
func CmdRan(err error) bool {
	if err == nil {
		return true
	}
	ee, ok := err.(*exec.ExitError)
	if ok {
		return ee.Exited()
	}
	return false
}

type exitStatus interface {
	ExitStatus() int
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
// 0 if it is nil or 1 if it is a different error.
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(exitStatus); ok {
		return e.ExitStatus()
	}
	if e, ok := err.(*exec.ExitError); ok {
		if ex, ok := e.Sys().(exitStatus); ok {
			return ex.ExitStatus()
		}
	}
	return 1
}

// Escape returns an escaped version of the argument such that when environment
// expansion occurs in the command-running functions of this module, the result
// will be the original argument.
func Escape(arg string) string {
	arg = strings.Replace(arg, `\`, `\\`, -1)
	arg = strings.Replace(arg, "$", `\$`, -1)
	return arg
}

var varExpr = regexp.MustCompile(`\\.|\$(\w+|[-*#$@!?0-9])|\$\{(\w+|[-*#$@!?0-9])\}|\$\{\}?`)

// Expand searches the input for segments of the form $var or ${var} and
// replaces them with the value returned by the given callback function, such
// as `os.Getenv]. It works just like [os.Expand], except that a backslash
// preceeding the dollar sign will escape it, allowing a literal dollar sign to
// appear in the output. Escape a backslash with another backslash. A backslash
// followed by any other character is reserved for future use; it may be
// omitted from the output or replaced by some other value determined in the
// future.
func Expand(s string, mapping func(string) string) string {
	return varExpr.ReplaceAllStringFunc(s, func(match string) string {
		switch match[0] {
		case '\\':
			// Escaped backslash or dollar expands to itself.
			// Escaped anything else is reserved and gets removed.
			if match[1] == '\\' || match[1] == '$' {
				return match[1:2]
			}
		case '$':
			if match[1] != '{' {
				// We got an ordinary word. Omit the dollar and
				// do the replacement.
				return mapping(match[1:])
			}
			if len(match) > 3 {
				// Omit the leading "${" and trailing "}" to
				// get the name.
				return mapping(match[2:len(match)-1])
			}
			// We got either "${" or "${}". They're both syntax
			// errors.
		default:
			// Should never get here.
		}
		return ""
	})
}
