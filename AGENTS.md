## Project Development Guidelines

- Stick to the standard Go tools when possible. For example, don't introduce Makefile's when a regular `go build` will suffice.

## Build & Test

```bash
go build          # Build the binary
go test ./...     # Run all tests
golangci-lint run # Run linter
```

Tests use a mock i3 client and do not require a running i3 instance.

## Project Structure

- `main.go` - Entry point, creates and runs the MCP server
- `server.go` - MCP server implementation with tool handlers (GetTree, GetWorkspaces, FindWindows, RunCommand)
- `i3client.go` - `I3Client` interface for i3 IPC, enables dependency injection for testing
- `server_test.go` - Unit tests using `mockI3Client`

## Key Dependencies

- `github.com/modelcontextprotocol/go-sdk` - MCP protocol implementation
- `go.i3wm.org/i3/v4` - i3 window manager IPC library

## CI

PRs run:
- `golangci-lint` for linting
- Tests on Go 1.24 and 1.25
