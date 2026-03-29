package mcp

var mcpMainfileTplString = `//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	_fmt "fmt"
	_log "log"
	"os"
	"time"

	_mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	{{range .Imports}}{{.UniqueName}} "{{.Path}}"
	{{end}}
)

// Suppress unused import errors.
var (
	_ = _fmt.Sprintf
	_ = time.Second
	_ _log.Logger
	_ bytes.Buffer
)

{{range .Funcs}}{{mcpStructDef .}}
{{end}}
{{range .Imports}}{{range .Info.Funcs}}{{mcpStructDef .}}
{{end}}{{end}}

func main() {
	server := _mcp.NewServer(&_mcp.Implementation{
		Name: "{{.BinaryName}}",
	}, nil)

	{{range .Funcs}}
	{{mcpAddTool .}}
	{{end}}

	{{range .Imports}}{{range .Info.Funcs}}
	{{mcpAddTool .}}
	{{end}}{{end}}

	if err := server.Run(context.Background(), &_mcp.StdioTransport{}); err != nil {
		_log.Fatal(err)
	}
}
`

