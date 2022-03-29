# Command Reference

## sake

sake is a CLI tool that enables you to run commands on servers via ssh

### Synopsis

sake is a CLI tool that enables you to run commands on servers via ssh.

Think of it like make, you define servers and tasks in a declarative configuration file and then run the tasks on the servers.


### Options

```
  -c, --config string        specify config
  -h, --help                 help for sake
      --no-color             disable color
  -u, --user-config string   specify user config
```

## run

Run tasks

### Synopsis

Run tasks specified in a sake.yaml file.

```
run <task> [flags]
```

### Examples

```
  # Run task <task> for all servers
  sake run <task> --all

  # Run task <task> for servers <server>
  sake run <task> --servers <server>

  # Run task <task> for all servers that have tags <tag>
  sake run <task> --tags <tag>
```

### Options

```
  -a, --all                       target all servers
      --any-errors-fatal          stop task execution on all servers on error
      --attach                    ssh to server after command
      --debug                     enable debug mode
      --describe                  print task information
      --dry-run                   print the task to see what will be executed
  -e, --edit                      edit task
  -h, --help                      help for run
  -i, --identity-file string      set identity file for all servers
      --ignore-errors             continue task execution on errors
      --ignore-unreachable        ignore unreachable hosts
      --known-hosts-file string   set known hosts file
      --local                     run task on localhost
      --omit-empty                omit empty results for table output
  -o, --output string             set task output [text|table|html|markdown]
  -p, --parallel                  run server tasks in parallel
      --password string           set ssh password for all servers
  -s, --servers strings           target servers by names
  -t, --tags strings              target servers by tags
      --theme string              set theme
      --tty                       replace the current process
```

## exec

Execute arbitrary commands

### Synopsis

Execute arbitrary commands.

Single quote your command if you don't want the
file globbing and environments variables expansion to take place
before the command gets executed in each directory.

```
exec <command> [flags]
```

### Examples

```
  # List files in all servers
  sake exec --all ls

  # List git files that have markdown suffix for all servers
  sake exec --all 'git ls-files | grep -e ".md"'
```

### Options

```
  -a, --all                       target all servers
      --any-errors-fatal          stop task execution on all servers on error
      --attach                    ssh to server after command
      --debug                     enable debug mode
      --dry-run                   prints the command to see what will be executed
  -h, --help                      help for exec
  -i, --identity-file string      set identity file for all servers
      --ignore-errors             continue task execution on errors
      --ignore-unreachable        ignore unreachable hosts
      --known-hosts-file string   set known hosts file
      --local                     run command on localhost
      --omit-empty                omit empty results for table output
  -o, --output string             set task output [text|table|markdown|html]
  -p, --parallel                  run server tasks in parallel
      --password string           set ssh password for all servers
  -s, --servers strings           target servers by names
  -t, --tags strings              target servers by tags
      --theme string              set theme (default "default")
      --tty                       replace the currenty process
```

## init

Initialize sake in the current directory

### Synopsis

Initialize sake in the current directory.

```
init [flags]
```

### Examples

```
  # Basic example
  sake init
```

### Options

```
  -h, --help   help for init
```

## edit

Open up sake config file in $EDITOR

### Synopsis

Open up sake config file in $EDITOR.

```
edit [flags]
```

### Examples

```
  # Edit current context
  sake edit
```

### Options

```
  -h, --help   help for edit
```

## edit server

Open up sake config file in $EDITOR and go to servers section

### Synopsis

Open up sake config file in $EDITOR and go to servers section.

```
edit server [server] [flags]
```

### Examples

```
  # Edit servers
  sake edit server

  # Edit server <server>
  sake edit server <server>
```

### Options

```
  -h, --help   help for server
```

## edit task

Open up sake config file in $EDITOR and go to tasks section

### Synopsis

Open up sake config file in $EDITOR and go to tasks section.

```
edit task [task] [flags]
```

### Examples

```
  # Edit tasks
  sake edit task

  # Edit task <task>
  sake edit task <task>
```

### Options

```
  -h, --help   help for task
```

## list servers

List servers

### Synopsis

List servers.

```
list servers [servers] [flags]
```

### Examples

```
  # List all servers
  sake list servers

  # List servers <server>
  sake list servers <server>

  # List servers that have tag <tag>
  sake list servers --tags <tag>
```

### Options

```
      --headers strings   set headers. Available headers: server, local, user, host, port, tag, description (default [server,host,tag,description])
  -h, --help              help for servers
  -t, --tags strings      filter servers by tags
```

### Options inherited from parent commands

```
  -o, --output string   set output [table|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## list tags

List tags

### Synopsis

List tags.

```
list tags [tags] [flags]
```

### Examples

```
  # List all tags
  sake list tags
```

### Options

```
      --headers strings   set headers. Available headers: tag, server (default [tag,server])
  -h, --help              help for tags
```

### Options inherited from parent commands

```
  -o, --output string   set output [table|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## list tasks

List tasks

### Synopsis

List tasks.

```
list tasks [tasks] [flags]
```

### Examples

```
  # List all tasks
  sake list tasks

  # List task <task>
  sake list task <task>
```

### Options

```
      --headers strings   set headers. Available headers: task, description, name (default [task,description])
  -h, --help              help for tasks
```

### Options inherited from parent commands

```
  -o, --output string   set output [table|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## describe servers

Describe servers

### Synopsis

Describe servers.

```
describe servers [servers] [flags]
```

### Examples

```
  # Describe all servers
  sake describe servers

  # Describe servers that have tag <tag>
  sake describe servers --tags <tag>
```

### Options

```
  -e, --edit           edit server
  -h, --help           help for servers
  -t, --tags strings   filter servers by their tag
```

## describe tasks

Describe tasks

### Synopsis

Describe tasks.

```
describe tasks [tasks] [flags]
```

### Examples

```
  # Describe all tasks
  sake describe tasks

  # Describe task <task>
  sake describe task <task>
```

### Options

```
  -e, --edit   edit task
  -h, --help   help for tasks
```

## ssh

ssh to server

### Synopsis

ssh to server.

```
ssh <server> [flags]
```

### Examples

```
  # ssh to server
  sake ssh <server>
```

### Options

```
  -h, --help   help for ssh
```

## gen

Generate man page

### Synopsis

Generate man page

```
gen [flags]
```

### Options

```
  -d, --dir string   directory to save manpage to (default "./")
  -h, --help         help for gen
```

