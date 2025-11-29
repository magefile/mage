package dryrun

import (
	"os"
	"os/exec"
	"sync"
)

// RequestedEnv is the environment variable that indicates the user requested dryrun mode when running mage.
const RequestedEnv = "MAGEFILE_DRYRUN"

// PossibleEnv is the environment variable that indicates we are in a context where a dry run is possible.
const PossibleEnv = "MAGEFILE_DRYRUN_POSSIBLE"

var (
	// Once-protected variables for whether the user requested dryrun mode.
	dryRunRequestedValue    bool
	dryRunRequestedEnvValue bool
	dryRunRequestedEnvOnce  sync.Once

	// Once-protected variables for whether dryrun mode is possible.
	dryRunPossible     bool
	dryRunPossibleOnce sync.Once
)

// SetRequested sets the dryrun requested state to the specified boolean value.
func SetRequested(value bool) {
	dryRunRequestedValue = value
}

// IsRequested checks if dry-run mode was requested, either explicitly or via an environment variable.
func IsRequested() bool {
	dryRunRequestedEnvOnce.Do(func() {
		if os.Getenv(RequestedEnv) != "" {
			dryRunRequestedEnvValue = true
		}
	})

	return dryRunRequestedEnvValue || dryRunRequestedValue
}

// IsPossible checks if dry-run mode is supported in the current context.
func IsPossible() bool {
	dryRunPossibleOnce.Do(func() {
		dryRunPossible = os.Getenv(PossibleEnv) != ""
	})

	return dryRunPossible
}

// Wrap creates an *exec.Cmd to run a command or simulate it in dry-run mode.
// If not in dry-run mode, it returns exec.Command(cmd, args...).
// In dry-run mode, it returns a command that prints the simulated command.
func Wrap(cmd string, args ...string) *exec.Cmd {
	if !IsDryRun() {
		return exec.Command(cmd, args...)
	}

	// Return an *exec.Cmd that just prints the command that would have been run.
	return exec.Command("echo", append([]string{"DRYRUN: " + cmd}, args...)...)
}

// IsDryRun determines if dry-run mode is both possible and requested.
func IsDryRun() bool {
	return IsPossible() && IsRequested()
}
