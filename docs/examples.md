# Examples

This is an example of how you can use `sake` for localhost. If you wish to run commands against remote servers via `ssh`, add your own remote servers.

## Simple Example

```yaml title=sake.yaml
servers:
  localhost:
    desc: localhost
    host: localhost
    local: true
    tags: [local]

tasks:
  ping:
    desc: ping server
    target:
      all: true
    cmd: echo pong

  print-host:
    name: Host
    desc: print host
    spec:
      output: table
    target:
      all: true
    cmd: echo $SAKE_SERVER_HOST

  info:
    desc: get remote info
    target:
      tags: [local]
    spec:
      output: table
      ignore_errors: true

    tasks:
      - task: ping
      - task: print-host
      - cmd: echo "Done"
```

Now let's run some commands:

```bash
# List all servers
$ sake list servers

 Server    | Host      | Tag   | Description
-----------+-----------+-------+-------------
 localhost | localhost | local | localhost

# Describe task
$ sake describe task print-host

Task: print-host
Name: Host
Desc: print host
Local: false
Theme: default
Target:
    All: true
    Servers:
    Tags:
Spec:
    Output: table
    Parallel: false
    AnyErrorsFatal: false
    IgnoreErrors: false
    IgnoreUnreachable: false
    OmitEmpty: false
Env:
    SAKE_TASK_ID: print-host
    SAKE_TASK_NAME: Host
    SAKE_TASK_DESC: print host
    SAKE_TASK_LOCAL: false
Cmd:
    echo $SAKE_HOST


# Run a task targeting servers with tag `local`
$ sake run ping --tags local

TASK ping: ping server ****************

localhost | pong

# Run task that has multiple commands
$ sake run info --all

 Server    | Ping | Host      | Output
-----------+------+-----------+--------
 localhost | pong | localhost | Done

# Same task but with text output
$ sake run info --all --output text

TASK (1/3) ping: ping server **********

localhost | pong

TASK (2/3) Host: print host ***********

localhost | localhost

TASK (3/3) Command ********************

localhost | Done

# Run runtime defined command for all servers
$ sake exec --all --output table --parallel 'cd ~ && ls -al | wc -l'

 Server    | Output
-----------+--------
 localhost | 42
```

## Advanced Example

Create the following files:

```bash
.
├── sake.yaml
├── servers.yaml
└── tasks.yaml
```

```yaml title=sake.yaml
import:
  - servers.yaml
  - tasks.yaml

env:
  DATE: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

tasks:
  ping:
    desc: ping server
    spec: info
    target: all
    cmd: echo pong

  overview:
    desc: get system overview
    spec: info
    target: all
    tasks:
      - name: date
        cmd: echo $DATE

      - name: pwd
        cmd: pwd

      - task: info
      - task: print-uptime
```

```yaml title=servers.yaml
servers:
  localhost:
    desc: localhost
    host: localhost
    local: true
    work_dir: /tmp
    tags: [local]
```

```yaml title=tasks.yaml
specs:
  info:
    output: table
    ignore_errors: true
    omit_empty: true
    any_fatal_errors: false
    ignore_unreachable: true
    parallel: true

targets:
  all:
    all: true
    tags: []
    servers: []

tasks:
  print-uptime:
    name: Uptime
    desc: print uptime
    spec: info
    target: all
    cmd: uptime | grep -E -o "[0-9]* (day|days)"

  print-host:
    name: Host
    desc: print host
    target: all
    cmd: echo $SAKE_SERVER_HOST

  print-hostname:
    name: Hostname
    desc: print hostname
    spec: info
    target: all
    cmd: hostname

  print-os:
    name: OS
    desc: print OS
    spec: info
    target: all
    cmd: |
      os=$(lsb_release -si)
      release=$(lsb_release -sr)
      echo "$os $release"

  print-kernel:
    name: Kernel
    desc: Print kernel version
    spec: info
    target: all
    cmd: uname -r | awk -v FS='-' '{print $1}'

  info:
    name: Info
    desc: Print system info
    tasks:
      - task: print-host
      - task: print-hostname
      - task: print-os
      - task: print-kernel
```

And run `sake run overview`:

```bash
 Server    | Date                   | Pwd  | Host      | Hostname | OS             | Kernel | Uptime
-----------+------------------------+------+-----------+----------+----------------+--------+--------
 localhost | 2022-06-09T10:02:10UTC | /tmp | localhost | HAL-9000 | Debian testing | 5.16.0 | 9 days
```

## Docker Compose Example

