# Agent Guidelines for go-go-mcp

## Build Commands
- Build: `go build ./...`
- Test: `go test ./...`
- Run single test: `go test ./pkg/path/to/package -run TestName`
- Generate: `go generate ./...`
- Lint: `golangci-lint run -v` or `make lint`
- Format: `go fmt ./...`

## Code Style Guidelines
- **Imports**: Standard library first, third-party next, local packages last
- **Naming**: CamelCase for exported, camelCase for unexported
- **Error Handling**: Return errors with context (pkg/errors), use zerolog for logging
- **Structure**: Follow Go project layout conventions
- **Types**: Define clear interfaces, use struct tags for serialization
- **Commands**: Use Cobra for CLI commands

## Project Structure
- `cmd/`: CLI commands and entry points
- `pkg/`: Library code organized by domain
- `examples/`: Example configurations and usage
- `doc/`: Documentation

<goGuidelines>
When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
When building web applications, use htmx, bootstrap and the templ templating language.
Always use a context argument when appropriate.
Use cobra for command-line applications.
Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
Use github.com/pkg/errors for wrapping errors.
When starting goroutines, use errgroup.
</goGuidelines>

<webGuidelines>
Use bun, react and rtk-query. Use typescript.
Use bootstrap for styling.
</webGuidelines>