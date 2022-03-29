# Project Background

This document contains a little bit of everything:

- Background to `sake` and core design decisions used to develop `sake`
- Comparisons with alternatives

It's a simplified Ansible ment to work with homelabs / hobby projects where you're not really interested in reading documentation to simply configure a server.

## Background

`sake` came about because I needed a simple tool to manage multiple servers. So, the premise is, you have a bunch of servers and want the following:

1. a central place for your servers, containing name, host, and a small description of the servers
2. ability to run ad-hoc commands (perhaps `hostnamectl` to see system hostname) on 1, a subset, or all of the servers
3. ability to run defined tasks on 1, a subset, or all of the servers
4. ability to get an overview of 1, a subset, or all of the servers and tasks

Now, there's plenty of existing software to do this, and while I've taken a lot of inspiration from them, there's some core design decision that led me to create `sake`, instead of forking or contributing to an existing solution.

## Design

`sake` prioritizes the following design principles:

1. Simple - both the implementation and the UX is designed to be simple first and foremost
2. Flexible - provide the user with the ability to shape `sake` to their user cases, instead of providing an opionated setups

### Pillars

`sake` prioritizes simplicity and customizability. It's not really opinionated as other alternatives are, it's up to you how you want to organize your scripts, all `sake` does is provide you with some primitives, **servers** and **tasks**, and leaves it up to the user to how they want to organize their setup.

`sake` is not about creating a complex DSL (see terraform, puppet, ansible, etc.), but instead let the user use their existing sysadm knowledge to manage servers, meaning, there should be minimal lookup of `sake` commands, etc. In reality, to setup simple servers, all you need is some bash scripting and environment variables.

`sake` doesn't include **ssh**, **scp**, etc. but leverages the users and allows the user to override such commands. For instance, if you want to use rsync to download/upload files, then you can override the default behavior to always use that. This has the added benefit of users not requiring to keep up with `sake` releases and avoid minimal patching of said services. Also because `ssh` and other softwares are established softwares. And there's so many different upload/download options available, that it would be foolish for myself to aim for feature parity, even for basic stuff as choosing a different auth protocol other than ssh.

`sake` is supposed to be a thin client, all useful stuff will come in the form of bash scripts, called recipes, which the user can download and customize, meaning there's no need to keep updating `sake` for feature parity.

`sake` supports easy editting of tasks as well, for instance, `sake run <task> --edit` or `sake edit task <task>`, instead of perusing a bunch of files

### Comparisons

A lot of the alternatives to `sake` are meant to be used in large teams with many servers, which often results in:

1. Opionated and implicit file structure that the user is required to know
2. Some of them are not daemon-less, meaning there's a background process keeping track of application state (see K8s, Chef, Puppet), which increases complexity
3. Lot's of boilerplate (with Ansible there's more than 8 directories/files per role that you can use to customize roles: tasks, handlers, templates, files, vars, defaults, meta, module_utils, lookup_plugins, library, etc.)
4. Having to provide overrides for different types of machines (CentOS, Ubuntu, etc.), in Ansible, this results in ugly conditional YAML using jinja templates

For DIY hobby projects, where you're only interested in:

1. Connecting to a machine
2. Installing some software
3. Starting a service
4. Querying some data (uptime,disk size, etc.)
5. Uploading/Downloading files

the alternatives are simply overkill.

In terms of features, `sake` mostly resembles [Sup](https://github.com/pressly/sup), an excellent deployment tool, which `sake` has drawn inspiration from. There are however some notable differences:

- `sake` has a more extensive and easy to use filtering system
- `sake` provides both text and table output
- `sake` provides auto-completion for servers and tasks
- `sake` provides `ssh` and `ssh-tunnel` commands for quickly ssh'ing to a server and setting up ssh-tunnels
- A task can run multiple commands (in Sup, you're left with manually invoking the Sup binary)
- You can import other `sake` configs natively (in Sup you're left with manually invoking the Sup binary)
- Better CLI ergonomics in my opinion, `sake run <task> -s server-1` versus `sup <network> <task>` (I often forget which one comes first, network or task)

### General UX

These various features make using `sake` feel more effortless:

- Rich auto-completion
- Edit the `sake` config file via the `sake edit` command, which opens up the config file in your preferred editor
- Single binary (most alternatives require Python)
- Pretty output when running tasks or listing servers/tasks
- Default tags/name filtering for tasks
- Export output as HTML/Markdown from list/run/exec commands

## Similar Software

- [Ansible](https://www.ansible.com)
- [Chef](https://www.chef.io)
- [Puppet](https://puppet.com)
- [Kubernetes](https://puppet.com)
- [Sup](https://pressly.github.io/sup)
