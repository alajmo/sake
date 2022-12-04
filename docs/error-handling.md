# Error Handling

`sake` has multiple ways to deal with errors, and depending on the task execution strategy and which error flags are used, you will get different behavior. The following properties can be used to control task execution and how errors are handled:

- **any_errors_fatal**: stop task execution on all servers on error, this is the same as setting `max_fail_percentage` to zero
  - note that when you run tasks in parallel, it will wait for the current tasks to finish before aborting
- **max_fail_percentage**: stop task execution on all servers when threshold reached
- **ignore_errors**: continue task execution on errors
- **ignore_unreachable**: ignore unreachable hosts

## Aborting on the First Error

If you wish to abort all tasks on all errors in case an error is encountered for any task, use the flag `--any-errors-fatal` or specify it in the task `spec`.

- `any-errors-fatal` set to false

  ```bash
  $ sake run fatal --any-errors-fatal=false

  TASKS *******************************************************************

   Host       | Task-0 | Task-1                       | Task-2
  ------------+--------+------------------------------+--------
   172.24.2.2 | 123    |                              |
              |        | Process exited with status 1 |
  ------------+--------+------------------------------+--------
   172.24.2.2 | 123    |                              |
              |        | Process exited with status 1 |

  RECAP *******************************************************************

   172.24.2.2      ok=1  unreachable=0  ignored=0  failed=1  skipped=1
   172.24.2.2      ok=1  unreachable=0  ignored=0  failed=1  skipped=1
  ---------------------------------------------------------------------
   Total           ok=2  unreachable=0  ignored=0  failed=2  skipped=2
   ```

- `any-errors-fatal` set to true
  ```bash
  $ sake run fatal --any-errors-fatal=true

  TASKS ******************************************************************

   Host       | Task-0 | Task-1                       | Task-2
  ------------+--------+------------------------------+--------
   172.24.2.2 | 123    |                              |
              |        | Process exited with status 1 |
  ------------+--------+------------------------------+--------
   172.24.2.2 |        |                              |

  RECAP ******************************************************************

   172.24.2.2      ok=1  unreachable=0  ignored=0  failed=1  skipped=1
   172.24.2.2      ok=0  unreachable=0  ignored=0  failed=0  skipped=3
  ---------------------------------------------------------------------
   Total           ok=1  unreachable=0  ignored=0  failed=1  skipped=4
  ```

## Ignoring Task Errors

If you wish to continue task execution even if an error is encountered, use the flag `--ignore-errors` or specify it in the task `spec`.

- `ignore-errors` set to false
  ```bash
  $ sake run errors --ignore-errors=false

  TASKS ******************************************************************

   Host       | Task-0 | Task-1                       | Task-2
  ------------+--------+------------------------------+--------
   172.24.2.2 | 123    |                              |
              |        | Process exited with status 1 |
  ------------+--------+------------------------------+--------
   172.24.2.2 | 123    |                              |
              |        | Process exited with status 1 |

  RECAP ******************************************************************

   172.24.2.2      ok=1  unreachable=0  ignored=0  failed=1  skipped=1
   172.24.2.2      ok=1  unreachable=0  ignored=0  failed=1  skipped=1
  ---------------------------------------------------------------------
   Total           ok=2  unreachable=0  ignored=0  failed=2  skipped=2
   ```

- `ignore-errors` set to true
  ```bash
  $ sake run errors --ignore-errors=true

  TASKS ********************************************************************

   Host       | Task-0 | Task-1                       | Task-2
  ------------+--------+------------------------------+--------
   172.24.2.2 | 123    |                              | 321
              |        | Process exited with status 1 |
  ------------+--------+------------------------------+--------
   172.24.2.2 | 123    |                              | 321
              |        | Process exited with status 1 |

  RECAP ********************************************************************

   172.24.2.2      ok=2  unreachable=0  ignored=1  failed=0  skipped=0
   172.24.2.2      ok=2  unreachable=0  ignored=1  failed=0  skipped=0
  ---------------------------------------------------------------------
   Total           ok=4  unreachable=0  ignored=2  failed=0  skipped=0
  ```

## Ignoring Unreachable Hosts

Sometimes you want to ignore remote hosts which are unreachable, for instance if it's a host that is flaky, then you can either use the `--ignore-unreachable` flag or specify it in the task `spec`.

- `ignore-unreachable` set to false
  ```bash
  $ sake run unreachable --ignore-unreachable=false

                                       Unreachable Hosts

   Server | Host         | User | Port | Error
  --------+--------------+------+------+-----------------------------------------------------
   list-1 | 172.24.2.222 | test | 22   | dial tcp 172.24.2.222:22: connect: no route to host
  ```

- `ignore-unreachable` set to false
  ```bash
  $ sake run unreachable --ignore-unreachable=true

                                       Unreachable Hosts

   Server | Host         | User | Port | Error
  --------+--------------+------+------+-----------------------------------------------------
   list-1 | 172.24.2.222 | test | 22   | dial tcp 172.24.2.222:22: connect: no route to host

  TASKS **************************************************************************************

   Host       | Task-0
  ------------+--------
   172.24.2.2 | 123

  RECAP **************************************************************************************

   172.24.2.2        ok=1  unreachable=0  ignored=0  failed=0  skipped=0
   172.24.2.222      ok=0  unreachable=1  ignored=0  failed=0  skipped=0
  -----------------------------------------------------------------------
   Total             ok=1  unreachable=1  ignored=0  failed=0  skipped=0

  ```
