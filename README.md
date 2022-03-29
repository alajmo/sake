[![Build Status](https://github.com/alajmo/sake/workflows/test/badge.svg)](https://github.com/alajmo/sake/actions)
[![Release](https://img.shields.io/github/release-pre/alajmo/sake.svg)](https://github.com/alajmo/sake/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](https://img.shields.io/badge/license-MIT-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/alajmo/sake)](https://goreportcard.com/report/github.com/alajmo/sake)

# Sake

`sake` is a CLI tool that enables you to run commands on servers via `ssh`. Think of it like `make`, you define servers and tasks in a declarative configuration file and then run the tasks on the servers.

It has many ergonomic features such as `auto-completion` of tasks, servers and tags. Additionally, it includes sub-commands to let you easily

- `ssh` into servers or docker containers
- list servers/tasks
- create tasks that queries server info and present it in a compact table format

Interested in managing your git repositiories in a similar way? Checkout [mani](https://github.com/alajmo/mani)!

## Table of Contents

- [Installation](#installation)
  - [Building From Source](#building-from-source)
- [Usage](#usage)
  - [Create a New Sake Config](#create-a-new-sake-config)
  - [Run Some Commands](#run-some-commands)
- [Documentation](#documentation)
- [License](#license)

## Installation

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

* Via GO install
    ```sh
    go get -u github.com/alajmo/sake
    ```

Auto-completion is available via `sake completion bash|zsh|fish` and man page via `sake gen`.

### Building From Source

Requires [go 1.18 or above](https://golang.org/doc/install).

1. Clone the repo
2. Build and run the executable
    ```sh
    make build && ./dist/sake
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
$ sake exec --all --output table --parallel 'find . -type f | wc -l'

 Server    | Output
-----------+--------
 localhost | 1
```

## Documentation

- [Examples](examples)
- [Recipes](docs/recipes.md)
- [Config Reference](docs/config-reference.md)
- [Command Reference](docs/command-reference.md)
- [Project Background](docs/project-background.md)
- [Changelog](docs/changelog.md)
- [sakecli.com](https://sakecli.com/)

## [License](LICENSE)

The MIT License (MIT)

Copyright (c) 2022 Samir Alajmovic
