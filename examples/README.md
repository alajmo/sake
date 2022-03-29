# Examples

This is an example of how you can use `sake` for localhost. If you wish to run commands against remote servers via `ssh`, add your own remote servers.

### Config

```yaml
servers:
  localhost:
    desc: localhost
    host: localhost
    local: true
    tags: [local]

# GLOBAL ENVS
env:
  DATE: $(date -u +"%Y-%m-%dT%H:%M:%S%Z")

targets:
  all:
    all: true

specs:
  table:
    output: table

tasks:
  ping:
    desc: ping server
    target: all
    cmd: echo pong

  print-host:
    name: Host
    desc: print host
    spec: table
    target: all
    cmd: echo $SAKE_SERVER_HOST

  info:
    desc: get remote info
    target:
      tags: [remote]
    spec:
      output: table
      ignore_errors: true

    tasks:
      - task: ping
      - task: print-host
      - cmd: echo "Done"
```

## Commands

### List All Servers

```bash
$ sake list servers

 Server    | Host      | Tag   | Description
-----------+-----------+-------+-------------
 localhost | localhost | local | localhost
```

### Describe Task

```bash
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
```

### Run a Task Targeting Servers With Tag `Remote`

```bash
$ sake run ping --tags local

TASK ping: ping server ****************

localhost | pong
```

### Run Task That Has Multiple Commands

```bash
$ sake run info --all

 Server    | Ping | Host      | Output
-----------+------+-----------+--------
 localhost | pong | localhost | Done

$ sake run info --all --output text

TASK (1/3) ping: ping server **********

localhost | pong

TASK (2/3) Host: print host ***********

localhost | localhost

TASK (3/3) Command ********************

localhost | Done
```

### Run Runtime Defined Command for All Servers

```bash
$ sake exec --all --output table --parallel 'cd ~ && ls -al | wc -l'

 Server    | Output
-----------+--------
 localhost | 42
```

