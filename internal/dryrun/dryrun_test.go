package dryrun

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// These tests verify dry-run behavior by spawning a fresh process of the
// current test binary with purpose-built helper flags defined in this package's
// TestMain (see testmain_test.go). Spawning a new process ensures the
// sync.Once guards inside dryrun.go evaluate environment variables afresh.

func TestIsDryRunRequestedEnv(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-printIsDryRunRequested")
	cmd.Env = append(os.Environ(), RequestedEnv+"=1", PossibleEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("expected true, got %q", strings.TrimSpace(string(out)))
	}
}

func TestIsDryRunPossibleEnv(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-printIsDryRunPossible")
	cmd.Env = append(os.Environ(), PossibleEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("expected true, got %q", strings.TrimSpace(string(out)))
	}
}

func TestIsDryRunRequiresBoth(t *testing.T) {
	// Only requested set => not possible, so overall false
	cmd := exec.Command(os.Args[0], "-printIsDryRun")
	cmd.Env = append(os.Environ(), RequestedEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "false" {
		t.Fatalf("expected false, got %q", strings.TrimSpace(string(out)))
	}

	// Only possible set => not requested, so overall false
	cmd = exec.Command(os.Args[0], "-printIsDryRun")
	cmd.Env = append(os.Environ(), PossibleEnv+"=1")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "false" {
		t.Fatalf("expected false, got %q", strings.TrimSpace(string(out)))
	}

	// Both set => true
	cmd = exec.Command(os.Args[0], "-printIsDryRun")
	cmd.Env = append(os.Environ(), RequestedEnv+"=1", PossibleEnv+"=1")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("expected true, got %q", strings.TrimSpace(string(out)))
	}
}
