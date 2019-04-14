package mg

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// CacheEnv is the environment variable that users may set to change the
// location where mage stores its compiled binaries.
const CacheEnv = "MAGEFILE_CACHE"

// VerboseEnv is the environment variable that indicates the user requested
// verbose mode when running a magefile.
const VerboseEnv = "MAGEFILE_VERBOSE"

// DebugEnv is the environment variable that indicates the user requested
// debug mode when running mage.
const DebugEnv = "MAGEFILE_DEBUG"

// GoCmdEnv is the environment variable that indicates the go binary the user
// desires to utilize for Magefile compilation.
const GoCmdEnv = "MAGEFILE_GOCMD"

// IgnoreDefaultEnv is the environment variable that indicates the user requested
// to ignore the default target specified in the magefile.
const IgnoreDefaultEnv = "MAGEFILE_IGNOREDEFAULT"

// MageTargetArgsPrefix -
const MageTargetArgsPrefix = "mage-target-args:"

// MageStopProcArgsPrefix -
const MageStopProcArgsPrefix = "mage-stop:"

// Verbose reports whether a magefile was run with the verbose flag.
func Verbose() bool {
	b, _ := strconv.ParseBool(os.Getenv(VerboseEnv))
	return b
}

// Debug reports whether a magefile was run with the verbose flag.
func Debug() bool {
	b, _ := strconv.ParseBool(os.Getenv(DebugEnv))
	return b
}

// GoCmd reports the command that Mage will use to build go code.  By default mage runs
// the "go" binary in the PATH.
func GoCmd() string {
	if cmd := os.Getenv(GoCmdEnv); cmd != "" {
		return cmd
	}
	return "go"
}

// IgnoreDefault reports whether the user has requested to ignore the default target
// in the magefile.
func IgnoreDefault() bool {
	b, _ := strconv.ParseBool(os.Getenv(IgnoreDefaultEnv))
	return b
}

// CacheDir returns the directory where mage caches compiled binaries.  It
// defaults to $HOME/.magefile, but may be overridden by the MAGEFILE_CACHE
// environment variable.
func CacheDir() string {
	d := os.Getenv(CacheEnv)
	if d != "" {
		return d
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), "magefile")
	default:
		return filepath.Join(os.Getenv("HOME"), ".magefile")
	}
}

// Namespace allows for the grouping of similar commands
type Namespace struct{}

// GetArgsForTarget returns array of arguments passed to given target from CLI,
// this array can be then processed by flags package by each target's code
// This function simplify the way in which you can obtain per target CLI args,
// without the need of calling this function cli args can be obtained using os.Args
// but those are encoded in the following format:
//-----------------------------------------------
// mage-target-args:main.target:--flag=value
// where:
// mage-target-args: - prefix added to non target arguments (args not matching any target or alias name)
// main.target:      - prefix based on target name, allows to indentify per target args
// --flag=value      - bash style flag in -f or --flag format
// value             - can be boolean, string, integer with or without quotes, for file paths
//                     passed as values please use single quotes to wrap it
//-----------------------------------------------------------------------------
// supported flags formats are:
//-----------------------------
// --                                          : stop processing any further arguments
// -f --f -flag --flag                         : boolean flags only
// -f=value --f=value -flag=value --flag=value : boolean, integer, string flags
// f=value flag=value                          : boolean, integer, string flags (read above below note)
//-----------------------------------------------------------------------------
// Note: for args passed in format: var=value or v=value there is implicit
// string "--" added at the beginning of it, this is to satisfy flags package
//-----------------------------------------------------------------------------
func GetArgsForTarget() []string {
	var args []string
	//--------------------------------------------
	// Get info about the name of calling function
	// by checking callstack using runtime package
	// Then use this information to identify cli
	// args for given target (calling function)
	//--------------------------------------------
	programCounter := make([]uintptr, 15)
	entriesCount := runtime.Callers(2, programCounter)
	frames := runtime.CallersFrames(programCounter[:entriesCount])
	frame, _ := frames.Next()
	// add ":" to the name of calling function, to form target prefix
	callingFunction := strings.ToLower(frame.Function) + ":"
	//--------------------------------------------
	// Loop trough all os.Args and check it's prefixes
	args = []string{}
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, MageStopProcArgsPrefix) {
			// if we found "mage-stop:" prefix then don't
			// process any more args, and return what has
			// been already collected
			return args
		}
		// first, look for args with prefix mage-target-args:
		if strings.HasPrefix(arg, MageTargetArgsPrefix) {
			arg = strings.TrimPrefix(arg, MageTargetArgsPrefix)
			// then, look for args with prefix main.callingFunction
			if strings.HasPrefix(arg, callingFunction) {
				arg = strings.TrimPrefix(arg, callingFunction)
				if !strings.HasPrefix(arg, "-") {
					// if given arg is missing "-" prefix (so it's in format var=value)
					// then we have to implicitly append "--" in front of it
					// in order to satisfy flags package
					arg = "--" + arg
				}
				// here we append raw flag without any prefixes added earlier
				// this can be easily parsed then by flags by each individial target
				args = append(args, arg)
			}
		}
	}
	return args
}
