# Changelog

## 0.15.1

### Fixes

- Fix resolving identity file in ssh config correctly when ~ is used.
- Fix public file not found when using ssh config

## 0.15.0

### Features

- Add support for multiple bastions

### Misc

- Update to go 1.20

### Fixes

- Fix issue where user/port was not set correctly when using shorthand format for host/user/port definition

## 0.14.0

### Features

- Add ability to modify prefix in text and table themes
- Hide tasks from auto-completion via spec attribute `hidden: true`
- Add print option to limit output to stdout|stderr
- Default to one of following identity files if no identity specified `~/.ssh/id_rsa`, `~/.ssh/id_ecdsa`, `~/.ssh/id_dsa`
- Add ability to modify default timeout for ssh connections

### Fixes

- [BREAKING CHANGE]: No more duplicate tasks, specs, targets, and themes
- Small fix when user config is specified but not found
- Fix some small validation issues with batch and batch-p
- A bunch of smaller fixes

## 0.13.0

### Features

- Add ability to register variables which are available to the next tasks
- Add option to ignore errors for indiviual tasks
- Add flag/spec `--list-hosts` option to list targetted hosts
- Support output options `csv`/`json`/`none`
- Add new task strategies: linear, host_pinned, free
  - `linear`: execute task for each host before proceeding to the next task (default)
  - `host_pinned`: executes tasks (serial) for a host before proceeding to the next host
  - `free`: tasks without waiting for other tasks
- Add host ordering
  - `inventory`: The order is as provided by the inventory
  - `reverse_inventory`: The order is the reverse of the inventory
  - `sorted`: Hosts are alphabetically sorted by host
  - `reverse_sorted`: Hosts are sorted by host in reverse alphabetical order
  - `random`: Hosts are randomly ordered
- Determine number of hosts to run in parallel
  - `batch`: specify number of hosts
  - `batch_p`: specify number of hosts in percentage
  - `forks`: max number of concurrent processes
- Add option to display reports at end of tasks by using `--report` flag or specifying it in `spec` definition
  - `recap`: show basic report
  - `rc`: show return code for each host and task
  - `task`: show task status for each host and task
  - `time`: show time report for each host and task
  - `all`: show all reports
- Add confirm/step task capability
  - `confirm`: for the root task
  - `step`: per task and host

### Fixes

- Fix omitting attribute `align` when creating a theme
- Abort tasks prematurely when running in parallel and AnyErrorsFatal set to true
- Fix server range (previously `[2:100]` didn't work as strings were compared)
- Fix empty error for non existing working directory and update how work_dir works

### Minor

- Add option to omit empty columns via flag `--omit-empty-columns` and spec `omit_empty_columns`
- Add option to specify target and spec via flags `--target`/`--spec`
- Add description to targets and specs
- Add server identity to environment variables
- Add silent/describe attribute to spec definition
- Add ssh user flag option

### Changes

- [**BREAKING CHANGE**]: Deprecated the parallel flag, use batch/batch_p/forks instead
- [**BREAKING CHANGE**]: Rename default environment variables from `SAKE_SERVER_*` to `S_*`, and remove task default environment variables
- [**BREAKING CHANGE**]: Shorthand flag for silent is now `Q`
- Switch to default shell when evaluating inventory
- If no command name is set on nested tasks, assign `task-$i` instead of `task`
- If `--limit` flag is higher than available hosts, then select all hosts filtered
- Update flag sorting
- Rename `--omit-empty` to `--omit-empty-rows`
- Building `sake` with go 1.19

## 0.12.1

### Fixes

- Fix port out of range when using shorthand format for hosts

## 0.12.0

### Features

- Add hosts keyword that supports having multiple hosts per server definition
  - Specify as a list
  - Specify as a string containing range (`192.168.0.[1:10:2]`)
  - Use `inventory` attribute (`kubectl get nodes`)
- Add silent flag to supress `Running...` spinner when running tasks
- Support connection string instead of 3 fields: `user@host:port`
- Support resolving IdentityFile in ssh config (`~/.ssh/config`)
- Support resolving Includes in ssh config (`~/.ssh/config`)
- Support glob pattern for Hosts (`Host *`)
- Add bastion headers to list servers
- Add flags/target config `--limit` & `--limit-p` to limit number of servers task is run on
- Add filtering servers on host regex
- Add invert flag on filtering servers
- Add flag `--all-headers` for tasks and servers
- Add sub-commands edit/list/describe [specs|targets]
- Add 3 new table outputs (table-2, table-3, table-4)
- [BREAKING CHANGE]: Simplified theme config, now it only accepts manipulation of rows and headers, not specific properties

### Fixes

- Use IdentitiesOnly if user specifies a IdentityFile
- Default to `Name`, if description is not set, in auto-completion for tasks
- Support lowercase ssh config keys (previously they had to be PascalCase)

### Deprecated

- [BREAKING CHANGE]: Removed environment variables `SAKE_IDENTITY_FILE` and `SAKE_PASSWORD`, users can use flags instead

## 0.11.0

### Fixes

- Fix not being able to parse ssh config if match keyword found [35](https://github.com/alajmo/sake/pull/35)

### Features

- Support Bastion/jump host [32](https://github.com/alajmo/sake/pull/32)

## 0.10.3

### Fixes

- Previously known_hosts didn't work correctly when specifying port other than 22
- Fix authentication failures [32](https://github.com/alajmo/sake/pull/30)

### Features

- Resolve hosts from ssh_config
- Support hashing known_host entries

## 0.10.2

### Fixes

- Allow duplicate hosts
- Fix correct exit code on remote/local task errors (#27)
- Fix local WorkDir when it's not explicitly set

## 0.10.1

### Fixes

- Small fix for WorkDir being related to calling file when server is local

## 0.10.0

### Fixes

- Fix issue where ipv6 was not added correctly to known_hosts (brackets without ip)
- Fix TTY in sub-tasks
- Only task or cmd allowed in inline `tasks` definition

### Changes

- [BREAKING CHANGE]: Updated prefix handling in text output, now supports golang templating
  - Old config:
  ```yaml
  themes:
    default:
      text:
        header: true
        header_prefix: "TASK"
        header_char: "*"
  ```

  - New config:
  ```yaml
  themes:
    default:
      text:
        header: '{{ .Style "TASK" "bold" }} {{ .Name }}'
        header_filler: "*"
  ```
- WorkDir is now relative to the calling task for local commands, previously it was to the users `cwd`
- Remove debug flag

### Features

- Add sub-command `check` to check for configuration errors
- Add `shell` property to override the default shell

## 0.1.8

### Fixes

- Support ipv6 hosts (#13)
- Fix identity_file when set via config (#7)
- Don't apply work_dir when running local tasks

## 0.1.7

### Fixes

- Use uint16 for port (#4)

## 0.1.6

This is the initial release. Basic functionality is supported: running tasks over multiple remote servers and localhost.

- Add `known_hosts_file` flag/env/config setting and `disable_verify_host` config setting
- Add `identity`/`password` pair flag/config setting
- Add sub-command ssh to easily ssh into servers
- Add `tty`/`local`/`attach` config settings
- Support nested tasks and pass down environment variables
- Add flag/config setting `ignore-unreachable` flag for ignoring unreachable servers
- Add flag/config setting `any-errors-fatal` for stopping all tasks for all servers when error is encountered
- Add `work_dir` config setting for servers and tasks
