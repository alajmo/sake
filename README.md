[![Build Status](https://github.com/alajmo/yac/workflows/test/badge.svg)](https://github.com/alajmo/yac/actions)
[![Release](https://img.shields.io/github/release-pre/alajmo/yac.svg)](https://github.com/alajmo/yac/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](https://img.shields.io/badge/license-MIT-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/alajmo/yac)](https://goreportcard.com/report/github.com/alajmo/yac)

# Yac

<img src="./res/logo-1.png" align="right"/>

`yac` is a CLI tool that helps you manage multiple repositories. It's useful when you are working with microservices, multi-project systems, many libraries or just a bunch of repositories and want a central place for pulling all repositories and running commands over them.

You specify repository and commands in a config file and then run the commands over all or a subset of the repositories.

![demo](res/output.gif)

## Features

- Clone multiple repositories in one command
- Run custom or ad-hoc commands over multiple repositories
- Flexible filtering
- Declarative configuration
- Portable, no dependencies
- Supports auto-completion

## Table of Contents

* [Installation](#installation)
  * [Building From Source](#building-from-source)
* [Usage](#usage)
  * [Create a New Yac Repository](#create-a-new-yac-repository)
  * [Common Commands](#common-commands)
  * [Documentation](#documentation)
* [License](#license)

## Installation

`yac` is available on Linux and Mac, with partial support for Windows.

* Binaries are available on the [release](https://github.com/alajmo/yac/releases) page

* via cURL (Linux & macOS)
  ```sh
  curl -sfL https://raw.githubusercontent.com/alajmo/yac/main/install.sh | sh
  ```

* Via GO install
    ```sh
    go get -u github.com/alajmo/yac
    ```

Auto-completion is available via `yac completion bash|zsh|fish|powershell`.

### Building From Source

1. Clone the repo
2. Build and run the executable
    ```sh
    make build && ./dist/yac
    ```

## Usage

### Create a New Yac Repository

Run the following command inside a directory containing your `git` repositories, to initialize a yac repo:

```sh
$ yac init
```

This will generate two files:

- `yac.yaml`: contains projects and custom tasks. Any sub-directory that has a `.git` inside it will be included (add the flag `--auto-discovery=false` to turn off this feature)
- `.gitignore`: includes the projects specified in `yac.yaml` file

It can be helpful to initialize the `yac` repository as a git repository so that anyone can easily download the `yac` repository and run `yac sync` to clone all repositories and get the same project setup as you.

### Common Commands

```sh
# Run arbitrary command (list all files for instance)
yac exec --all-projects 'ls -alh'

# List all repositories
yac list projects

# List repositories in a tree-like format
yac tree

# Describe available tasks
yac describe tasks

# Run task for projects that have the frontend tag
yac run list-files -t frontend

# Run task for projects under a specific directory
yac run list-files -d work/

# Run task for specific project
yac run list-files -p project-a

# Open up yac.yaml in your preferred editor
yac edit
```
### Documentation

Checkout the following to learn more about yac:

- [Examples](_example)
- [Documentation](docs/DOCUMENTATION.md)
- [List of Useful Git Commands](docs/COMMANDS.md)
- [Project Background](docs/PROJECT-BACKGROUND.md)

## [License](LICENSE)

The MIT License (MIT)

Copyright (c) 2020-2021 Samir Alajmovic
