# Roadmap

`sake` is under active development. Before v1.0.0, I want to finish the following tasks:

- [] Improve servers
  - [x] Resolve hostnames from ssh_config
  - [x] Support Bastion/Jumphost
  - [] Define multiple hosts without creating individual servers
  - [] Dynamically fetch hosts
  - [] Regex filtering of servers
- [] Improve tasks

- [] Improve tasks
  - [] Return correct error exit codes when running tasks
    - [x] serial playbook
    - [] parallel playbook
  - [] Omit certain tasks from auto-completion and/or being called directly (mainly tasks which are called by other tasks)
  - [] Repress certain task output if exit code is 0, otherwise displayed
  - [] Summary of task execution at the end
  - [] Pass environment variables between tasks
  - [] Access the previous exit code of the previous task
  - [] Conditional task execution
  - [] Tags/servers filtering launching different comands on different servers #6
  - [] Ensure command (check existence of file, software, etc.)
  - [] on-error/on-success task execution
  - [] Cleanup task that is always ran
  - [] Log task execution to a file
  - [] Abort if certain env variables are not present (required envs)
  - [] Add --step mode flag or config setting to prompt before executing a task

After v1.0.0, focus will be directed to implementing a `tui` for `sake`.
