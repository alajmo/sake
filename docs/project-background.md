# Project Background

> To configure servers, all you need is bash scripting, environment variables, and some know-how.

`sake` came about because I needed a simple tool to manage my servers. There's tons of software in this category, Ansible, Puppet, Chef, Salt, Sup, and probably many more. However, most of them are geared towards enterprise server management and are often not the most ergonomic. So, `sake` was born, an ergonomic CLI tool to configure servers.

## Premise

The premise is you have a bunch of servers and want the following:

1. A central place for your servers, containing name, host, and a small description of the servers
2. Ability to run ad-hoc commands (perhaps `hostnamectl` to see system hostname) on 1, a subset, or all of the servers
3. Ability to run defined tasks on 1, a subset, or all of the servers
4. Ability to get an overview of 1, a subset, or all of the servers and tasks

## Design

`sake` prioritizes simplicity and flexibility, principles that are at odds with each other at times:

1. Simplicity - both the implementation and the UX is designed to be simple. However, this rule can be bent to some degree if the reward is substantial
2. Flexibility - provide the user with the ability to shape `sake` to their user case, instead of providing an opinionated setup

With these principles in mind, I've elected not to create a complex DSL (see Terraform, Puppet, Ansible, etc.), but instead just add a few primitives and let the user leverage their existing sysadmin knowledge to create their ideal setup. This results in not forcing users to learn yet another DSL, avoiding continuous lookup of `sake` commands, and updating `sake` to get new features.

It does, however, push complexity onto the user, for instance, there's no built-in primitive to download files and the user must define a task to do so. It would be foolish for me to aim for feature parity with 3rd party software like rsync for downloading files (there are over 150 flags for rsync).
In this sense, `sake` follows the Unix principle:

- Write programs that do one thing and do it well - run commands on multiple servers
- Write programs to work together - invoke other programs to do your bidding (rsync)
- Write programs to handle text streams - output of all `sake` commands is just text

## Comparisons

A lot of the alternatives to `sake` are meant to be used in large teams, which often results in:

1. Opinionated and implicit file structure that the user is required to know
2. Some of them are not daemon-less, there's a background process keeping track of application state (see Kubernetes, Chef, Puppet), which increases complexity and debuggability
3. Lots of boilerplate (with Ansible there are more than 8 directories/files per role that you can use to customize roles: tasks, handlers, templates, files, vars, defaults, meta, module_utils, lookup_plugins, library, etc.)

For DIY hobby projects, where you're only interested in:

1. Connecting to a machine
2. Installing some software
3. Starting a service
4. Querying some data (uptime, disk size, etc.)
5. Uploading/Downloading files

the alternatives can be overkill.

In terms of features, `sake` mostly resembles [Sup](https://github.com/pressly/sup), an excellent deployment tool, which `sake` has drawn inspiration from. There are however some notable differences:

- `sake` has a more extensive and easy to use filtering system
- `sake` provides both text and table output
- `sake` has auto-completion
- `sake` has a sub-command to easily ssh into servers
- `sake` supports tasks
- `sake` has native import capability (in Sup you're left with manually invoking the Sup binary, which is by design as they want to keep it as close to `make` as possible it seems)
- `sup` is not maintained anymore
- Better CLI ergonomics in my opinion, `sake run <task> -s server-1` versus `sup <network> <task>` (I often forget which one comes first, network or task)

### User Experience

These features make using `sake` feel more effortless:

- Single binary
- Rich auto-completion
- Edit the `sake` config file via the `sake edit` command, which opens up the config file in your preferred editor
- Multiple output formats (text, table, HTML, Markdown)
- Default target filtering for tasks

## Similar Software

- [Ansible](https://www.ansible.com)
- [Chef](https://www.chef.io)
- [Puppet](https://puppet.com)
- [Sup](https://pressly.github.io/sup)
