# Copilot Instructions for Mage

Mage is a make-like build tool that uses Go functions as build targets. Users write plain Go functions in "magefiles" and mage makes them runnable from the command line.

## Build, Test, and Lint

```bash
# Build
go build ./...

# Test (full suite, including race detector as required by CI)
go test -race ./...

# Test a single package
go test -race ./parse/

# Test a single test function
go test -race ./mage/ -run TestGoCmd

# CI runs tests with: go test -v -vet=all -tags CI -race ./...

# Lint (requires golangci-lint)
golangci-lint run ./...
```

Mage builds itself with mage. The bootstrap path (`go run bootstrap.go`) is for building mage when mage isn't installed yet. The project's own build targets live in `magefiles/`.

## Architecture

Mage works by **parsing user Go source files and generating a temporary CLI binary** that dispatches to the user's target functions:

1. **Entry** — `main.go` calls `mage.Main()` which parses CLI flags and dispatches commands.
2. **File scanning** — `mage/main.go` finds magefiles (files with `//go:build mage` or `// +build mage` tags) in the current directory.
3. **AST parsing** — `parse.PrimaryPackage()` uses `go/parser` and `go/doc` to extract exported functions, namespaces (types embedding `mg.Namespace`), `//mage:import` directives, aliases, and the default target.
4. **Code generation** — `GenerateMainfile()` renders `mage/magefile_tmpl.go` (a Go `text/template`) into a wrapper `main` package that handles flag parsing, help output, and dispatching to user targets.
5. **Compilation & caching** — `Compile()` runs `go build` on the magefiles plus the generated wrapper. The output binary is cached by content hash in the user's cache directory.

### Package Map

- **`mage/`** — Core library: CLI entry point, file scanning, code generation, compilation, and execution. Can be used as a library (`mage.Invoke()`).
- **`mg/`** — User-facing API for magefiles: `Deps`/`CtxDeps` for dependency declaration, `mg.F()` for parameterized targets, `mg.Namespace` for grouping targets, `Fatal`/`Fatalf` for error handling.
- **`parse/`** — Go AST parser that extracts target metadata (functions, namespaces, imports, aliases, defaults) from magefiles into a `parse.PkgInfo` model consumed by code generation.
- **`sh/`** — Shell helper functions (`sh.Run`, `sh.Output`, `sh.Exec`) for use in magefiles.
- **`internal/`** — Shared low-level utilities for command execution and debug output.
- **`target/`** — Timestamp-based rebuild helpers (`target.Path`, `target.Dir`, `target.Glob`) for use in magefiles.

## Key Conventions

- **Zero external dependencies.** Mage uses only the Go standard library. This is intentional — since mage is often vendored into projects, adding dependencies to mage adds them to every project that uses it. Do not add external dependencies.
- **Go 1.18 minimum.** The `go.mod` specifies Go 1.18. CI tests against both Go 1.18 and stable. Avoid language features or stdlib APIs from newer Go versions.
- **Target function signatures** follow strict rules enforced by `parse/parse.go` (`funcType`): optional leading `context.Context` parameter, supported arg types (`string`, `int`, `bool`, `time.Duration`), and must return either nothing or `error`. Pointer args become optional CLI arguments.
- **`//mage:import`** comments on blank imports cause mage to recursively parse imported packages and surface their exported functions as targets.
- **Namespace targets** are methods on types that embed `mg.Namespace`. The type name becomes a CLI prefix (e.g., `mage ns:target`).
- **Formatting** uses `goimports` (configured in `.golangci.toml`).
- **Tests** are primarily integration-style: `mage/main_test.go` calls `Invoke()` against fixture directories under `testdata/`. Table-driven unit tests are used in `parse/`, `sh/`, `internal/`, and `target/`. Always run tests with `-race`.
