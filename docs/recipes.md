# Recipes

A list of useful recipes.

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
  cmd: rsync --recursive --verbose --archive --update $SRC $S_HOST:$DEST
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
  cmd: rsync --recursive --verbose --archive --update $S_HOST:$SRC $DEST
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
  cmd: ssh $S_USER@$S_HOST -N -L $LOCAL:localhost:$REMOTE
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
  cmd: ssh -t $S_USER@$S_HOST "docker exec -it $NAME bash"
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
    rsync --compress --recursive --archive --update $FILE $S_HOST:$temp_file

    # Run script
    ssh $S_USER@$S_HOST "$temp_file"

    # Remove script
    ssh $S_USER@$S_HOST "rm $temp_file"
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

## Change Shell

You can change the default `shell` for tasks by setting the `shell` property in the global scope, server section or the task section (nested tasks/commands included).

The order of precedence is as follows (first takes precedence):

1. task list
2. task
3. referenced task
4. server
5. global
6. default which is `bash` for Linux, `powershell` for windows, and `zsh` for MacOS.

For remote servers, the default shell is the users default shell.

```yaml

shell: bash # 5

servers:
  localhost:
    host: localhost
    shell: bash # 4
    local: true

tasks:
  work-ref:
    name: pwd
    shell: bash # 3
    cmd: pwd

  work-dir:
    shell: bash # 2
    tasks:
      - task: work-ref

      - cmd: pwd
        name: pwd

      - cmd: pwd
        name: pwd
        shell: bash # 1
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

## Edit a Config, Task or Server via `sake`

You can open up your preferred editor and edit a `sake` config directly via `sake edit [task|server] [name]`. For this to work, the `EDITOR` environment variable must be set.

## Modify Theme

`sake` allows you to modify the output of tasks by creating themes for different situations. You can do so either inline when defining tasks, refer to a theme in from the global `themes` definition, or provide the `--theme` flag.
A `theme` has two objects that alter the style of the different outputs: `text` and `table`. See the [config-reference](/config-reference) document for more details.

```yaml
themes:
  advanced:
    text:
      prefix: true
      prefix_colors: [red,green,blue]
      header: "TASK"
      header_filler: "-"

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
