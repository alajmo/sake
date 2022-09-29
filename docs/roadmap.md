# Roadmap

`sake` is under active development. Before **v1.0.0**, I want to finish the following tasks, some miscellaneous fixes and improve code documentation:

- [x] Improve servers
  - [x] Resolve hostnames from ssh_config
  - [x] Support Bastion/Jumphost
  - [x] Define multiple hosts without creating individual servers
  - [x] Dynamically fetch hosts
  - [x] Regex filtering of servers
  - [x] Support glob pattern for Hosts (`Host *`)
  - [x] Support resolving Includes in ssh config (`~/.ssh/config`)
  - [x] Add limit and limit-p flag/target
  - [x] Add filtering servers on host regex
  - [x] Add invert flag on filtering servers

- [ ] Improve output
  - Add new table format output (tasks in 1st column, output in 2nd, one table per server)
  - Add new format output (tasks in column, project output in row)

- [ ] Improve tasks
  - [ ] Return correct error exit codes when running tasks
    - [x] serial playbook
    - [ ] parallel playbook
  - [ ] Omit certain tasks from auto-completion and/or being called directly (mainly tasks which are called by other tasks)
  - [ ] Repress certain task output if exit code is 0, otherwise displayed
  - [ ] Summary of task execution at the end
  - [ ] Pass environment variables between tasks
  - [ ] Access exit code of the previous task
  - [ ] Conditional task execution
  - [ ] Tags/servers filtering launching different comands on different servers #6
  - [ ] Ensure command (check existence of file, software, etc.)
  - [ ] on-error/on-success task execution
  - [ ] Cleanup task that is always ran
  - [ ] Log task execution to a file
  - [ ] Abort if certain env variables are not present (required envs)
  - [ ] Add --step mode flag or config setting to prompt before executing a task
  - [ ] Add yaml to command mapper

## Future

After **v1.0.0**, focus will be directed to implementing a `tui` for `sake`. The idea is to create something similar to `k9s`, where you have can peruse your servers, tasks, and tags via a `tui`, and execute tasks for selected servers, ssh into servers, etc.
