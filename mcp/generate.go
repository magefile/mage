// Package mcp provides MCP server code generation for mage targets.
package mcp

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/magefile/mage/parse"
)

// mainfileTemplateData matches the data structure expected by the template.
// It is identical to the one in mage/main.go.
type mainfileTemplateData struct {
	Description string
	Funcs       []*parse.Function
	DefaultFunc parse.Function
	Aliases     map[string]*parse.Function
	Imports     []*parse.Import
	BinaryName  string
}

var mcpMainfileTemplate = template.Must(template.New("").Funcs(template.FuncMap{
	"lower":        strings.ToLower,
	"mcpStructDef": mcpStructDef,
	"mcpAddTool":   mcpAddTool,
}).Parse(mcpMainfileTplString))

// GenerateMainfile generates the MCP server mainfile at path.
func GenerateMainfile(binaryName, path string, info *parse.PkgInfo) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating generated MCP mainfile: %w", err)
	}
	defer f.Close()

	data := mainfileTemplateData{
		Description: info.Description,
		Funcs:       info.Funcs,
		Aliases:     info.Aliases,
		Imports:     info.Imports,
		BinaryName:  binaryName,
	}

	if info.DefaultFunc != nil {
		data.DefaultFunc = *info.DefaultFunc
	}

	if err := mcpMainfileTemplate.Execute(f, data); err != nil {
		return fmt.Errorf("can't execute MCP mainfile template: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing generated MCP mainfile: %w", err)
	}
	// Set an old modtime on the generated mainfile so that the go tool
	// won't think it has changed more recently than the compiled binary.
	longAgo := time.Now().Add(-time.Hour * 24 * 365 * 10)
	if err := os.Chtimes(path, longAgo, longAgo); err != nil {
		return fmt.Errorf("error setting old modtime on generated MCP mainfile: %w", err)
	}
	return nil
}

// mcpStructDef generates an input struct type definition for a function's MCP tool.
// Returns empty string if the function has no arguments.
func mcpStructDef(f *parse.Function) string {
	if len(f.Args) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "type %s struct {\n", inputStructName(f))
	for _, arg := range f.Args {
		jsonType := jsonType(arg.Type)
		fieldName := exportName(arg.Name)

		tag := fmt.Sprintf("`"+`json:"%s`, strings.ToLower(arg.Name))
		if arg.Optional {
			tag += ",omitempty"
		}
		tag += `" jsonschema:"` + arg.Name
		if arg.Type == "time.Duration" {
			tag += " (duration string, e.g. 5m30s)"
		}
		tag += `"` + "`"

		fieldType := jsonType
		if arg.Optional {
			fieldType = "*" + jsonType
		}

		fmt.Fprintf(&b, "\t%s %s %s\n", fieldName, fieldType, tag)
	}
	b.WriteString("}\n")
	return b.String()
}

// mcpAddTool generates the mcp.AddTool call for a function.
func mcpAddTool(f *parse.Function) string {
	var b strings.Builder

	// Determine the target name and description.
	targetName := strings.ToLower(f.TargetName())
	desc := f.Synopsis
	if desc == "" {
		desc = f.Name
	}

	// Determine input type.
	inputType := "struct{}"
	inputParam := "_"
	if len(f.Args) > 0 {
		inputType = inputStructName(f)
		inputParam = "input"
	}

	// Build the qualified function name.
	funcName := f.Name
	if f.Receiver != "" {
		funcName = f.Receiver + "{}." + funcName
	}
	if f.Package != "" {
		funcName = f.Package + "." + funcName
	}

	// Start the AddTool call.
	fmt.Fprintf(&b, "_mcp.AddTool(server, &_mcp.Tool{\n")
	fmt.Fprintf(&b, "\t\tName:        %q,\n", targetName)
	fmt.Fprintf(&b, "\t\tDescription: %q,\n", desc)
	fmt.Fprintf(&b, "\t}, func(ctx context.Context, req *_mcp.CallToolRequest, %s %s) (*_mcp.CallToolResult, any, error) {\n", inputParam, inputType)

	// Capture stdout so target output doesn't corrupt the MCP protocol on stdout.
	b.WriteString("\t\t_origStdout := os.Stdout\n")
	b.WriteString("\t\t_r, _w, _ := os.Pipe()\n")
	b.WriteString("\t\tos.Stdout = _w\n")
	b.WriteString("\t\t_restoreStdout := func() string {\n")
	b.WriteString("\t\t\t_w.Close()\n")
	b.WriteString("\t\t\tos.Stdout = _origStdout\n")
	b.WriteString("\t\t\tvar _buf bytes.Buffer\n")
	b.WriteString("\t\t\t_buf.ReadFrom(_r)\n")
	b.WriteString("\t\t\t_r.Close()\n")
	b.WriteString("\t\t\treturn _buf.String()\n")
	b.WriteString("\t\t}\n\n")

	// Parse time.Duration args from strings.
	for i, arg := range f.Args {
		if arg.Type != "time.Duration" {
			continue
		}
		fieldName := "input." + exportName(arg.Name)
		if arg.Optional {
			fmt.Fprintf(&b, "\t\tvar _dur%d *time.Duration\n", i)
			fmt.Fprintf(&b, "\t\tif %s != nil {\n", fieldName)
			fmt.Fprintf(&b, "\t\t\t_d, _derr := time.ParseDuration(*%s)\n", fieldName)
			fmt.Fprintf(&b, "\t\t\tif _derr != nil {\n")
			fmt.Fprintf(&b, "\t\t\t\t_restoreStdout()\n")
			fmt.Fprintf(&b, "\t\t\t\treturn &_mcp.CallToolResult{\n")
			fmt.Fprintf(&b, "\t\t\t\t\tContent: []_mcp.Content{&_mcp.TextContent{Text: _fmt.Sprintf(\"invalid duration for %s: %%v\", _derr)}},\n", arg.Name)
			fmt.Fprintf(&b, "\t\t\t\t\tIsError: true,\n")
			fmt.Fprintf(&b, "\t\t\t\t}, nil, nil\n")
			fmt.Fprintf(&b, "\t\t\t}\n")
			fmt.Fprintf(&b, "\t\t\t_dur%d = &_d\n", i)
			fmt.Fprintf(&b, "\t\t}\n")
		} else {
			fmt.Fprintf(&b, "\t\t_dur%d, _derr%d := time.ParseDuration(%s)\n", i, i, fieldName)
			fmt.Fprintf(&b, "\t\tif _derr%d != nil {\n", i)
			fmt.Fprintf(&b, "\t\t\t_restoreStdout()\n")
			fmt.Fprintf(&b, "\t\t\treturn &_mcp.CallToolResult{\n")
			fmt.Fprintf(&b, "\t\t\t\tContent: []_mcp.Content{&_mcp.TextContent{Text: _fmt.Sprintf(\"invalid duration for %s: %%v\", _derr%d)}},\n", arg.Name, i)
			fmt.Fprintf(&b, "\t\t\t\tIsError: true,\n")
			fmt.Fprintf(&b, "\t\t\t}, nil, nil\n")
			fmt.Fprintf(&b, "\t\t}\n")
		}
	}

	// Build the function call.
	b.WriteString("\t\t")
	if f.IsError {
		b.WriteString("_fnErr := ")
	}
	b.WriteString(funcName + "(")
	var callArgs []string
	if f.IsContext {
		callArgs = append(callArgs, "ctx")
	}
	for i, arg := range f.Args {
		if arg.Type == "time.Duration" {
			callArgs = append(callArgs, fmt.Sprintf("_dur%d", i))
		} else {
			callArgs = append(callArgs, "input."+exportName(arg.Name))
		}
	}
	b.WriteString(strings.Join(callArgs, ", "))
	b.WriteString(")\n\n")

	// Close pipe, restore stdout, read captured output.
	b.WriteString("\t\t_output := _restoreStdout()\n\n")

	// Handle error return.
	if f.IsError {
		b.WriteString("\t\tif _fnErr != nil {\n")
		b.WriteString("\t\t\t_msg := _fnErr.Error()\n")
		b.WriteString("\t\t\tif _output != \"\" {\n")
		b.WriteString("\t\t\t\t_msg = _output + \"\\n\" + _msg\n")
		b.WriteString("\t\t\t}\n")
		b.WriteString("\t\t\treturn &_mcp.CallToolResult{\n")
		b.WriteString("\t\t\t\tContent: []_mcp.Content{&_mcp.TextContent{Text: _msg}},\n")
		b.WriteString("\t\t\t\tIsError: true,\n")
		b.WriteString("\t\t\t}, nil, nil\n")
		b.WriteString("\t\t}\n\n")
	}

	// Return success.
	fmt.Fprintf(&b, "\t\tif _output == \"\" {\n")
	fmt.Fprintf(&b, "\t\t\t_output = %q\n", targetName+" completed successfully")
	fmt.Fprintf(&b, "\t\t}\n")
	b.WriteString("\t\treturn &_mcp.CallToolResult{\n")
	b.WriteString("\t\t\tContent: []_mcp.Content{&_mcp.TextContent{Text: _output}},\n")
	b.WriteString("\t\t}, nil, nil\n")
	b.WriteString("\t})")

	return b.String()
}

// inputStructName returns a unique Go identifier for the input struct of a function.
func inputStructName(f *parse.Function) string {
	name := strings.ReplaceAll(f.TargetName(), ":", "_")
	return "_mcpInput_" + strings.ToLower(name)
}

// exportName capitalizes the first letter of a name for use as an exported struct field.
func exportName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// jsonType maps a Go type to the corresponding type for JSON deserialization in a struct.
func jsonType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int":
		return "int"
	case "float64":
		return "float64"
	case "bool":
		return "bool"
	case "time.Duration":
		// Duration is represented as a string in JSON and parsed in the handler.
		return "string"
	default:
		return "string"
	}
}
