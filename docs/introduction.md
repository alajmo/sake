---
slug: /
---

# Introduction

`sake` is a command runner for local and remote hosts. You define servers and tasks in a `sake.yaml` config file and then run the tasks on the servers.

`sake` has tons of features:

- auto-completion of tasks, servers and tags
- SSH into servers or docker containers `sake ssh <server>`
- list servers/tasks via `sake list servers|tasks`
- present task output in a compact table format `sake run <task> --output table`
- open task/server in your preferred editor `sake edit task <task>`
- import other `sake.yaml` configs
- and [many more!](recipes.md)

## Demo

![demo](/img/output.gif)

## Example

You specify servers and tasks in a config file:

```yaml title=sake.yaml
servers:
  localhost:
    desc: my workstation
    host: localhost
    local: true
    tags: [local]

  server-1:
    desc: hosts mealie, Node-RED
    host: server-1.lan
    user: samir
    tags: [remote,pi]

  pihole:
    desc: runs pihole
    host: pihole.lan
    user: samir
    tags: [remote,pi]

tasks:
  ping:
    desc: ping server
    cmd: echo pong

  print-host:
    name: Host
    desc: print host
    cmd: echo $S_HOST

  print-os:
    name: OS
    desc: print OS
    cmd: |
      os=$(lsb_release -si)
      release=$(lsb_release -sr)
      echo "$os $release"

  print-kernel:
    name: Kernel
    desc: Print kernel version
    cmd: uname -r | awk -v FS='-' '{print $1}'

  info:
    desc: get remote info
    target:
      tags: [remote]
    spec:
      output: table
      strategy: free
      ignore_errors: true
      ignore_unreachable: true
    tasks:
      - task: print-os
      - task: print-kernel
```

and then run the tasks over all or a subset of the repositories:

```bash
# Simple ping command
$ sake run ping --tags remote

TASK [ping: ping server] **************

server-1.lan | pong

TASK [ping: ping server] **************

pihole.lan | pong

# Multiple tasks
$ sake run info --all

 Server   | OS        | Kernel
----------+-----------+---------
 server-1 | Debian 11 | 5.10.92
 pihole   | Debian 11 | 5.10.92

# Runtime defined command
sake exec 'sudo apt install rsync' --tags remote

TASK Command ******************************************************************

server-1.lan | Reading package lists...
server-1.lan | Building dependency tree...
server-1.lan | Reading state information...
server-1.lan | rsync is already the newest version (3.2.3-4+deb11u1).
server-1.lan | 0 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.

TASK Command ******************************************************************

pihole.lan | Reading package lists...
pihole.lan | Building dependency tree...
pihole.lan | Reading state information...
pihole.lan | rsync is already the newest version (3.2.3-4+deb11u1).
pihole.lan | 0 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.

$ sake ssh server-1
samir@server-1:~ $
```
