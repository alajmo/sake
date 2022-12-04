# Examples

This is an example of how you can use `sake` for localhost. If you wish to run commands against remote servers via `ssh`, add your own remote servers.

- [Simple Example](#simple-example)
- [Advanced Example](#advanced-example)
- [Real World Example](#real-world-example)

## Simple Example

This is a simple example where we just define 1 server and a couple of tasks.

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
    cmd: echo $S_HOST

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

task: print-host
name: Host
desc: print host
theme: default
target:
    all: true
spec:
    output: table
cmd:
    echo $S_HOST


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
$ sake exec --all --output table --strategy=free 'cd ~ && ls -al | wc -l'

 Server    | Output
-----------+--------
 localhost | 42
```

## Advanced Example

This is a more advanced example where we introduce `spec`, `target`, `imports`, and nested tasks.

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
    omit_empty_rows: true
    omit_empty_columns: true
    any_fatal_errors: false
    ignore_unreachable: true
    strategy: free

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
    cmd: echo $S_HOST

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

## Real World Example

This is an example of how you can setup and manage a bunch of Raspberry Pi's.

It contains the following workflows:

- Create tasks to setup Raspberry pi
  - clean home directory
  - create new directories
  - install some default packages
  - set default editor and hostname
  - disable swap
  - increase inotify
  - turn off wifi and bluetooth
- Install scripts for
  - Docker, Docker Compose, and Pi-Hole
- Tasks that query for machine info (cpu, memory, Pi Version, etc.)
- Tasks to upload files and run docker-compose commands

Let's start by creating the following files:

```bash
.
├── sake.yaml           # base config
├── common.yaml         # tasks used across all servers
├── utils.yaml          # utility tasks
├── install.yaml        # install various software
└── docker-compose.yaml # Docker containers
```

and populate them with the associated content.

### sake.yaml

This is our base config which will be used everytime we run a command.

```yaml title=sake.yaml
import:
  - common.yaml
  - utils.yaml
  - install.yaml

servers:
  server-1:
    desc: server-1 hosts nodered, syncthing and mealie
    host: server-1.lan
    tags: [active, remote, pi, server]
    env:
      HOSTNAME: server-1

  pihole:
    desc: pihole and local DNS resolver
    host: pihole.lan
    tags: [active, remote, pi, pihole]
    env:
      HOSTNAME: pihole

targets:
  all:
    all: true

specs:
  info:
      output: table
      strategy: free
      ignore_errors: true
      ignore_unreachable: true
      any_errors_fatal: false

tasks:
  ping:
    desc: print pong from server
    target: all
    cmd: echo pong

  real-ping:
    name: Ping for real
    desc: ping server
    target: all
    local: true
    cmd: ping $S_HOST -c 2

  # Setup
  setup-pi:
    name: Setup pi
    desc: update hostname, install common software, etc.
    target: all
    tasks:
      - task: clean-home
      - task: install-default-packages
      - task: secure-pi
      - task: disable-swap
      - task: set-default-editor
      - task: increase-inotify
      - task: set-hostname

  setup-pihole:
    name: Setup Pi-hole
    target:
      tags: [pihole]
    tasks:
      - task: setup-pi
      - task: install-pihole

  setup-server:
    name: Setup Server
    target:
      tags: [server]
    tasks:
      - task: setup-pi
      - task: install-docker
      - task: upload
        name: upload docker compose config
        env:
          SRC: docker-compose.yaml
          DEST: docker-compose.yaml

  # Daily Dev

  attach-mealie:
    desc: attach to mealie
    env:
      NAME: "mealie"
    task: docker-exec

  attach-nodered:
    desc: attach to nodered
    env:
      NAME: "nodered"
    task: docker-exec

  server-workflow:
    desc: Upload docker-compose config and restart services
    targets:
      tags: [server]
    tasks:
      - task: upload
        name: upload docker compose config
        env:
          SRC: docker-compose.yaml
          DEST: docker-compose.yaml
      - task: docker-start
```

### common.yaml

Common tasks used to setup Raspberry Pi's.

```yaml title=common.yaml
tasks:
  clean-home:
    name: Cleanup home
    desc: Remove unused directories in home and create some defaults
    cmd: |
      cd ~
      rm Bookshelf Desktop Documents Pictures Public Templates Music Downloads Videos -rf
      mkdir -p downloads tmp sandbox

  install-default-packages:
    name: Install default packages
    desc: install default packages
    cmd: |
      sudo apt-get update -y
      sudo apt-get upgrade -y
      sudo apt-get install sysstat vim vifm rfkill tree htop jq curl sqlite3 -y
      sudo apt autoremove -y

  secure-pi:
    name: Secure PI
    desc: secure pi, block wifi, bluetooth, etc.
    cmd: |
      sudo rfkill block wifi
      sudo rfkill block bluetooth

  set-default-editor:
    name: Set default editor
    desc: set default editor
    cmd: |
      sudo update-alternatives --install /usr/bin/editor editor /usr/bin/vim 1
      sudo update-alternatives --set editor /usr/bin/vim

  increase-inotify:
    name: Increase inotify
    desc: increase inotify watches, useful for syncthing
    cmd: echo "fs.inotify.max_user_watches=204800" | sudo tee -a /etc/sysctl.conf

  disable-swap:
    name: Disable swap
    desc: disable swap
    cmd: sudo systemctl disable dphys-swapfile.service

  set-hostname:
    name: Set hostname
    desc: sets the hostname
    cmd: |
      sudo hostnamectl set-hostname $HOSTNAME
      sudo sed -i -r 's/raspberrypi/$HOSTNAME/' /etc/hosts
```

### utils.yaml

Useful utility tasks that provide machine information (OS, Kernel, Pi Version, etc.), upload/download tasks, and docker-compose commands.

```yaml title="utils.yaml"
tasks:
  print-host:
    name: Host
    desc: print host
    spec: info
    target: all
    cmd: echo $S_HOST

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

  print-pi-version:
    name: Pi
    desc: print pi version
    spec: info
    target: all
    cmd: cat /proc/device-tree/model

  print-kernel:
    name: Kernel
    desc: Print kernel version
    spec: info
    target: all
    cmd: uname -r | awk -v FS='-' '{print $1}'

  print-mem:
    name: Memory
    desc: print memory stats
    spec: info
    target: all
    cmd: |
      mem_tot=$(awk '$1 == "MemTotal:" { print $2 / 1024 / 1024 }' /proc/meminfo)
      mem_tot=$(printf "%.1f" $mem_tot)

      mem_free=$(awk '$1 == "MemAvailable:" { print $2 / 1024 / 1024 }' /proc/meminfo)
      mem_free=$(printf "%.1f" $mem_free)
      mem_used=$(echo "$mem_tot-$mem_free" | bc)

      echo "$mem_used / $mem_tot Gb"

  print-cpu:
    name: CPU
    desc: print memory stats
    spec: info
    target: all
    cmd: |
      num_cores=$(nproc --all)
      cpu_load=$(mpstat 1 2 | awk 'END{print 100-$NF"%"}')
      echo "$cpu_load, ($num_cores cores)"

  print-disk:
    name: Disk
    desc: print disk usage
    spec: info
    target: all
    cmd: |
      disk=$(/usr/bin/df -BG 2>/dev/null | tail -n +2 | sort -h -k2,2 -r | awk -F " " '{print $1, $2, $3}' | head -n 1)

      tot_disk=$(echo $disk | awk '{print $2}')
      used_disk=$(echo $disk | awk '{print $3}')

      echo "$used_disk / $tot_disk"

  print-uptime:
    name: Uptime
    desc: print uptime
    spec: info
    target: all
    cmd: uptime | grep -E -o "[0-9]* (day|days)"

  info:
    desc: get remote info
    spec: info
    target: all

    tasks:
      - task: print-os
      - task: print-pi-version
      - task: print-kernel
      - task: print-disk
      - task: print-mem
      - task: print-cpu
      - task: print-uptime

  # Upload

  upload:
    desc: upload file or directory
    env:
      SRC: ""
      DEST: ""
    local: true
    cmd: rsync --recursive --verbose --archive --update $SRC $S_HOST:$DEST

  # Docker

  docker-exec:
    desc: attach to docker container
    env:
      NAME: ""
    tty: true
    cmd: ssh -t $S_USER@$S_HOST "docker exec -it $NAME bash"

  docker-start:
    desc: create and start services
    cmd: docker-compose up --detach

  docker-stop:
    desc: stop services
    cmd: docker-compose stop

  docker-pause:
    desc: stop services
    cmd: docker-compose pause

  docker-unpause:
    desc: stop services
    cmd: docker-compose unpause
```

### install.yaml

Tasks to install `Pi-hole`,`Docker`, and `Docker Compose`.

```yaml title=install.yaml
tasks:
  install-pihole:
    name: Install pihole
    desc: Install pihole
    cmd: curl -sSL https://install.pi-hole.net | bash

  install-docker:
    name: Install Docker
    desc: Install docker and docker-compose
    env:
      USER: samir
    cmd: |
      sudo apt-get remove docker docker.io containerd runc -y
      sudo apt-get update -y
      sudo apt-get install    \
                   apt-transport-https \
                   ca-certificates     \
                   curl                \
                   gnupg               \
                   lsb-release \
                   -y

      curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

      echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian \
        $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

      sudo apt-get update -y
      sudo apt-get install \
                   docker-ce \
                   docker-ce-cli \
                   containerd.io \
                   -y

      # Install docker-compose
      mkdir -p ~/.docker/cli-plugins \
        && wget https://github.com/docker/compose/releases/download/v2.1.1/docker-compose-linux-aarch64 -O ~/.docker/cli-plugins/docker-compose \
        && chmod +x ~/.docker/cli-plugins/docker-compose

      # Add USER user to docker group
      sudo usermod -aG docker $USER
```

### docker-compose.yaml

This is a docker-compose file used to start multiple Docker containers. Currently three services are ran `syncthing`, `mealie`, and `Node-RED`.

```yaml title="docker-compose.yaml"
version: "3.9"

services:
  syncthing:
    image: syncthing/syncthing
    container_name: syncthing
    ports:
      - 8384:8384
      - 22000:22000/tcp
      - 22000:22000/udp
      - 21027:21027/udp

    environment:
      PUID: 1001
      PGID: 1001
      TZ: Europe/Stockholm

    volumes:
      - "./.config/syncthing:/var/syncthing"

    restart: unless-stopped

  node-red:
    image: nodered/node-red:latest
    container_name: nodered
    ports:
      - "1880:1880"
    user: root

    environment:
      TZ: Europe/Stockholm

    volumes:
      - "./.config/syncthing/nodered:/data"

    restart: unless-stopped

  mealie:
    container_name: mealie
    image: hkotel/mealie:latest
    restart: always
    ports:
      - 9925:80

    environment:
      DB_ENGINE: sqlite
      PUID: 1000
      PGID: 1000
      TZ: Europe/Stockholm

      # Default Recipe Settings
      RECIPE_PUBLIC: "true"
      RECIPE_SHOW_NUTRITION: "true"
      RECIPE_SHOW_ASSETS: "true"
      RECIPE_LANDSCAPE_VIEW: "true"
      RECIPE_DISABLE_COMMENTS: "false"
      RECIPE_DISABLE_AMOUNT: "false"

    volumes:
      - ./.config/mealie/data/:/app/data

networks:
  server:
    external: true
```

### Workflow

Now we can run some commands:

```bash
# Try to ping servers
$ sake run ping

# Setup Pi's
$ sake run setup-pi

# Setup Pi-Hole
$ sake run setup-pihole

# Setup Generic Server
$ sake run setup-server

# Get some machine info
$ sake run info --all

# Make modifications to docker-compose, upload it and restart it
$ sake run server-workflow --tags server-1
```
