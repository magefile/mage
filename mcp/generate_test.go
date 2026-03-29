package mcp

import (
	"strings"
	"testing"

	"github.com/magefile/mage/parse"
)

func TestMCPStructDef_NoArgs(t *testing.T) {
	f := &parse.Function{
		Name: "Build",
	}
	result := mcpStructDef(f)
	if result != "" {
		t.Errorf("expected empty string for no-arg function, got %q", result)
	}
}

func TestMCPStructDef_StringArg(t *testing.T) {
	f := &parse.Function{
		Name: "Deploy",
		Args: []parse.Arg{
			{Name: "env", Type: "string"},
		},
	}
	result := mcpStructDef(f)
	if !strings.Contains(result, "type _mcpInput_deploy struct") {
		t.Errorf("expected struct definition, got %q", result)
	}
	if !strings.Contains(result, "Env string") {
		t.Errorf("expected Env field, got %q", result)
	}
	if !strings.Contains(result, `json:"env"`) {
		t.Errorf("expected json tag, got %q", result)
	}
}

func TestMCPStructDef_MultipleArgs(t *testing.T) {
	f := &parse.Function{
		Name: "Deploy",
		Args: []parse.Arg{
			{Name: "env", Type: "string"},
			{Name: "port", Type: "int"},
			{Name: "verbose", Type: "bool"},
		},
	}
	result := mcpStructDef(f)
	if !strings.Contains(result, "Env string") {
		t.Errorf("expected Env field, got %q", result)
	}
	if !strings.Contains(result, "Port int") {
		t.Errorf("expected Port field, got %q", result)
	}
	if !strings.Contains(result, "Verbose bool") {
		t.Errorf("expected Verbose field, got %q", result)
	}
}

func TestMCPStructDef_OptionalArg(t *testing.T) {
	f := &parse.Function{
		Name: "Greet",
		Args: []parse.Arg{
			{Name: "name", Type: "string"},
			{Name: "greeting", Type: "string", Optional: true},
		},
	}
	result := mcpStructDef(f)
	if !strings.Contains(result, "Name string") {
		t.Errorf("expected Name field (non-pointer), got %q", result)
	}
	if !strings.Contains(result, "Greeting *string") {
		t.Errorf("expected Greeting pointer field, got %q", result)
	}
	if !strings.Contains(result, "omitempty") {
		t.Errorf("expected omitempty on optional field, got %q", result)
	}
}

func TestMCPStructDef_DurationArg(t *testing.T) {
	f := &parse.Function{
		Name: "Wait",
		Args: []parse.Arg{
			{Name: "duration", Type: "time.Duration"},
		},
	}
	result := mcpStructDef(f)
	// Duration should be represented as string in the struct.
	if !strings.Contains(result, "Duration string") {
		t.Errorf("expected Duration string field for time.Duration, got %q", result)
	}
	if !strings.Contains(result, "duration string, e.g. 5m30s") {
		t.Errorf("expected duration hint in jsonschema tag, got %q", result)
	}
}

func TestMCPStructDef_Namespace(t *testing.T) {
	f := &parse.Function{
		Name:     "Build",
		Receiver: "Docker",
		Args: []parse.Arg{
			{Name: "tag", Type: "string"},
		},
	}
	result := mcpStructDef(f)
	if !strings.Contains(result, "type _mcpInput_docker_build struct") {
		t.Errorf("expected namespaced struct name with underscores, got %q", result)
	}
}

func TestMCPAddTool_NoArgs(t *testing.T) {
	f := &parse.Function{
		Name:     "Build",
		IsError:  true,
		Synopsis: "Build the project",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, `Name:        "build"`) {
		t.Errorf("expected tool name, got %q", result)
	}
	if !strings.Contains(result, `Description: "Build the project"`) {
		t.Errorf("expected tool description, got %q", result)
	}
	if !strings.Contains(result, "_ struct{}") {
		t.Errorf("expected struct{} input for no-arg function, got %q", result)
	}
	if !strings.Contains(result, "_fnErr := Build()") {
		t.Errorf("expected error-returning call, got %q", result)
	}
	if !strings.Contains(result, "IsError: true") {
		t.Errorf("expected IsError handling, got %q", result)
	}
}

func TestMCPAddTool_NoError(t *testing.T) {
	f := &parse.Function{
		Name:     "Clean",
		IsError:  false,
		Synopsis: "Clean up",
	}
	result := mcpAddTool(f)
	// Should not assign to _fnErr.
	if strings.Contains(result, "_fnErr") {
		t.Errorf("expected no error handling for non-error function, got %q", result)
	}
	if !strings.Contains(result, "Clean()") {
		t.Errorf("expected call to Clean(), got %q", result)
	}
}

func TestMCPAddTool_WithContext(t *testing.T) {
	f := &parse.Function{
		Name:      "Build",
		IsError:   true,
		IsContext:  true,
		Synopsis:  "Build the project",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, "Build(ctx)") {
		t.Errorf("expected context passed to function, got %q", result)
	}
}

func TestMCPAddTool_WithArgs(t *testing.T) {
	f := &parse.Function{
		Name:    "Deploy",
		IsError: true,
		Args: []parse.Arg{
			{Name: "env", Type: "string"},
			{Name: "port", Type: "int"},
		},
		Synopsis: "Deploy the application",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, "input _mcpInput_deploy") {
		t.Errorf("expected input struct type, got %q", result)
	}
	if !strings.Contains(result, "Deploy(input.Env, input.Port)") {
		t.Errorf("expected args passed from input struct, got %q", result)
	}
}

func TestMCPAddTool_WithContextAndArgs(t *testing.T) {
	f := &parse.Function{
		Name:      "Deploy",
		IsError:   true,
		IsContext:  true,
		Args: []parse.Arg{
			{Name: "env", Type: "string"},
		},
		Synopsis: "Deploy",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, "Deploy(ctx, input.Env)") {
		t.Errorf("expected ctx and args, got %q", result)
	}
}

func TestMCPAddTool_DurationArg(t *testing.T) {
	f := &parse.Function{
		Name:    "Wait",
		IsError: true,
		Args: []parse.Arg{
			{Name: "duration", Type: "time.Duration"},
		},
		Synopsis: "Wait for a duration",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, "time.ParseDuration(input.Duration)") {
		t.Errorf("expected duration parsing, got %q", result)
	}
	if !strings.Contains(result, "Wait(_dur0)") {
		t.Errorf("expected parsed duration passed to function, got %q", result)
	}
}

func TestMCPAddTool_ImportedPackage(t *testing.T) {
	f := &parse.Function{
		Name:    "Build",
		Package: "docker",
		IsError: true,
		Synopsis: "Build docker image",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, "docker.Build()") {
		t.Errorf("expected package-qualified call, got %q", result)
	}
}

func TestMCPAddTool_Receiver(t *testing.T) {
	f := &parse.Function{
		Name:     "Build",
		Receiver: "Docker",
		IsError:  true,
		Synopsis: "Build docker image",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, "Docker{}.Build()") {
		t.Errorf("expected receiver call, got %q", result)
	}
}

func TestMCPAddTool_StdoutCapture(t *testing.T) {
	f := &parse.Function{
		Name:     "Build",
		IsError:  true,
		Synopsis: "Build",
	}
	result := mcpAddTool(f)
	if !strings.Contains(result, "_origStdout := os.Stdout") {
		t.Errorf("expected stdout capture setup, got %q", result)
	}
	if !strings.Contains(result, "os.Stdout = _w") {
		t.Errorf("expected stdout redirect, got %q", result)
	}
	if !strings.Contains(result, "_restoreStdout") {
		t.Errorf("expected _restoreStdout call, got %q", result)
	}
	if !strings.Contains(result, "_r.Close()") {
		t.Errorf("expected pipe reader close, got %q", result)
	}
}

func TestInputStructName(t *testing.T) {
	tests := []struct {
		name     string
		f        *parse.Function
		expected string
	}{
		{
			name:     "simple",
			f:        &parse.Function{Name: "Build"},
			expected: "_mcpInput_build",
		},
		{
			name:     "namespace",
			f:        &parse.Function{Name: "Build", Receiver: "Docker"},
			expected: "_mcpInput_docker_build",
		},
		{
			name:     "with package alias",
			f:        &parse.Function{Name: "Build", PkgAlias: "tools"},
			expected: "_mcpInput_tools_build",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inputStructName(tt.f)
			if got != tt.expected {
				t.Errorf("inputStructName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExportName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "Name"},
		{"env", "Env"},
		{"", ""},
		{"Port", "Port"},
	}
	for _, tt := range tests {
		got := exportName(tt.input)
		if got != tt.expected {
			t.Errorf("exportName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestJSONType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"int", "int"},
		{"float64", "float64"},
		{"bool", "bool"},
		{"time.Duration", "string"},
		{"unknown", "string"},
	}
	for _, tt := range tests {
		got := jsonType(tt.input)
		if got != tt.expected {
			t.Errorf("jsonType(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
