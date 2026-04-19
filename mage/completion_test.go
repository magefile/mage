package mage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCompletionBash(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	stdout := &bytes.Buffer{}
	err := installCompletion(stdout, "bash")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	// Verify the completion script was written
	scriptPath := filepath.Join(home, ".config", "mage", "completion.bash")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal("completion script not found:", err)
	}
	if !strings.Contains(string(content), "_mage_completions") {
		t.Error("completion script missing _mage_completions function")
	}
	if !strings.Contains(string(content), "complete -F _mage_completions mage") {
		t.Error("completion script missing complete command")
	}

	// Verify .bashrc was updated
	rcPath := filepath.Join(home, ".bashrc")
	rc, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatal(".bashrc not found:", err)
	}
	rcStr := string(rc)
	if !strings.Contains(rcStr, mageCompletionMarker) {
		t.Error(".bashrc missing completion marker")
	}
	if !strings.Contains(rcStr, "source") {
		t.Error(".bashrc missing source line")
	}
}

func TestInstallCompletionZsh(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	stdout := &bytes.Buffer{}
	err := installCompletion(stdout, "zsh")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	scriptPath := filepath.Join(home, ".config", "mage", "completion.zsh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal("completion script not found:", err)
	}
	if !strings.Contains(string(content), "#compdef mage") {
		t.Error("completion script missing #compdef header")
	}
	if !strings.Contains(string(content), "_mage") {
		t.Error("completion script missing _mage function")
	}

	rcPath := filepath.Join(home, ".zshrc")
	rc, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatal(".zshrc not found:", err)
	}
	if !strings.Contains(string(rc), mageCompletionMarker) {
		t.Error(".zshrc missing completion marker")
	}
}

func TestInstallCompletionFish(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	stdout := &bytes.Buffer{}
	err := installCompletion(stdout, "fish")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	scriptPath := filepath.Join(home, ".config", "fish", "completions", "mage.fish")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal("completion script not found:", err)
	}
	if !strings.Contains(string(content), "complete -c mage") {
		t.Error("completion script missing complete command")
	}
}

func TestInstallCompletionFishXDG(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(home, "custom-config")
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)

	stdout := &bytes.Buffer{}
	err := installCompletion(stdout, "fish")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	scriptPath := filepath.Join(xdg, "fish", "completions", "mage.fish")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Fatal("completion script not found at XDG location:", err)
	}
}

func TestInstallCompletionPowerShell(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	stdout := &bytes.Buffer{}
	err := installCompletion(stdout, "powershell")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	scriptPath := filepath.Join(home, ".config", "mage", "completion.ps1")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatal("completion script not found:", err)
	}
	if !strings.Contains(string(content), "Register-ArgumentCompleter") {
		t.Error("completion script missing Register-ArgumentCompleter")
	}

	// PowerShell should print instructions, not modify files
	output := stdout.String()
	if !strings.Contains(output, "$PROFILE") {
		t.Error("output should contain instructions mentioning $PROFILE")
	}
}

func TestInstallCompletionUnsupportedShell(t *testing.T) {
	err := installCompletion(io.Discard, "tcsh")
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("expected 'unsupported shell' error, got: %v", err)
	}
}

func TestInstallCompletionCaseInsensitive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	stdout := &bytes.Buffer{}
	err := installCompletion(stdout, "BASH")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	scriptPath := filepath.Join(home, ".config", "mage", "completion.bash")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Fatal("completion script not found:", err)
	}
}

func TestAddGuardedBlockNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	err := addGuardedBlock(path, "test content")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, mageCompletionMarker) {
		t.Error("missing start marker")
	}
	if !strings.Contains(s, mageCompletionMarkerEnd) {
		t.Error("missing end marker")
	}
	if !strings.Contains(s, "test content") {
		t.Error("missing content")
	}
}

func TestAddGuardedBlockExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	// Write initial content
	if err := os.WriteFile(path, []byte("existing stuff\n"), 0o600); err != nil {
		t.Fatal("could not write initial content:", err)
	}

	err := addGuardedBlock(path, "test content")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if !strings.Contains(s, "existing stuff") {
		t.Error("lost existing content")
	}
	if !strings.Contains(s, "test content") {
		t.Error("missing new content")
	}
}

func TestAddGuardedBlockReplace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testrc")

	// First install
	err := addGuardedBlock(path, "old content")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	// Reinstall should replace
	err = addGuardedBlock(path, "new content")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	content, _ := os.ReadFile(path)
	s := string(content)
	if strings.Contains(s, "old content") {
		t.Error("old content should have been replaced")
	}
	if !strings.Contains(s, "new content") {
		t.Error("missing new content")
	}
	// Ensure markers appear exactly once
	if strings.Count(s, mageCompletionMarker) != 1 {
		t.Error("expected exactly one start marker")
	}
}

func TestParseInstall(t *testing.T) {
	inv, cmd, err := Parse(io.Discard, io.Discard, []string{"-install", "bash"})
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if cmd != Install {
		t.Errorf("expected Install command, got %v", cmd)
	}
	if inv.InstallShell != "bash" {
		t.Errorf("expected InstallShell 'bash', got %q", inv.InstallShell)
	}
}

func TestParseInstallConflictsWithOtherCommands(t *testing.T) {
	_, _, err := Parse(io.Discard, io.Discard, []string{"-install", "bash", "-autocomplete"})
	if err == nil {
		t.Fatal("expected error when using -install with -autocomplete")
	}
}

func TestCompletionScriptContents(t *testing.T) {
	bin := "/usr/local/bin/mage"

	t.Run("bash", func(t *testing.T) {
		s := bashCompletionScript(bin)
		if !strings.Contains(s, bin) {
			t.Error("script should contain the mage binary path")
		}
		if !strings.Contains(s, "-autocomplete") {
			t.Error("script should call -autocomplete")
		}
	})

	t.Run("zsh", func(t *testing.T) {
		s := zshCompletionScript(bin)
		if !strings.Contains(s, bin) {
			t.Error("script should contain the mage binary path")
		}
		if !strings.Contains(s, "#compdef mage") {
			t.Error("script should have compdef header")
		}
	})

	t.Run("fish", func(t *testing.T) {
		s := fishCompletionScript(bin)
		if !strings.Contains(s, bin) {
			t.Error("script should contain the mage binary path")
		}
		if !strings.Contains(s, "complete -c mage") {
			t.Error("script should have complete commands")
		}
	})

	t.Run("powershell", func(t *testing.T) {
		s := powerShellCompletionScript(bin)
		if !strings.Contains(s, bin) {
			t.Error("script should contain the mage binary path")
		}
		if !strings.Contains(s, "Register-ArgumentCompleter") {
			t.Error("script should have Register-ArgumentCompleter")
		}
	})
}
