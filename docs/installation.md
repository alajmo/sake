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

* via Go
  ```bash
  go install github.com/alajmo/sake@latest
  ```

## Building From Source

Requires [go 1.19 or above](https://golang.org/doc/install).

1. Clone the repo
2. Build and run the executable

    ```bash
    make build && ./dist/sake
    ```
