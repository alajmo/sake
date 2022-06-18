# Recipes

A list of useful recipes.

- [Validate Config](#validate-config)
- [Upload File](#upload-file)
- [Download File](#download-file)
- [SSH to Server Using `sake`](#ssh-to-server-using-sake)
- [List Servers, Tasks and Tags](#list-servers-tasks-and-tags)
- [Describe Servers and Tasks](#describe-servers-and-tasks)
- [Edit a Config, Task or Server via `sake`](#edit-a-config-task-or-server-via-sake)
- [Run Command and SSH Afterwords](#run-command-and-ssh-afterwords)
- [Create SSH Tunnel / Port Forward](#create-ssh-tunnel--port-forward)
- [Attach to a Docker Container on a Remote Server](#attach-to-a-docker-container-on-a-remote-server)
- [Run a Local Script on a Remote Server](#run-a-local-script-on-a-remote-server)
- [Replace Current Process](#replace-current-process)
- [Run Server Tasks in Parallel](#run-server-tasks-in-parallel)
- [Aborting on the First Error](#aborting-on-the-first-error)
- [Ignoring Task Errors](#ignoring-task-errors)
- [Ignoring Unreachable Hosts](#ignoring-unreachable-hosts)
- [Omit Table Rows That Return Empty Output](#omit-table-rows-that-return-empty-output)
- [Change Task Output](#change-task-output)
- [Change Working Directory](#change-working-directory)
- [Provide Identity and Password Credentials](#provide-identity-and-password-credentials)
- [Disable Verify Host](#disable-verify-host)
- [Change known_hosts Path](#change-known_hosts-path)
- [List Default Variables](#list-default-variables)
- [Change Default Behavior of `sake`](#change-default-behavior-of-sake)
- [Invoke `sake` From Any Directory](#invoke-sake-from-any-directory)
- [Import a Default User Config for Any `sake` Project](#import-a-default-user-config-for-any-sake-project)
- [What's the Difference Between TTY, Attach and Local?](#whats-the-difference-between-tty-attach-and-local)
- [Disable Colors](#disable-colors)
- [Performing a Dry Run](#performing-a-dry-run)
- [Modify Theme](#modify-theme)

## Validate Config

To check for syntax errors and invalid configurations run:

```bash
$ sake check
```

## Upload File

A common use-case is to upload a file to a server. `sake` doesn't come with any built-in task to accomplish this, but it's quite easy to define one yourself:

```yaml
upload:
  desc: upload file or directory
  env:
    SRC: ""
    DEST: ""
  local: true # Command should be run from local host
  cmd: rsync --recursive --verbose --archive --update $SRC $SAKE_SERVER_HOST:$DEST
```

Then you can refer to the `upload` task:

```yaml
upload-compose:
  desc: upload docker-compose
  env:
    SRC: "./docker-compose.yaml"
    DEST: "/tmp/docker-compose.yaml"
  task: upload
```

To upload a file:

```bash
$ sake run get-backups --server <server>
```

You can also override the `SRC` and `DEST` variables at the command line:

```bash
$ sake run upload --server <server> SRC=/some/file DEST=/some/file
```

Note that rsync is required both on the client and remote machine.

## Download File

A common use-case is to download a file or directory from a server. `sake` doesn't come with any built-in task to accomplish this, but it's quite easy to define one yourself:

Define the `download` task:

```yaml
download:
  desc: download file or directory
  env:
    SRC: ""
    DEST: ""
  local: true # Command should be run from local host
  cmd: rsync --recursive --verbose --archive --update $SAKE_SERVER_HOST:$SRC $DEST
```

Then you can refer to the `download` task:

```yaml
get-backups:
  desc: get backups from remote server
  env:
    SRC: "/tmp/backup.db"
    DEST: "backup.db"
  task: download
```

To download backups:

```bash
$ sake run get-backups --server <server>
```

You can also override the `SRC` and `DEST` variables at the command line:

```bash
$ sake run download --server <server> SRC=/some/file DEST=/some/file
```

Note that rsync is required both on the client and remote machine.

## SSH to Server Using `sake`

You can SSH to any server via `sake ssh <server>`.

## List Servers, Tasks and Tags

The list sub-command will list servers, tasks, and tags in a table, HTML, or Markdown format.

- **Servers**: To list servers run `sake list servers [--tags=<tag>] [server]`

  ```bash
  $ sake list servers --tags remote

   Server    | Host         | Tag        | Description
  -----------+--------------+------------+------------------------
   server-1  | server-1.lan | remote, pi | hosts mealie, node-red
   pihole    | pihole.lan   | remote, pi | runs pihole
  ```

- **Tasks**: To list tasks run `sake list tasks [task]`

  ```bash
  $ sake list tasks

   Task        | Description
  -------------+-------------------------------------
   ping        | ping server
   Host        | print host
   Hostname    | print hostname
   OS          | print OS
   Kernel      | Print kernel version
  ```


- **Tags**: To list tags run `sake list tags [tag]`

  ```bash
  $ sake list tags

   Tag    | Server
  --------+-----------
   local  | localhost
   remote | server-1
          | pihole
   pi     | server-1
          | pihole
  ```

## Describe Servers and Tasks

The describe sub-command describes servers and tasks.

- **Servers**: To describe all servers run `sake describe servers [--tags=<tag>] [server]`

  ```bash
  $ sake describe server pihole

  Name: pihole
  User: samir
  Host: pihole.lan
  Port: 22
  Local: false
  WorkDir:
  Desc: runs pihole
  Tags: remote, pi
  Env:
      SAKE_SERVER_NAME: pihole
      SAKE_SERVER_DESC: runs pihole
      SAKE_SERVER_TAGS: remote,pi
      SAKE_SERVER_HOST: pihole.lan
      SAKE_SERVER_USER: samir
      SAKE_SERVER_PORT: 22
      SAKE_SERVER_LOCAL: false
  ```

- **Tasks**: To describe all tasks run `sake describe tasks [task]`

  ```bash
  $ sake describe task info

  Task: info
  Name: info
  Desc: get remote info
  Local: false
  WorkDir:
  Theme: default
  Target:
      All: true
      Servers:
      Tags:
  Spec:
      Output: table
      Parallel: true
      AnyErrorsFatal: false
      IgnoreErrors: true
      IgnoreUnreachable: true
      OmitEmpty: false
  Env:
      SAKE_TASK_ID: info
      SAKE_TASK_NAME:
      SAKE_TASK_DESC: get remote info
      SAKE_TASK_LOCAL: false
  Tasks:
      - OS: print OS
      - Kernel: Print kernel version
      - Disk: print disk usage
      - Memory: print memory stats
      - CPU: print memory stats
      - Uptime: print uptime
  ```

## Edit a Config, Task or Server via `sake`

You can open up your preferred editor and edit a `sake` config directly via `sake edit [task|server] [name]`. For this to work, the `EDITOR` environment variable must be set.

## Run Command and SSH Afterwords

Sometimes you want to run a command and then `ssh` into the server:

```yaml
ssh-and-cmd:
  desc: run command and ssh to server after
  attach: true # Attach will SSH to server
  cmd: ls -alh
```

Then run:

```bash
$ sake run get-backups --server <server>
```

You can also provide the `--attach` flag to arbitrary commands:

```bash
$ sake run some-task --server <server> --attach
```

## Create SSH Tunnel / Port Forward

Create a SSH tunnel:

```yaml
ssh-tunnel:
  desc: create ssh tunnel
  tty: true # Replacing the current process is necessary if you want to be able to kill the tunnel (in order to respond SIGINT signals)
  env:
    LOCAL:
    REMOTE:
  cmd: ssh $SAKE_SERVER_USER@$SAKER_SERVER_HOST -N -L $LOCAL:localhost:$REMOTE
```

Then run:

```bash
$ sake run ssh-tunnel --server <server> LOCAL=8000 REMOTE=LOCAL=8000
```

## Attach to a Docker Container on a Remote Server

If you have a bunch of Docker containers running on a remote server, you can easily SSH into a remote server and attach to a Docker container.

```yaml
docker-exec:
  desc: attach to docker container
  env:
    NAME: ""
  tty: true # Replacing the current process is necessary since SSH requires TTY if you wish to exec to a container
  cmd: ssh -t $SAKE_SERVER_USER@$SAKE_SERVER_HOST "docker exec -it $NAME bash"
```

Then you can run:

```bash
$ sake run docker-exec --server <server> NAME=<container-name>
```

## Run a Local Script on a Remote Server

Sometimes you have bash script that you want to run on the remote server and after it's done, remove it.
We can do that by defining the following script:

```yaml
script:
  desc: run local script on remote server
  env:
    FILE: ""
  local: true
  cmd: |
    # Get filename
    file=$(basename $FILE)

    # Create temp file
    temp_file="$(mktemp /tmp/$FILE.XXXXXXXXX -u)"

    # Upload script
    rsync --compress --recursive --archive --update $FILE $SAKE_SERVER_HOST:$temp_file

    # Run script
    ssh $SAKE_SERVER_USER@$SAKE_SERVER_HOST "$temp_file"

    # Remove script
    ssh $SAKE_SERVER_USER@$SAKE_SERVER_HOST "rm $temp_file"
```

Then run:

```bash
$ sake run script --server <server> FILE=./script.sh
```

## Replace Current Process

Normally `sake` runs the commands in a new process but you're able to circumvent this by using the `tty: true` setting or provide the `--tty` flag. You rarely need to do this, but there are occassions when it's required, for instance, when you're running interactive tasks that require TTY.

```
echo:
  tty: true
  cmd: echo 123
```

## Run Server Tasks in Parallel

Sometimes you wish to run tasks in parallel, especially when you're just querying information from the machines. In this case, you can use the `--parallel` flag or specify it in the task `spec`. Note that if your tasks has multiple commands, the commands will still execute sequentially for each server, it's just that the overall server execution will happen in parallel.

```yaml
tasks:
  print-kernel:
    name: Kernel
    desc: Print kernel version
    spec:
      parallel: true
    cmd: uname -r | awk -v FS='-' '{print $1}'
```

```bash
$ sake run print-kernel --all
```

## Aborting on the First Error

If you wish to abort all tasks on all errors in case an error is encountered for any task, use the flag `--any-errors-fatal` or specify it in the task `spec`.

```yaml
fatal:
  spec:
    any_errors_fatal: true
  tasks:
    - cmd: echo 123
    - cmd: exit 1
    - cmd: echo 321
```

See example:

```bash
# any-errors-fatal set to false
$ sake run fatal --all --output table --any-errors-fatal=false

 Server    | Output | Output                       | Output
-----------+--------+------------------------------+--------
 localhost | 123    |                              |
           |        | exit status 1                |
 server-1  | 123    |                              |
           |        | Process exited with status 1 |
 pihole    | 123    |                              |
           |        | Process exited with status 1 |

# any-errors-fatal set to true
$ sake run fatal --all --output table --any-errors-fatal=true

Server    | Output | Output        | Output
----------+--------+---------------+--------
localhost | 123    |               |
          |        | exit status 1 |
server-1  |        |               |
pihole    |        |               |
```

## Ignoring Task Errors

If you wish to continue task execution even if an error is encountered, use the flag `--ignore-errors` or specify it in the task `spec`.

```yaml
errors:
  spec:
    ignore_errors: false
  tasks:
    - cmd: echo 123
    - cmd: exit 1
    - cmd: echo 321
```

See example:

```bash
# ignore-errors set to false
$ sake run errors --all --output table --ignore-errors=false

 Server    | Output | Output                       | Output
-----------+--------+------------------------------+--------
 localhost | 123    |                              |
           |        | exit status 1                |
 server-1  | 123    |                              |
           |        | Process exited with status 1 |
 pihole    | 123    |                              |
           |        | Process exited with status 1 |

# ignore-errors set to true
$ sake run errors --all --output table --ignore-errors=true

 Server    | Output | Output                        | Output
-----------+--------+-------------------------------+--------
 localhost | 123    |                               | 321
           |        | exit status 65                |
 server-1  | 123    |                               | 321
           |        | Process exited with status 65 |
 pihole    | 123    |                               | 321
           |        | Process exited with status 65 |
```

## Ignoring Unreachable Hosts

Sometimes you want to ignore remote hosts which are unreachable, for instance if it's a host that is flaky, then you can either use the `--ignore-unreachable` flag or specify it in the task `spec`.

```yaml
unreachable:
  spec:
    ignore_unreachable: false
  cmd: echo 123
```

See example:

```bash
# ignore-unreachable set to false
$ sake run unreachable --all --output table --ignore-unreachable=false

Unreachable Hosts

 Server   | Host        | User  | Port | Error
----------+-------------+-------+------+----------------------------------------------------------------
 server-1 | server1.lan | samir | 22   | dial tcp: lookup server1.lan on x.y.z.k:33: no such host

# ignore-unreachable set to true
$ sake run unreachable --all --output table --ignore-unreachable=true

Unreachable Hosts

 Server   | Host        | User  | Port | Error
----------+-------------+-------+------+----------------------------------------------------------------
 server-1 | server1.lan | samir | 22   | dial tcp: lookup server1.lan on 192.168.1.209:53: no such host

 Server    | Unreachable
-----------+-------------
 localhost | 123
 pihole    | 123
```

## Omit Table Rows That Return Empty Output

If you wish to omit rows that return empty outputs, you can do so via the `--omit-empty` flag or specify it in the task `spec`.
For instance, the below command will check if a directory `.ssh` exists in the users home directory.

```yaml
empty:
  spec:
    omit_empty: false
  cmd: |
    if [[ -d ".ssh" ]]
    then
        echo "Exists"
    fi
```

See example:

```bash
# omit-empty set to false
$ sake run empty --all --output table --omit-empty=false

 Server    | Empty
-----------+--------
 localhost |
 server-1  | Exists
 pihole    | Exists

# omit-empty set to true
$ sake run empty --all --output table --omit-empty=true

 Server   | Empty
----------+--------
 server-1 | Exists
 pihole   | Exists
```

## Change Task Output

`sake` supports different output formats for tasks. By default it will use `text` output, but it's possible to change this via the `--output` flag or specify it in the task `spec`. Possible formats are `text`, `table`, `html` and `markdown`.

```yaml
output:
  spec:
    output: text
  tasks:
    - cmd: echo "Hello world"
    - cmd: echo "Bye world"
    - cmd: echo "Hello again world"
```

See example:

```bash
# output set to text
$ sake run output --all --output text

TASK (1/3) Command **********************

server-1.lan | Hello world

TASK (2/3) Command **********************

server-1.lan | Bye world

TASK (3/3) Command **********************

server-1.lan | Hello again world

# output set to table
$ sake run output --all --output table

 Server   | Output      | Output    | Output
----------+-------------+-----------+-------------------
 server-1 | Hello world | Bye world | Hello again world

# output set to html
$ sake run output --all --output html

<table class="">
  <thead>
  <tr>
    <th align="left" class="bold">server</th>
    <th align="left" class="bold">output</th>
    <th align="left" class="bold">output</th>
    <th align="left" class="bold">output</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td align="left">server-1</td>
    <td align="left">Hello world</td>
    <td align="left">Bye world</td>
    <td align="left">Hello again world</td>
  </tr>
  </tbody>
</table>

# output set to markdown
$ sake run output --all --output markdown

| server | output | output | output |
|:--- |:--- |:--- |:--- |
| server-1 | Hello world | Bye world | Hello again world |
```

## Change Working Directory

You can change the default `work_dir` in the server section and the task section (nested tasks/commands included).
The order of precedence is as follows:

1. task list
2. task
3. referenced task
4. server
5. default, which is the current working directory for local clients and `/home/user` for remote clients

```yaml
servers:
  localhost:
    host: localhost
    work_dir: "/opt" # 4
    local: true

tasks:
  work-ref:
    name: pwd
    work_dir: "/usr" # 3
    cmd: pwd

  work-dir:
    work_dir: "/home" # 2
    tasks:
      - task: work-ref

      - cmd: pwd
        name: pwd

      - cmd: pwd
        name: pwd
        work_dir: "/" # 1
```

See example:

```bash
$ sake run work-dir --output table

 Server    | Pwd   | Pwd   | Pwd
-----------+-------+-------+-----
 localhost | /home | /home | /

# if we comment work_dir (# 2) then we get

 Server    | Pwd  | Pwd  | Pwd
-----------+------+------+-----
 localhost | /usr | /opt | /
```

## Provide Identity and Password Credentials

By default `sake` will attempt to load identity keys from an SSH agent if it's running in the background. However, if you wish to provide credentials manually, you can do so by (first takes precedence):

1. setting `--identity-file` and/or `--password` flags
2. providing environment variables `SAKE_IDENTITY_FILE` and `SAKE_PASSWORD`
3. specifying it in the server definition

The type of auth used is determined by:

- if `identity-file` and `password` are provided, then it assumes password protected identity key
- if only `identity-file` is provided, then it assumes a passwordless identity key
- if only `password` is provided, then it assumes password protected auth

```yaml
servers:
  server-1:
    host: server-1.lan
    identity_file: id_rsa
    password: $(echo $MY_SECRET_PASSWORD)
```

## Disable Verify Host

By default a `known_hosts` file is used to verify host connections. If you wish to disable verification, set the global property `disable_verify_host` to true:

```yaml
disable_verify_host: true
```

## Change known_hosts Path

By default a `known_hosts` file is used to verify host connections. It's default location is `$HOME/.ssh/known_hosts`. If you wish change this to another file, then set the global property `known_hosts_file` to your desired filepath:

```yaml
known_hosts_file: ./known_hosts
```

## List Default Variables

Each task has access to a number of default environment variables.

```yaml
  env:
    cmd: |
      echo "# SERVER"
      echo "SAKE_SERVER_NAME $SAKE_SERVER_NAME"
      echo "SAKE_SERVER_DESC $SAKE_SERVER_DESC"
      echo "SAKE_SERVER_TAGS $SAKE_SERVER_TAGS"
      echo "SAKE_SERVER_HOST $SAKE_SERVER_HOST"
      echo "SAKE_SERVER_USER $SAKE_SERVER_USER"
      echo "SAKE_SERVER_PORT $SAKE_SERVER_PORT"
      echo "SAKE_SERVER_LOCAL $SAKE_SERVER_LOCAL"

      echo
      echo "# TASK"
      echo "SAKE_TASK_ID $SAKE_TASK_ID"
      echo "SAKE_TASK_NAME $SAKE_TASK_NAME"
      echo "SAKE_TASK_DESC $SAKE_TASK_DESC"
      echo "SAKE_TASK_LOCAL $SAKE_TASK_LOCAL"

      echo
      echo "# CONFIG"
      echo "SAKE_DIR $SAKE_DIR"
      echo "SAKE_PATH $SAKE_PATH"
      echo "SAKE_IDENTITY_FILE $SAKE_IDENTITY_FILE"
      echo "SAKE_PASSWORD $SAKE_PASSWOD"
      echo "SAKE_KNOWN_HOSTS_FILE $SAKE_KNOWN_HOSTS_FILE"
```

See example:

```bash
$ sake run env -s server-1

 Server   | Env
----------+---------------------------------------------------------------
 server-1 | # SERVER
          | SAKE_SERVER_NAME server-1
          | SAKE_SERVER_DESC server-1 description
          | SAKE_SERVER_TAGS remote,pi
          | SAKE_SERVER_HOST server-1.lan
          | SAKE_SERVER_USER test
          | SAKE_SERVER_PORT 22
          | SAKE_SERVER_LOCAL false
          |
          | # TASK
          | SAKE_TASK_ID env
          | SAKE_TASK_NAME
          | SAKE_TASK_DESC print all default env variables
          | SAKE_TASK_LOCAL false
          |
          | # CONFIG
          | SAKE_DIR /tmp
          | SAKE_PATH /tmp/sake.yaml
          | SAKE_IDENTITY_FILE
          | SAKE_PASSWORD
          | SAKE_KNOWN_HOSTS_FILE
```

## Change Default Behavior of `sake`

`sake` comes with default definitions for `specs`, `targets` and `themes` (see [config reference](config-reference) for their default values). This means when you run `sake list servers` or `sake run <task>` without specifying any spec/target/theme on the command line or in the config, it will use the default definition for those primitives.
To override the default config, we can define a spec/target/theme that has the name `default`:

For instance, let's target all servers by default:

```yaml
targets:
 default:
   all: true
```

Now when you run `sake run <task>`, it will target all servers by default.

## Invoke `sake` From Any Directory

When you invoke a `sake` command it will check the current directory and all parent directories for the following files: `sake.yaml`, `sake.yml`, `.sake.yaml`, `.sake.yml`. If you wish to invoke `sake` from any directory, you can:

- set the environment variable to `SAKE_CONFIG=/path/to/my/config`, or
- specify a runtime flag `sake list servers --config /path/to/my/config`

## Import a Default User Config for Any `sake` Project

By default `sake` will attempt to load a config file (if it exists) from your default config directory:

- Linux: `$XDG_CONFIG_HOME/sake/config.yaml` or `$HOME/.config/sake/config.yaml` if `$XDG_CONFIG_HOME` is not set.
- Darwin: `$HOME/Library/Application/sake/config.yaml`

You can override this location by:

- setting the environment variable to `SAKE_USER_CONFIG=/path/to/my/config`, or
- specifying a runtime flag `sake list servers --user-config /path/to/my/config`

## What's the Difference Between TTY, Attach and Local?

- When specifying `tty: true` in a task config, the calling executable will be replaced by the command invoked by the task. This is useful when you require `tty`, for instance if you want to SSH and then attach to a running Docker container
- If `attach: true` is set in a task config, then after running all the commands, `sake` will SSH into the first remote server
- Setting `local: true` means the task will be executed on localhost, this can be useful for tasks that upload files via `rsync` for instance

## Disable Colors

To disable colors from `sake`, either add the flag `--no-color` or set the environment variable `NO_COLOR`.

## Performing a Dry Run

If you wish to perform a dry run you can do so by adding the flag `--dry-run`. It will then only print out the task for each server.

## Modify Theme

`sake` allows you to modify the output of tasks by creating themes for different situations. You can do so either inline when defining tasks, refer to a theme in from the global `themes` definition, or provide the `--theme` flag.
A `theme` has two objects that alter the style of the different outputs: `text` and `table`. See the [config-reference](/config-reference) document for more details.

```yaml
themes:
  advanced:
    text:
      prefix: true
      header: true
      header_prefix: TASK
      header_char: "-"
      colors: [red,green,blue]

    table:
      style: connected-light

tasks:
  ping:
    cmd: echo pong
    theme: advanced
    # or define inline
    theme:
      text:
        prefix: true

      table:
        style: connected-light
```
