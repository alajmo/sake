# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sake is a task runner for local and remote hosts written in Go. Users define servers and tasks in `sake.yaml` files and execute tasks on servers via SSH or locally.

## Common Commands

```bash
# Build
make build              # Build binary to dist/sake
make build-all          # Cross-platform builds (requires goreleaser)

# Test
make test               # Run all tests (requires docker-compose for mock SSH servers)
make test-unit          # Run unit tests only
make test-integration   # Run integration tests only (requires mock-ssh running)
make mock-ssh           # Start mock SSH servers for integration tests
make update-golden-files # Update golden test files

# Code quality
make lint               # Run golangci-lint and deadcode checker
make gofmt              # Format Go code

# Development
go run ../main.go run ping -a  # Quick debug from examples directory
```

## Architecture

### Entry Points
- `main.go` → calls `cmd.Execute()`
- `cmd/root.go` → Cobra CLI setup and command registration

### Core Packages

**cmd/** - CLI command handlers using Cobra framework. Each command in its own file (run.go, exec.go, ssh.go, list.go, etc.)

**core/dao/** - Data Access Objects for config parsing:
- `config.go` - Main Config struct, YAML parsing, imports
- `server.go` - Server definitions with SSH/local connection details
- `task.go` - Task commands and subtasks (TaskCmd, TaskRef)
- `spec.go` - Execution specs (strategy, batch, forks, output format)
- `target.go` - Server filtering (by name, tags, regex, limits)
- `theme.go` - Output formatting themes

**core/run/** - Task execution engine:
- `exec.go` - Main orchestrator, handles execution strategies
- `ssh.go` - SSH client wrapper (key auth, password auth, agent)
- `localhost.go` - Local execution via exec.Cmd
- `client.go` - Client interface definition

**core/print/** - Output formatting (table, text, JSON, CSV, HTML, Markdown)

### Data Flow
```
CLI Input → cmd/root.go → core/dao/config.go (parse YAML)
→ Create Config (Servers, Tasks, Specs, Targets, Themes)
→ core/run/exec.go (create SSH/Local clients, execute with strategy)
→ core/print/ (format output)
```

### Key Patterns

**YAML Struct Conversion**: Separate `*YAML` structs for unmarshaling that convert to domain structs after validation (e.g., `ServerYAML` → `Server`)

**Execution Strategies**: linear (sequential), host_pinned (serial per host), free (concurrent)

**Platform-Specific Code**: `unix.go` and `windows.go` for OS-specific handling

## Testing

Integration tests require mock SSH servers running via Docker:
```bash
make mock-ssh           # Terminal 1: start mock servers
make test-integration   # Terminal 2: run tests
```

Golden files in test/integration/ validate output. Update with `make update-golden-files`.
