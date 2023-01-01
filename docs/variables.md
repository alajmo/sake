# Variables

sake supports setting variables for both servers and tasks. The variable can either be a string or a command (in which case it's encapsulated by `$()`) which will be evaluated (once) for each task.

```yaml
servers:
  webserver:
    host: 172.1.2.3
    env:
      string: hello world

tasks:
  ping:
    cmd: echo "$msg"
    env:
      msg: pong
      date: $(date)
```

Additionally, the following environment variables are available by default for all tasks:

- Server specific:
  - `S_NAME`
  - `S_HOST`
  - `S_USER`
  - `S_PORT`
  - `S_TAGS`

- Config specific:
  - `SAKE_DIR`
  - `SAKE_PATH`
  - `SAKE_KNOWN_HOSTS_FILE`

## Pass Variables from CLI

To pass variables from the CLI prompt, simply pass an argument, for instance:

```bash
sake run msg option=123
```

Now the environment variable `option` can be used in the task.

## Register Variables in Tasks

To access a previous tasks output, you can register a variable in the previous task, which will be available as an environment variable in the current task. In addition to just capturing output, the following environment variables will be available:

- `<name>_status`:
- `<name>_rc`:
- `<name>_failed`:
- `<name>_stdout`:
- `<name>_stderr`:

```yaml
tasks:
  ping:
    tasks:
      - cmd: echo "foo" && >&2 echo "error"
        register: out

      - cmd: |
          echo "status: $out_status"
          echo "rc: $out_rc"
          echo "failed: $out_failed"
          echo "stdout: $out_stdout"
          echo "stderr: $out_stderr"
          echo "out:"
          echo "$out"
```

Output:

```bash
$ sake run ping

TASKS ******************************

TASK (1/2) [task-0] ****************

172.24.2.2 | error
172.24.2.2 | foo

TASK (2/2) [task-1] ****************

172.24.2.2 | status: ok
172.24.2.2 | rc: 0
172.24.2.2 | failed: false
172.24.2.2 | stdout: foo
172.24.2.2 | stderr: error
172.24.2.2 | out:
172.24.2.2 | error
172.24.2.2 | foo
```
