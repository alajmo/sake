# Command Reference

## sake

sake is a command runner for local and remote hosts

### Synopsis

sake is a command runner for local and remote hosts.

You define servers and tasks in a sake.yaml config file and then run the tasks on the servers.


### Options

```
  -c, --config string        specify config
  -h, --help                 help for sake
      --no-color             disable color
  -U, --ssh-config string    specify ssh config
  -u, --user-config string   specify user config
```

## check

Validate config

### Synopsis

Validate config.

```
check [flags]
```

### Examples

```
  # Validate config
  sake check
```

### Options

```
  -h, --help   help for check
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
      --describe                  print task information
      --dry-run                   print the task to see what will be executed
  -e, --edit                      edit task
  -h, --help                      help for run
  -i, --identity-file string      set identity file for all servers
      --ignore-errors             continue task execution on errors
      --ignore-unreachable        ignore unreachable hosts
  -v, --invert                    invert matching on servers
      --known-hosts-file string   set known hosts file
  -l, --limit uint32              set limit of servers to target
  -L, --limit-p uint8             set percentage of servers to target [0-100]
      --local                     run task on localhost
      --omit-empty                omit empty results for table output
  -o, --output string             set task output [text|table|table-2|table-3|table-4|html|markdown]
  -p, --parallel                  run server tasks in parallel
      --password string           set ssh password for all servers
  -r, --regex string              filter servers on host regex
  -s, --servers strings           target servers by names
  -S, --silent                    omit showing loader when running tasks
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
      --dry-run                   prints the command to see what will be executed
  -h, --help                      help for exec
  -i, --identity-file string      set identity file for all servers
      --ignore-errors             continue task execution on errors
      --ignore-unreachable        ignore unreachable hosts
  -v, --invert                    invert matching on servers
      --known-hosts-file string   set known hosts file
  -l, --limit uint32              set limit of servers to target
  -L, --limit-p uint8             set percentage of servers to target
      --local                     run command on localhost
      --omit-empty                omit empty results for table output
  -o, --output string             set task output [text|table|table-2|table-3|table-4|html|markdown]
  -p, --parallel                  run server tasks in parallel
      --password string           set ssh password for all servers
  -r, --regex string              filter servers on host regex
  -s, --servers strings           target servers by names
  -S, --silent                    omit showing loader when running tasks
  -t, --tags strings              target servers by tags
      --theme string              set theme (default "default")
      --tty                       replace the current process
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

Edit server

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

## edit spec

Edit spec

### Synopsis

Open up sake config file in $EDITOR and go to specs section.

```
edit spec [spec] [flags]
```

### Examples

```
  # Edit specs
  sake edit spec

  # Edit spec <spec>
  sake edit spec <spec>
```

### Options

```
  -h, --help   help for spec
```

## edit target

Edit target

### Synopsis

Open up sake config file in $EDITOR and go to targets section.

```
edit target [target] [flags]
```

### Examples

```
  # Edit targets
  sake edit target

  # Edit target <target>
  sake edit target <target>
```

### Options

```
  -h, --help   help for target
```

## edit task

Edit task

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
  -H, --all-headers       select all server headers
      --headers strings   set headers (default [server,host,tag,desc])
  -h, --help              help for servers
  -v, --invert            invert matching on servers
  -r, --regex string      filter servers on host regex
  -t, --tags strings      filter servers by tags
```

### Options inherited from parent commands

```
  -o, --output string   set table output [table|table-2|table-3|table-4|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## list specs

List specs

### Synopsis

List specs.

```
list specs [specs] [flags]
```

### Examples

```
  # List all specs
  sake list specs
```

### Options

```
      --headers strings   set headers (default [spec,output,parallel,any_errors_fatal,ignore_errors,ignore_unreachable,omit_empty])
  -h, --help              help for specs
```

### Options inherited from parent commands

```
  -o, --output string   set table output [table|table-2|table-3|table-4|markdown|html] (default "table")
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
      --headers strings   set headers (default [tag,server])
  -h, --help              help for tags
```

### Options inherited from parent commands

```
  -o, --output string   set table output [table|table-2|table-3|table-4|markdown|html] (default "table")
      --theme string    set theme (default "default")
```

## list targets

List targets

### Synopsis

List targets.

```
list targets [targets] [flags]
```

### Examples

```
  # List all targets
  sake list targets
```

### Options

```
      --headers strings   set headers. Available headers: name, regex (default [target,all,servers,tags,regex,invert,limit,limit_p])
  -h, --help              help for targets
```

### Options inherited from parent commands

```
  -o, --output string   set table output [table|table-2|table-3|table-4|markdown|html] (default "table")
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
  -H, --all-headers       select all task headers
      --headers strings   set headers (default [task,desc])
  -h, --help              help for tasks
```

### Options inherited from parent commands

```
  -o, --output string   set table output [table|table-2|table-3|table-4|markdown|html] (default "table")
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
  -v, --invert         invert matching on servers
  -r, --regex string   filter servers on host regex
  -t, --tags strings   filter servers by their tag
```

## describe specs

Describe specs

### Synopsis

Describe specs.

```
describe specs [specs] [flags]
```

### Examples

```
  # Describe all specs
  sake describe specs
```

### Options

```
  -e, --edit   edit spec
  -h, --help   help for specs
```

## describe targets

Describe targets

### Synopsis

Describe targets.

```
describe targets [targets] [flags]
```

### Examples

```
  # Describe all targets
  sake describe targets
```

### Options

```
  -e, --edit   edit target
  -h, --help   help for targets
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
  -h, --help                   help for ssh
  -i, --identity-file string   set identity file for all servers
      --password string        set ssh password for all servers
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

