# Recipes

This file contains various recipes for common tasks you might need, such as:

- Upload/download files via `rsync`
- Run command then SSH into server
- Attach to a Docker container on a remote machine
- Create a SSH Tunnel / Port forward
- Run local script on a remote machine
- Replace current process

## Upload File

Define the `upload` task:

```yaml
upload:
  desc: upload file or directory
  env:
    SRC: ""
    DEST: ""
  local: true # Command should be run from local host
  cmd: rsync --recursive --verbose --archive --update $SRC $SAKE_SERVER_HOST:$DEST
```

Then you can refer to `upload` task:

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

## Download File

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

Then you can refer to `download` task:

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

## Attach to a Docker Instance on a Remote Server

If you have a bunch of Docker containers running on a remote server, you can easily ssh into the remote server and attach to the Docker instance.

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


## Run Local Script on Remote Server

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

Normally `sake` runs the commands in a new process but you're able to circumvent this by using the `tty: true` setting or provide the `--tty` flag.` You rarely need to do this, but there are occassions when it's required, for instance, when you're running interactive tasks that require TTY.

```
echo:
  tty: true
  cmd: echo 123
```
