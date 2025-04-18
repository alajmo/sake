# Output

`sake` supports different output formats for tasks. By default it will use `text` output, but it's possible to change this via the `--output` flag or specify it in the task `spec`.

The following output formats are available:

- **text** (default), use this when you want streamed output to terminal
  ```
  TASK (1/2) [task-0] ***********

  172.24.2.2 | ping
  172.24.2.2 | ping

  TASK (2/2) [task-1] ***********

  172.24.2.2 | pong
  172.24.2.2 | pong
  ```
- **table**, useful when you have many hosts but few tasks
  ```
   Host       | Task-0 | Task-1
  ------------+--------+--------
   172.24.2.2 | ping   | pong
  ------------+--------+--------
   172.24.2.2 | ping   | pong
  ```
- **table-2**, useful when you have many tasks but few hosts
  ```
   Task   | 172.24.2.2 | 172.24.2.2
  --------+------------+------------
   task-0 | ping       | ping
  --------+------------+------------
   task-1 | pong       | pong
  ```
- **table-3**, useful when you want separate tables per host
  ```
      172.24.2.2

   Task-0 | Task-1
  --------+--------
   ping   | pong

      172.24.2.2

   Task-0 | Task-1
  --------+--------
   ping   | pong
  ```
- **table-4**, useful when you have many hosts and many tasks
  ```
   Task   | 172.24.2.2
  --------+------------
   task-0 | ping
  --------+------------
   task-1 | pong

   Task   | 172.24.2.2
  --------+------------
   task-0 | ping
  --------+------------
   task-1 | pong
  ```
- **html**
  ```html
  <table class="">
    <thead>
    <tr>
      <th align="left">host</th>
      <th align="left">task-0</th>
      <th align="left">task-1</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td align="left">172.24.2.2</td>
      <td align="left">ping</td>
      <td align="left">pong</td>
    </tr>
    <tr>
      <td align="left">172.24.2.2</td>
      <td align="left">ping</td>
      <td align="left">pong</td>
    </tr>
    </tbody>
  </table>
  ```
- **markdown**
  ```markdown
  | host | task-0 | task-1 |
  |:--- |:--- |:--- |
  | 172.24.2.2 | ping | pong |
  | 172.24.2.2 | ping | pong |
  |  |  |  |
  ```
- **json**
  ```json
  [
    {
      "host": "172.24.2.2",
      "task-0": "ping",
      "task-1": "pong"
    },
    {
      "host": "172.24.2.2",
      "task-0": "ping",
      "task-1": "pong"
    }
  ]
  ```
- **csv**
  ```csv
  host,task-0,task-1
  172.24.2.2,ping,pong
  172.24.2.2,ping,pong
  ```
- none

## Omit Empty Table Rows and Columns

If you wish to omit rows/columns that return empty outputs, you can do so via the `--omit-empty-rows`/`--omit-empty-columns` flag or specify it in the task `spec`. Note, this only works for the tables, json, csv, markdown, and html.

See below for an example:

```bash
$ sake run empty -s server-3,server-4 -o table

TASKS *******************************

 Host       | Task-0 | Task-1
------------+--------+--------
 172.24.2.4 | 123    |
------------+--------+--------
 172.24.2.5 |        |

$ sake run empty -s server-3,server-4 -o table --omit-empty-rows --omit-empty-columns

TASKS *******************************

 Host       | Task-0
------------+--------
 172.24.2.4 | 123
```

## Print Reports

`sake` comes with a few reports that gives you an overview of task execution:

The available reports are:

- **recap**: show basic report (default)
- **rc**: show return code for each host and task
- **task**: show task status for each host and task
- **time**: show time report for each host and task
- **all**: show available reports

```bash
$ sake run task --report=all

TASKS **************************************************************************************

 Host       | Task-0                       | Task-1                       | Task-2
------------+------------------------------+------------------------------+--------
 172.24.2.2 | foo                          | bar                          | xyz
------------+------------------------------+------------------------------+--------
 172.24.2.2 |                              |                              |
            | Process exited with status 1 | Process exited with status 1 |

RETURN CODES *******************************************************************************

 host            task-0  task-1  task-2
----------------------------------------
 172.24.2.2      0       0       0
 172.24.2.2      1       1

TASK STATUS ********************************************************************************

 host            task-0   task-1  task-2
------------------------------------------
 172.24.2.2      ok       ok      ok
 172.24.2.2      ignored  failed  skipped

TIME ***************************************************************************************

 host            task-0  task-1  task-2  Total
------------------------------------------------
 172.24.2.2      0.09 s  0.01 s  0.01 s  0.10 s
 172.24.2.2      0.08 s  0.01 s          0.09 s
------------------------------------------------
 Total           0.17 s  0.02 s  0.01 s  0.20 s

RECAP **************************************************************************************

 172.24.2.2      ok=3  unreachable=0  ignored=0  failed=0  skipped=0
 172.24.2.2      ok=0  unreachable=0  ignored=1  failed=1  skipped=1
---------------------------------------------------------------------
 Total           ok=3  unreachable=0  ignored=1  failed=1  skipped=1
```

