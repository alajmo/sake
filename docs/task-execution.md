# Task Execution

Sake offers multiple execution strategies that controls how tasks are executed across the hosts. The available ones are:

- **linear**: execute task for each host before proceeding to the next task (default)
- **host_pinned**: executes tasks (serial) for a host before proceeding to the next host
- **free**: executes tasks without waiting for other tasks

You can set the strategy via the `stragegy` property in a `spec` definition or via a flag `--strategy [option]`.

Additionally, the following properties are available to further control task execution:

- **batch**: number of hosts to run in parallel
- **batch_p**: percentage of hosts to run in parallel
- **forks**: max number of concurrent processes

## Linear Strategy

When the following properties are set:

- **strategy**: linear
- **batch**: 2
- **forks**: 10000

Sake will execute according to the following image:

![linear](/img/linear-strategy.png)

1. Task `T1` will run in parallel for hosts `H1` and `H2`
2. Task `T1` will run for host `H3`
3. Task `T2` will run in parallel for hosts `H1` and `H2`
4. Task `T2` will run for host `H3`
5. Task `T3` will run in parallel for hosts `H1` and `H2`
6. Task `T3` will run for host `H3`

## Host Pinned Strategy

When the following properties are set:

- **strategy**: host_pinned
- **batch**: 2
- **forks**: 10000

Sake will execute according to the following image:

![linear](/img/host_pinned-strategy.png)

1. Tasks `T1 - T3` will run serially for `H1` and `H2` (in parallel)
2. Tasks `T1 - T3` will run serially for `H3`

## Free Strategy

When the following properties are set:

- **strategy**: free
- **batch_p**: 100%
- **forks**: 10000

Sake will execute according to the following image:

![linear](/img/free-strategy.png)

1. All tasks for all hosts will run in parallel

## Ordering Hosts

There are multiple host ordering options available:

- **inventory**: the order is as provided by the inventory
- **reverse_inventory**: the order is the reverse of the inventory
- **sorted**: hosts are alphabetically sorted by host
- **reverse_sorted**: hosts are sorted by host in reverse alphabetical order
- **random**: hosts are randomly ordered

## Confirm Before Running Tasks

sake comes with two options to confirm tasks before execution:

- **confirm**: this property is used when you want to simply confirm the task you invoked before running
  - it can be specified either via flag `--confirm` or the spec property `confirm`
- **step**: this property is used when you want to confirm each task individually
  - it can be specified either via flag `--step` or the spec property `step`
    - Invoking `--step` will provide the user 3 options:
      - `yes`: run the task
      - `no`: skip the task
      - `continue`: run the task and don't prompt for the next tasks

