# Development

## Build instructions

### Prerequisites

- [go 1.19 or above](https://golang.org/doc/install)
- [goreleaser](https://goreleaser.com/install/) (optional)
- [golangci-lint](https://github.com/golangci/golangci-lint) (optional)

### Building

```bash
# Build sake for your platform target
make build

# Build sake binaries and archives for all platforms using goreleaser
make build-all

# Generate Manpage
make gen-man
```

## Developing

```bash
# Format code
make gofmt

# Manage dependencies (download/remove unused)
make tidy

# Lint code
make lint

# Standing in examples directory you can run the following to debug faster
go run ../main.go run ping -a
```

## Releasing

The following workflow is used for releasing a new `sake` version:

1. Create pull request with changes
2. Verify build works (especially windows build)
   - `make build`
   - `make build-all`
3. Pass all integration and unit tests locally
   - `make integration-test`
   - `make unit-test`
4. Run benchmarks and profiler to check performance
   - `make benchmark`
5. Update `config-reference.md` and `config.man` if any config changes and generate manpage
   - `make gen-man`
6. Update `Makefile` and `CHANGELOG.md` with correct version, and add all changes to `CHANGELOG.md`
7. Squash-merge to main with `Release vx.y.z` and description of changes
8. Run `make release`, which will:
   - Create a git tag with release notes
   - Trigger a build in Github that builds cross-platform binaries and generates release notes of changes between current and previous tag

## Overview of How Sake Works

1. Parse & validate CLI arguments
2. Parse `sake` config files and create config, inventory, tasks, specs, and target states
3. Create clients for remote and local task execution for the selected hosts/tasks
4. Execute tasks on remote/local hosts
6. Disconnect from remote hosts
7. Print any output (results, reports, errors, etc.)

## Dependency Graph

Create SVG dependency graphs using graphviz and [goda](https://github.com/loov/goda).

```bash
goda graph "github.com/alajmo/sake/..." | dot -Tsvg -o res/graph.svg
goda graph "github.com/alajmo/sake:all" | dot -Tsvg -o res/graph-full.svg
```
