package mcp

import (
"fmt"
"github.com/magefile/mage/parse"
)

func main() {
// Test optional duration arg
f := &parse.Function{
Name:    "Wait",
IsError: true,
Args: []parse.Arg{
{Name: "duration", Type: "time.Duration", Optional: true},
},
Synopsis: "Wait for a duration",
}

fmt.Println("=== Optional Duration Arg ===")
fmt.Println(mcpAddTool(f))

// Test required and optional duration args together
f2 := &parse.Function{
Name:    "SleepAndWait",
IsError: true,
Args: []parse.Arg{
{Name: "sleep", Type: "time.Duration", Optional: false},
{Name: "wait", Type: "time.Duration", Optional: true},
},
Synopsis: "Sleep and wait",
}

fmt.Println("\n=== Mixed Duration Args ===")
fmt.Println(mcpAddTool(f2))
}
