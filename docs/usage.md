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
$ sake exec --all --output table --parallel 'find . -type f | wc -l'

 Server    | Output
-----------+--------
 localhost | 1
```

Next up:

- [Some more examples](/examples)
- [Familiarize yourself with the sake.yaml config](/config)
- [Checkout sake commands](/commands)
