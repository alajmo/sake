# Usage

## Create a New Sake Config

Run the following command:

```bash
$ sake init

Initialized sake in /tmp/sake
- Created sake.yaml

Following servers were added to sake.yaml

 Server    | Host
-----------+---------
 localhost | 0.0.0.0
```

Our `sake.yaml` config file should look like this:

```yaml title=sake.yaml
servers:
  localhost:
    host: 0.0.0.0
    local: true

tasks:
  ping:
    desc: Pong
    cmd: echo "pong"
"```

## Run Some Commands

Now let's run some commands to see everything is working as expected.

```bash
# List all servers
$ sake list servers

 Server    | Host
-----------+---------
 localhost | 0.0.0.0

# List all tasks
$ sake list tasks

 Task | Description
------+-------------
 ping | Pong

# Run Task
$ sake run ping --all

TASK ping: Pong ************

0.0.0.0 | pong

# Count number of files in each servers in parallel
$ sake exec --all --output table --strategy=free 'find . -type f | wc -l'

 Server    | Output
-----------+--------
 localhost | 1
```

Next up:

- [Simple examples](/examples)
- [Recipes](/recipes)
- [Config Reference](/config-reference)
- [Command Reference](/command-reference)
