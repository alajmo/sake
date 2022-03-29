# Config Reference

The sake.yaml config is based on the following concepts:

- **servers** are servers, local or remote, that have a host
- **tasks** are shell commands that you write and then run for selected **servers**
- **specs** are configs that alter **task** execution and output
- **targets** are configs that provide shorthand filtering of **servers** when executing **tasks**
- **themes** are used to modify the output of `sake` commands

**Specs**, **targets** and **themes** come with a default setting that the user can override.

Check the [files](#files) and [environment](#environment) section to see how the config file is loaded.

Below is a config file detailing all of the available options and their defaults.

```yaml
# Import servers/tasks/env/specs/themes/targets from other configs [optional]
import:
  - ./some-dir/sake.yaml

# Verify SSH host connections. Set this to true if you wish to circumvent verify host [optional]
disable_verify_host: false

# Set known_hosts_file path. Default is users ssh home directory [optional]
known_hosts_file: $HOME/.ssh/known_hosts

# List of Servers
servers:
  # Server name [required]
  media:
    # Server description [optional]
    desc: media server

    # Host [required]
    host: media.lan

    # User to connect as. It defaults to the current user [optional]
    user: whoami

    # Port for ssh [optional]
    port: 22

    # Set identity file. By default it will attempt to establish a connection using a SSH auth agent [optional]
    identity_file: ./id_rsa

    # Set password. Accepts either a string or a shell command [optional]
    password: $(echo $MY_SECRET_PASSWORD)

    # Run on localhost [optional]
    local: false

    # Set default working directory for task execution [optional]
    work_dir: ""

    # List of tags [optional]
    tags: [remote]

    # List of server specific environment variables [optional]
    env:
      # Simple string value
      key: value

      # Shell command substitution (evaluated on localhost)
      date: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

# List of environment variables that are available to all tasks
env:
  # Simple string value
  AUTHOR: "alajmo"

  # Shell command substitution (evaluated on localhost)
  DATE: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

# List of themes
themes:
  # Theme name
  default:
    # Text options [optional]
    text:
      # Include server name prefix for each line [optional]
      prefix: true

      # Colors to alternate between for each server prefix [optional]
      # Available options: green, blue, red, yellow, magenta, cyan
      prefix_colors: ["green", "blue", "red", "yellow", "magenta", "cyan"]

      # Add a header before each server [optional]
      header: true

      # String value that appears before the server name in the header [optional]
      header_prefix: "TASK"

      # Fill remaining spaces with a character after the prefix [optional]
      header_char: "*"

    # Table options [optional]
    table:
      # Table style [optional]
      # Available options: ascii, default
      style: ascii

      # Text format options for headers and rows in table output [optional]
      # Available options: default, lower, title, upper
      format:
        header: default
        row: default

      # Border options for table output [optional]
      options:
        draw_border: false
        separate_columns: true
        separate_header: true
        separate_rows: false
        separate_footer: false

      # Color, attr and align options [optional]
      # Available options for fg/bg: green, blue, red, yellow, magenta, cyan
      # Available options for align: left, center, justify, right
      # Available options for attr: normal, bold, faint, italic, underline, crossed_out
      color:
        header:
          server:
            fg:
            bg:
            align: left
            attr: normal

          user:
            fg:
            bg:
            align: left
            attr: normal

          host:
            fg:
            bg:
            align: left
            attr: normal

          port:
            fg:
            bg:
            align: left
            attr: normal

          local:
            fg:
            bg:
            align: left
            attr: normal

          tag:
            fg:
            bg:
            align: left
            attr: normal

          desc:
            fg:
            bg:
            align: left
            attr: normal

          task:
            fg:
            bg:
            align: left
            attr: normal

          output:
            fg:
            bg:
            align: left
            attr: normal

        row:
          server:
            fg:
            bg:
            align: left
            attr: normal

          user:
            fg:
            bg:
            align: left
            attr: normal

          host:
            fg:
            bg:
            align: left
            attr: normal

          port:
            fg:
            bg:
            align: left
            attr: normal

          local:
            fg:
            bg:
            align: left
            attr: normal

          tag:
            fg:
            bg:
            align: left
            attr: normal

          desc:
            fg:
            # bg:
            align: left
            attr: normal

          task:
            fg:
            # bg:
            align: left
            attr: normal

          output:
            fg:
            bg:
            align: left
            attr: normal

        border:
          header:
            fg:
            bg:

          row:
            fg:
            bg:

          row_alt:
            fg:
            bg:

          footer:
            fg:
            bg:


# List of Specs [optional]
specs:
  default:
    # Set task output [text|table|html|markdown]
    output: text

    # Run server tasks in parallel
    parallel: false

    # Continue task execution on errors
    ignore_errors: true

    # Stop task execution on all servers on error
    any_errors_fatal: false

    # Ignore unreachable hosts
    ignore_unreachable: false

    # Omit empty results for table output
    omit_empty: false

# List of targets [optional]
targets:
  default:
    # Target all servers
    all: false

    # Specify servers via server name
    servers: []

    # Specify servers via server tags
    tags: []

# List of tasks
tasks:
  # Command ID [required]
  simple-1:
    # The name that will be displayed when executing or listing tasks. Defaults to task ID [optional]
    name: Simple

    # Script to run
    cmd: |
      echo "hello world"
    desc: simple command 1

  # Short-form for a command
  simple-2: echo "hello world"

  # Command ID [required]
  advanced-command:
    # The name that will be displayed when executing or listing tasks. Defaults to task ID [optional]
    name: Advanced Command

    # Task description [optional]
    desc: Advanced task

    # Specify theme [optional]
    theme: default

    # Spec reference [optional]
    # spec: default

    # Or specify specs inline
    spec:
      output: table
      parallel: true
      ignore_errors: true
      ignore_unreachable: true
      any_errors_fatal: false
      omit_empty: true

    # Target reference [optional]
    # target: default

    # Or specify targets inline
    target:
      all: true
      servers: [media]
      tags: [remote]

    # List of environment variables [optional]
    env:
      # Simple string value
      release: v1.0.0

      # Shell command substitution
      num_lines: $(ls -1 | wc -l)

      # The following variables are available by default:
      #   SAKE_DIR
      #   SAKE_PATH
      #
      #   SAKE_TASK_ID
      #   SAKE_TASK_NAME
      #   SAKE_TASK_DESC
      #   SAKE_TASK_LOCAL
      #
      #   SAKE_SERVER_NAME
      #   SAKE_SERVER_DESC
      #   SAKE_SERVER_TAGS
      #   SAKE_SERVER_HOST
      #   SAKE_SERVER_USER
      #   SAKE_SERVER_PORT
      #   SAKE_SERVER_LOCAL

    # Run on localhost [optional]
    local: false

    # Set default working directory for task [optional]
    work_dir: ""

    # Each task can only define:
    # - a single cmd
    # - or a single task reference
    # - or a list of task references or commands

    # Single command
    cmd: |
      echo complex
      echo command

    # Task reference. work_dir and env variables are passed down.
    task: simple-1

    # List of task references or commands.
    tasks:
      # Command
      - name: inline-command
        cmd: echo "Hello World"
        work_dir: /tmp
        env:
          foo: bar

      # Task reference. work_dir and env variables are passed down.
      # Nested task referencing is supported and will result in a
      # flat list of commands
      - task: simple-1
        work_dir: /tmp
        env:
          foo: bar
```

## Files

When running a command, `sake` will check the current directory and all parent directories for the following files: `sake.yaml`, `sake.yml`, `.sake.yaml`, `.sake.yml` .

Additionally, it will import (if found) a config file from:

- Linux: `$XDG_CONFIG_HOME/sake/config.yaml` or `$HOME/.config/sake/config.yaml` if `$XDG_CONFIG_HOME` is not set.
- Darwin: `$HOME/Library/Application/sake`

Both the config and user config can be specified via flags or environments variables.

## Environment

```txt
SAKE_CONFIG
    Override config file path

SAKE_USER_CONFIG
    Override user config file path

SAKE_KNOWN_HOSTS_FILE
    Override known_hosts file path

SAKE_IDENTITY_FILE
    Override identity file path

SAKE_PASSWORD
    Override SSH password

NO_COLOR
    If this env variable is set (regardless of value) then all colors will be disabled
```
