# Installation

`sake` is available on Linux and Mac:

* Binaries are available on the [release](https://github.com/alajmo/sake/releases) page

* via cURL
  ```bash
  curl -sfL https://raw.githubusercontent.com/alajmo/sake/main/install.sh | sh
  ```

* via Homebrew
  ```bash
  brew tap alajmo/sake
  brew install sake
  ```

* via Go
  ```bash
  go get -u github.com/alajmo/sake
  ```

## Building From Source

Requires [go 1.18 or above](https://golang.org/doc/install).

1. Clone the repo
2. Build and run the executable

    ```bash
    make build && ./dist/sake
    ```
