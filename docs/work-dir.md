# Working Directory

## Change Working Directory

You can change the default `work_dir` in the server section and the task section (nested tasks/commands included).
The order of precedence is as follows:

1. task list
2. task
3. referenced task
4. server
5. default, which is
    - the executed task directory for local clients
    - the users home directory for remote clients

```yaml
servers:
  localhost:
    host: localhost
    work_dir: "/opt" # 4
    local: true

tasks:
  work-ref:
    name: pwd
    work_dir: "/usr" # 3
    cmd: pwd

  work-dir:
    work_dir: "/home" # 2
    tasks:
      - task: work-ref

      - cmd: pwd
        name: pwd

      - cmd: pwd
        name: pwd
        work_dir: "/" # 1
```

See example:

```bash
$ sake run work-dir --output table

 Server    | Pwd   | Pwd   | Pwd
-----------+-------+-------+-----
 localhost | /home | /home | /

# if we comment work_dir (# 2) then we get

 Server    | Pwd  | Pwd  | Pwd
-----------+------+------+-----
 localhost | /usr | /opt | /
```

The complete decision tree for composing a working directory is found below.

Note:
  - Absolute directories (`/opt`) won't be joined.
  - The variables `Task Context` and `Server Context` are the local directories where the tasks/servers are defined.

## Remote Tasks

Resolve `work_dir` according to `Server Dir` and `Task Dir`:


| Host   | Task   | Server Dir | Task Dir | work_dir                  |
|--------|--------|------------|----------|---------------------------|
| remote | remote | ""         | ""       | `/home/user`              |
| remote | remote | ""         | "task"   | `/home/user/task`         |
| remote | remote | "server"   | ""       | `/home/user/server`       |
| remote | remote | "server"   | "task"   | `/home/user/server/task`  |

## Local Tasks

Resolve `work_dir` according to `Task Context`:

| Host    | Task   | Server Dir | Task Dir | work_dir                  |
|---------|--------|------------|----------|---------------------------|
| local   | local  | ""         | ""       | `[Task Context]`          |
| local   | remote | ""         | ""       | `[Task Context]`          |
| remote  | local  | ""         | ""       | `[Task Context]`          |
| remote  | local  | "server"   | ""       | `[Task Context]`          |

Resolve `work_dir` according to `Task Context` and `Task Dir`:

| Host   | Task   | Server Dir | Task Dir  | work_dir                  |
|--------|--------|------------|-----------|---------------------------|
| local  | local  | ""         | "task"    | `[Task Context]/task`    |
| local  | remote | ""         | "task"    | `[Task Context]/task`    |
| remote | local  | ""         | "task"    | `[Task Context]/task`    |
| remote | local  | "server"   | "task"    | `[Task Context]/task`    |

Resolve `work_dir` according to `Server Context` and `Server Dir`:

| Host   | Task   | Server Dir | Task Dir | work_dir                  |
|--------|--------|------------|----------|---------------------------|
| local  | remote | "server"   | ""       | `[Server Context]/server` |
| local  | local  | "server"   | ""       | `[Server Context]/server` |

Resolve `work_dir` according to `Server Context`, `Server Dir` and `Task Dir`:

| Host   | Task   | Server Dir | Task Dir  | work_dir                       |
|--------|--------|------------|-----------|--------------------------------|
| local  | local  | "server"   | "task"    | `[Server Context]/server/cmd` |
| local  | local  | "server"   | "task"    | `[Server Context]/server/cmd` |
