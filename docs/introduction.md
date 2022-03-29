---
slug: /
---

# Introduction

`sake` is a CLI tool that enables you to run commands on servers via `ssh`. Think of it like `make`, you define servers and tasks in a declarative configuration file and then run the tasks on the servers.

It has many ergonomic features such as `auto-completion` of tasks, servers and tags.
Additionally, it includes sub-commands to let you easily

- `ssh` into servers or docker containers
- list and describe servers/tasks
- open up the `sake.yaml` file and go directly to the server/task you wish to edit


![demo](/img/output.gif)

# Example

You specify servers and tasks in a config file:

```yaml title="sake.yaml"
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
    cmd: echo $SAKE_SERVER_HOST

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
      parallel: true
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
