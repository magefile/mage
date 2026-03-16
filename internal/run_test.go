package internal

import (
	"runtime"
	"strings"
	"testing"
)

func TestSplitEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     []string
		wantLen int
		wantErr bool
	}{
		{
			name:    "normal env vars",
			env:     []string{"FOO=bar", "BAZ=qux"},
			wantLen: 2,
		},
		{
			name:    "empty value",
			env:     []string{"FOO="},
			wantLen: 1,
		},
		{
			name:    "value with equals sign",
			env:     []string{"FOO=bar=baz"},
			wantLen: 1,
		},
		{
			name:    "empty input",
			env:     nil,
			wantLen: 0,
		},
		{
			name:    "malformed entry",
			env:     []string{"NO_EQUALS"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitEnv(tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("SplitEnv() returned %d entries, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestSplitEnvValues(t *testing.T) {
	env := []string{"FOO=bar", "BAZ=qux=quux"}
	got, err := SplitEnv(env)
	if err != nil {
		t.Fatal(err)
	}
	if got["FOO"] != "bar" {
		t.Errorf("expected FOO=bar, got FOO=%s", got["FOO"])
	}
	if got["BAZ"] != "qux=quux" {
		t.Errorf("expected BAZ=qux=quux, got BAZ=%s", got["BAZ"])
	}
}

func TestEnvWithCurrentGOOS(t *testing.T) {
	env, err := EnvWithCurrentGOOS()
	if err != nil {
		t.Fatal(err)
	}

	var foundGOOS, foundGOARCH bool
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "GOOS":
			foundGOOS = true
			if parts[1] != runtime.GOOS {
				t.Errorf("expected GOOS=%s, got %s", runtime.GOOS, parts[1])
			}
		case "GOARCH":
			foundGOARCH = true
			if parts[1] != runtime.GOARCH {
				t.Errorf("expected GOARCH=%s, got %s", runtime.GOARCH, parts[1])
			}
		default:
			// ignore other env vars
			continue
		}
	}
	if !foundGOOS {
		t.Error("GOOS not found in env")
	}
	if !foundGOARCH {
		t.Error("GOARCH not found in env")
	}
}

func TestRunDebug(t *testing.T) {
	// Test successful command
	err := RunDebug("echo", "hello")
	if err != nil {
		t.Fatalf("RunDebug with valid command failed: %v", err)
	}

	// Test failed command
	err = RunDebug("false")
	if err == nil {
		t.Fatal("RunDebug with failing command should return error")
	}
}

func TestOutputDebug(t *testing.T) {
	// Test successful command
	out, err := OutputDebug("echo", "hello")
	if err != nil {
		t.Fatalf("OutputDebug with valid command failed: %v", err)
	}
	if out != "hello" {
		t.Errorf("expected 'hello', got %q", out)
	}

	// Test failed command
	_, err = OutputDebug("false")
	if err == nil {
		t.Fatal("OutputDebug with failing command should return error")
	}
}
