<h1 align="center"><code>sake</code></h1>

<div align="center">
  <a href="https://github.com/alajmo/sake/releases">
    <img src="https://img.shields.io/github/release-pre/alajmo/sake.svg" alt="version">
  </a>

  <a href="https://github.com/alajmo/sake/actions">
    <img src="https://github.com/alajmo/sake/workflows/build/badge.svg" alt="build status">
  </a>

  <a href="https://img.shields.io/badge/license-MIT-green">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="license">
  </a>

  <a href="https://goreportcard.com/report/github.com/alajmo/sake">
    <img src="https://goreportcard.com/badge/github.com/alajmo/sake" alt="Go Report Card">
  </a>

  <a href="https://pkg.go.dev/github.com/alajmo/sake">
    <img src="https://pkg.go.dev/badge/github.com/alajmo/sake.svg" alt="reference">
  </a>
</div>

<br>

`sake` is a command runner for local and remote hosts. You define servers and tasks in `sake.yaml` file and then run the tasks on the servers.

This readme is also accessible on [sakecli.com](https://sakecli.com/).

`sake` has tons of features:

- auto-completion of tasks, servers and tags
- SSH into servers or docker containers `sake ssh <server>`
- list servers/tasks via `sake list servers|tasks`
- present task output in a compact table format `sake run <task> --output table`
- open task/server in your preferred editor `sake edit task <task>`
- import other `sake.yaml` configs
- and [many more!](docs/recipes.md)

![demo](res/output.gif)

Interested in managing your git repositories in a similar way? Check out [mani](https://github.com/alajmo/mani)!

## Table of Contents

- [Installation](#installation)
  - [Building From Source](#building-from-source)
- [Usage](#usage)
  - [Create a New Sake Config](#create-a-new-sake-config)
  - [Run Some Commands](#run-some-commands)
- [Documentation](#documentation)
- [License](#license)

## Installation

[![Packaging status](https://repology.org/badge/vertical-allrepos/sake.svg)](https://repology.org/project/sake/versions)

`sake` is available on Linux and Mac.

* Binaries are available on the [release](https://github.com/alajmo/sake/releases) page

* via cURL
  ```sh
  curl -sfL https://raw.githubusercontent.com/alajmo/sake/main/install.sh | sh
  ```

* via Homebrew
  ```sh
  brew tap alajmo/sake
  brew install sake
  ```

* via MacPorts
  ```sh
  sudo port install sake
  ```

* via Arch
  ```sh
  pacman -S sake
  ```

* via pkg
  ```sh
  pkg install sake
  ```

* Via Go
    ```sh
    go install github.com/alajmo/sake@latest
    ```

Auto-completion is available via `sake completion bash|zsh|fish` and man page via `sake gen`.

### Building From Source

Requires [go 1.19 or above](https://golang.org/doc/install).

1. Clone the repo
2. Build and run the executable
    ```sh
    make build && ./dist/sake

    # To build for all target platforms run (requires goreleaser CLI)
    make build-all
    ```

## Usage

### Create a New Sake Config

Run the following command:

```bash
$ sake init

Initialized sake in /tmp/sake
- Created sake.yaml

Following servers were added to sake.yaml

 Server    | Host
-----------+---------
 localhost | 0.0.0.0
```

### Run Some Commands

```bash
# List all servers
$ sake list servers

 Server    | Host
-----------+---------
 localhost | 0.0.0.0

# List all tasks
$ sake list tasks

 Task | Description
------+-------------
 ping | Pong

# Run Task
$ sake run ping --all

TASK ping: Pong ************

0.0.0.0 | pong

# Count number of files in each server in parallel
$ sake exec --all --output table --strategy=free 'find . -type f | wc -l'

 Server    | Output
-----------+--------
 localhost | 1
```

### What's Next

Check out the [examples page](/docs/examples.md) for more advanced examples and the [recipes page](/docs/recipes.md) for a list of useful recipes.

## Documentation

- [Examples](docs/examples.md)
- [Recipes](docs/recipes.md)
- [Config Reference](docs/config-reference.md)
- [Command Reference](docs/command-reference.md)
- Documentation
  - [Inventory](docs/inventory.md)
  - [Task Execution](docs/task-execution.md)
  - [Error Handling](docs/error-handling.md)
  - [Variables](docs/variables.md)
  - [Working Directory](docs/work-dir.md)
  - [Output](docs/output.md)
- Project
  - [Background](docs/background.md)
  - [Roadmap](docs/roadmap.md)
  - [Ansible](docs/ansible.md)
  - [Performance](docs/performance.md)
- Development
  - [Development](docs/development.md)
  - [Contributing](docs/contributing.md)
- [Changelog](docs/changelog.md)

## [License](LICENSE)

The MIT License (MIT)

Copyright (c) 2022 Samir Alajmovic
