package mcp

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/magefile/mage/parse"
)

func TestTemplateGeneratesValidGo(t *testing.T) {
	info := &parse.PkgInfo{
		Description: "Test magefiles",
		Funcs: parse.Functions{
			{
				Name:     "Build",
				IsError:  true,
				Synopsis: "Build the project",
				Comment:  "Build compiles the project.",
			},
			{
				Name:     "Clean",
				IsError:  false,
				Synopsis: "Clean up build artifacts",
			},
			{
				Name:      "Deploy",
				IsError:   true,
				IsContext:  true,
				Synopsis:  "Deploy the application",
				Args: []parse.Arg{
					{Name: "env", Type: "string"},
					{Name: "port", Type: "int"},
				},
			},
			{
				Name:    "Wait",
				IsError: true,
				Args: []parse.Arg{
					{Name: "duration", Type: "time.Duration"},
				},
				Synopsis: "Wait for a duration",
			},
			{
				Name:    "Greet",
				IsError: false,
				Args: []parse.Arg{
					{Name: "name", Type: "string"},
					{Name: "greeting", Type: "string", Optional: true},
				},
				Synopsis: "Greet someone",
			},
		},
		Aliases: map[string]*parse.Function{},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp_test_output.go")

	err := GenerateMainfile("test-mcp", path, info)
	if err != nil {
		t.Fatalf("GenerateMainfile failed: %v", err)
	}

	// Read the generated file and verify it contains expected patterns.
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}
	src := string(content)

	// Check key patterns exist in the generated code.
	patterns := []string{
		"package main",
		`_mcp "github.com/modelcontextprotocol/go-sdk/mcp"`,
		`_mcp.NewServer`,
		`_mcp.AddTool`,
		`_mcp.StdioTransport`,
		`Name:        "build"`,
		`Name:        "clean"`,
		`Name:        "deploy"`,
		`Name:        "wait"`,
		`Name:        "greet"`,
		"_mcpInput_deploy",
		"_mcpInput_wait",
		"_mcpInput_greet",
		"Build()",
		"Clean()",
		"Deploy(ctx, input.Env, input.Port)",
		"time.ParseDuration(input.Duration)",
		"Greeting *string",
	}
	for _, p := range patterns {
		if !bytes.Contains(content, []byte(p)) {
			t.Errorf("generated code missing expected pattern: %q\n\nFull output:\n%s", p, src)
		}
	}

	// Verify the generated code parses as valid Go (syntax check only).
	cmd := exec.Command("gofmt", "-e", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nFull source:\n%s", out, src)
	}
}
