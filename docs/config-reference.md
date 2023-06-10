# Config Reference

The sake.yaml config is based on the following concepts:

- **servers** are servers, local or remote, that have a host
- **tasks** are shell commands that you write and then run for selected **servers**
- **specs** are configs that alter **task** execution and output
- **targets** are configs that provide shorthand filtering of **servers** when executing **tasks**
- **themes** are used to modify the output of `sake` commands
- **env** are environment variables that can be defined globally, per server and per task

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
# known_hosts_file: $HOME/.ssh/known_hosts

# Set timeout for ssh connections in seconds
# default_timeout: 20

# Shell used for commands [optional]
# If you use any other program than bash, zsh, sh, node, or python
# then you have to provide the command flag if you want the command-line string evaluted
# For instance: bash -c
shell: bash

# List of Servers
servers:
 # Server name [required]
 media:
   # Server description [optional]
   desc: media server

   # Host [required]
   host: media.lan
   # one-line for setting user and port
   # host: samir@media.lan:22

   # Specify multiple hosts:
   # hosts:
   # - samir@192.168.0.1:22
   # - samir@l92.168.1.1:22

   # or use a host range generator
   # hosts: samir@192.168.[0:1].1:22

   # generate hosts by local command
   # inventory: echo samir@192.168.0.1:22 samir@192.168.1.1:22

   # Bastion [optional]
   bastion: samir@192.168.1.1:2222

   # Bastions [optional]
   # bastions: [samir@192.168.1.1:2222, samir@192.168.1.2:3333]

   # User to connect as. It defaults to the current user [optional]
   user: samir

   # Port for ssh [optional]
   port: 22

   # Shell used for commands [optional]
   shell: bash

   # Run on localhost [optional]
   local: false

   # Set default working directory for task execution [optional]
   work_dir: ""

   # Set identity file. By default it will attempt to establish a connection using a SSH auth agent [optional]
   # sake respects users ssh config, so you can set auth credentials in the users ssh config
   identity_file: ./id_rsa

   # Set password. Accepts either a string or a shell command [optional]
   password: $(echo $MY_SECRET_PASSWORD)

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
     # Set host prefix for each line [optional]
     # Available variables: `.Name`, `.Index`, `.Host`, `.Port`, `.User`
     prefix: '{{ .Host }}'

     # Colors to alternate between for each server prefix [optional]
     # Available options: green, blue, red, yellow, magenta, cyan
     prefix_colors: ["green", "blue", "red", "yellow", "magenta", "cyan"]

     # Customize the task header that is printed before each task when output is set to text (to opt out, set it to empty string) [optional]
     # Available variables: `.Name`, `.Desc`, `.Index`, `.NumTasks`
     # Available methods: `.Style`, which takes in 1 or more parameters, first is the string to be styled, and the rest are styling options
     # Available styling options:
     #   Colors (prefix with `fg_` for foreground, and `bg_` for background): black, red, green, yellow, blue, magenta, cyan, white, hi_black, hi_red, hi_green, hi_yellow, hi_blue, hi_magenta, hi_cyan, hi_white
     #   Attributes: normal, bold, faint, italic, underline crossed_out
     header: '{{ .Style "TASK" "bold" }}{{ if ne .NumTasks 1 }} ({{ .Index }}/{{ .NumTasks }}){{end}}{{ if and .Name .Desc }} [{{.Style .Name "bold"}}: {{ .Desc }}] {{ else if .Name }} [{{ .Name }}] {{ else if .Desc }} [{{ .Desc }}] {{end}}'

     # Fill remaining spaces with a character after the header, if set to empty string, no filler characters will be displayed [optional]
     header_filler: "*"

   # Table options [optional]
   table:
     # Table style [optional]
     # Available options: ascii, connected-light
     style: ascii

     # Set host prefix [optional]
     # Available variables: `.Name`, `.Index`, `.Host`, `.Port`, `.User`
     prefix: '{{ .Host }}'

     # Border options for table output [optional]
     options:
       draw_border: false
       separate_columns: true
       separate_header: true
       separate_rows: false
       separate_footer: false

     # Color, attr, align, and format options [optional]
     # Available options for fg/bg: green, blue, red, yellow, magenta, cyan, hi_green, hi_blue, hi_red, hi_yellow, hi_magenta, hi_cyan
     # Available options for align: left, center, justify, right
     # Available options for attr: normal, bold, faint, italic, underline, crossed_out
     # Available options for format: default, lower, title, upper
     title:
       fg:
       bg:
       align:
       attr:
       format:

     header:
       fg:
       bg:
       align:
       attr:
       format:

     row:
       fg:
       bg:
       align:
       attr:
       format:

     footer:
       fg:
       bg:
       align:
       attr:
       format:

     border:
       header:
         fg:
         bg:
         attr:

       row:
         fg:
         bg:
         attr:

       row_alt:
         fg:
         bg:
         attr:

       footer:
         fg:
         bg:
         attr:

# List of Specs [optional]
specs:
 default:
   # Spec description
   desc: default spec

   # Print task description
   describe: false

   # Print list of hosts that will be targetted
   list_hosts: false

   # Order hosts [inventory|reverse_inventory|sorted|reverse_sorted|random]
   order: inventory

   # Omit showing loader when running tasks
   silent: false

   # Execution strategy [linear|host_pinned|free]
   strategy: linear

   # Number of hosts to run in parallel
   batch: 1

   # Number of hosts in percentage to run in parallel [0-100]
   # batch_p: 100

   # Max number of forks
   forks: 10000

   # Set task output [text|table|table-2|table-3|table-4|html|markdown|json|csv|none]
   output: text

   # Limit output [stdout|stderr|all]
   print: all

   # Hide task from auto-completion
   hidden: false

   # Continue task execution on errors
   ignore_errors: true

   # Stop task execution on any error
   any_errors_fatal: false

   # Max number of tasks to fail before aborting
   max_fail_percentage: 100

   # Ignore unreachable hosts
   ignore_unreachable: false

   # Omit empty rows for table output
   omit_empty_rows: false

   # Omit empty columns for table output
   omit_empty_columns: false

   # Show task reports [recap|rc|task|time|all]
   report: [recap]

   # Verbose turns on describe, list_hosts and report set to all
   verbose: false

   # Confirm invoked task before running
   confirm: false

   # Confirm each task before running
   step: false

# List of targets [optional]
targets:
 default:
   # Target description
   desc: ""

   # Target all hosts
   all: false

   # Specify hosts via server name
   servers: []

   # Specify hosts via server tags
   tags: []

   # Limit number of hosts to target
   limit: 0

   # Limit number of hosts to target in percentage
   limit_p: 100

   # Invert matching on hosts
   invert: false

   # Specify host regex
   regex: ""

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
     ignore_errors: true
     ignore_unreachable: true
     any_errors_fatal: false
     omit_empty_rows: true
     omit_empty_columns: true

   # Target reference [optional]
   # target: default

   # Or specify targets inline
   target:
     all: true
     servers: [media]
     tags: [remote]
     limit: 1

   # List of environment variables [optional]
   env:
     # Simple string value
     release: v1.0.0

     # Shell command substitution
     num_lines: $(ls -1 | wc -l)

     # The following variables are available by default:
     #   S_NAME
     #   S_HOST
     #   S_USER
     #   S_PORT
     #   S_BASTION
     #   S_TAGS
     #   S_IDENTITY
     #   SAKE_DIR
     #   SAKE_PATH

   # Run on localhost [optional]
   local: false

   # Set default working directory for task [optional]
   work_dir: ""

   # Shell used for commands [optional]
   shell: bash

   # Each task can only define:
   # - a single cmd
   # - or a single task reference
   # - or a list of task references and commands

   # Single command
   cmd: |
     echo complex
     echo command

   # Task reference. work_dir and env variables are passed down
   task: simple-1

   # List of task references or commands
   tasks:
     # Command
     - name: inline-command
       cmd: echo "Hello World"
       ignore_errors: true
       work_dir: /tmp
       shell: bash
       env:
         foo: bar

     # Task reference. work_dir and env variables are passed down.
     # Nested task referencing is supported and will result in a
     # flat list of commands
     - task: simple-1
       ignore_errors: true
       work_dir: /tmp
       register: results
       env:
         foo: bar

     - name: output
       cmd: echo $results_stdout
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

SAKE_SSH_CONFIG
    Override ssh config file path

SAKE_KNOWN_HOSTS_FILE
    Override known_hosts file path

NO_COLOR
    If this env variable is set (regardless of value) then all colors will be disabled
```
