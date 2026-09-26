# claude-statusline

Go status line for Claude Code, with a `--hook` mode that tracks running subagents.

## Code exploration
- Go is covered by `gopls`: load the LSP tool (`ToolSearch("select:LSP")`) before touching code.
- Locate symbols with `workspaceSymbol`, `goToDefinition`, `findReferences`, `hover` before any Read/rg on `.go` files, then read only the relevant lines.

## Checks
- `go vet ./... && go test ./...`
- `make install` rebuilds and installs `~/.local/bin/claude-statusline`.
