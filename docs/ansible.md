# Ansible

Ansible is by far the most well-recognized software when it comes to configuring servers and/or running tasks and as such, warrants an in-depth comparison.

Ansible and sake overlap in one segment, namely ad-hoc task execution over remote hosts using ssh. They both allow you to write server configurations in YAML and target multiple hosts in parallel. There's some small cosmetic differences in how they accomplish that.

Where they notably differ is that Ansible has modules which allows you to easily write idempotent tasks for common use-cases , whereas with sake it's up to the developers to write tasks. Furthermore, Ansible is more opinionated than sake when it comes to structuring projects and is not as flexible. For instance, in Ansible roles require a specific structure. sake on the other hand is more bare-bone and opinionless.

## Comparison

|  | sake | Ansible
---|---|---
 Dependencies | Single Binary | Python runtime on control node and hosts
 Host Filtering | sake can filter on name, range, tags, and regex via config or CLI | Ansible can filter on name, group, range, regex via config or CLI
 Tasks | User writes tasks | Provides declarative idempotent modules
 Templating | No built-in templating | Jinja templates
 Auto-completion | Rich | Limited
 Config Format | YAML | YAML/Jinja + INI inventory
 Community | no | Big community and lots of resources
 Stable | no | yes
 Opionated | no | yes
 Learning curve | sake has less concepts to learn but puts more work on the user | While not complex, it's safe to say you'll be revisiting Ansible docs even after you've mastered it
 Documentation | yes | yes, has more stackoverflow posts, etc.
 Performance | [sake is around 4-8 times faster than Ansible depending on the number of hosts ](/performance) | Requires optimization

## Should I use Ansible or Sake?

If any of the following is true:

- you prefer or mostly rely on shell scripts
- you find that you're fighting with Jinja templates a bit too much
- you are tired of reading Ansible documentation
- you are using software which is already idempotent and prefer sakes simplicity

then sake can be an alternative to Ansible.

However, if you:

- are already familiar with Ansible, and don't find any of sakes features particularly useful
- want to use established software
- want to use idempotent and declarative configurations
- work in a big team and/or want industry relevant skills
- manage many and very complex infrastructure

then go with Ansible.

With that said, nothing is preventing you from using Ansible and sake side by side, perhaps using Ansible playbooks/roles for configuring your servers, and leveraging sakes performance for running ad-hoc tasks.

## Introduction

We'll begin with the problem both sake and Ansible are trying to solve and then compare how they try to solve it:

1. Connect to a machine
2. Install some services and configure them
3. Start a service
4. Query some data (uptime, disk size, etc.)
5. Upload/Download files

## What is Ansible

> Ansible is a radically simple IT automation system. It handles configuration management, application deployment, cloud provisioning, ad-hoc task execution, network automation, and multi-node orchestration

Ansible provides over 3000 ready-to-use modules, such as `shell`, `apt`, `file`, `copy`, etc. which you compose as playbooks and roles that you then execute for many hosts. The configuration files are written in a declarative DSL (YAML + Jinja templates).

Almost all Ansible modules are idempotent, meaning that if you run the same playbook multiple times, you will get the same results. Furthermore, Ansible is quite opinionated in terms of how you set up your roles:

> An Ansible role has a defined directory structure with eight main standard directories. You must include at least one of these directories in each role. You can omit any directories the role does not use
>
> [docs.ansible.com](https://docs.ansible.com/ansible/latest/user_guide/playbooks_reuse_roles.html#id2)

This makes sense for multiple reasons. First, it gives developers a common language to configure servers and removes yak-shaving to a degree because there are enforced decisions on how you're supposed to structure projects, and secondly, you can easily switch between projects that use Ansible.

There are some drawbacks though, the user has to become familiar with the opinionated and implicit file structure of Ansible, and there's a lot of boilerplate (more than 8 directories/files per role that you can use to customize roles: tasks, handlers, templates, files, vars, defaults, meta, module_utils, lookup_plugins, library, etc.).

## What is sake

> sake is a task runner for remote hosts

That's it basically. Its focus is the `ad-hoc task execution` segment that Ansible also does, but it tries to do it a bit better and more performant. You can use it to configure, deploy, and automate tasks, just like Ansible, but it's not as capable or robust as Ansible in that sense.

Sake provides only one module, namely the `shell` module. The `shell` module allows you to write commands in your preferred language that you then execute for many hosts.

Similar to Ansible, the configuration files in sake are written in YAML but without any support for Jinja templates. Instead, all the conditional logic is handled inside the task definitions.

So, as you can see, sake doesn't have as valiant a goal as Ansible, since to accomplish what Ansible has, you'd have to write over 3387 (and counting) modules! The benefit is that if you already have basic shell knowledge, you don't have to learn another software to set up your infrastructure, except of course for how to write tasks in sake, but that's quite simple.

Furthermore, there is no best-practice setup of how you organize your configuration files, that's all up to the user. The benefit is you don't have to try to shoehorn your project to fit a defined structure.

When it comes to idempotency, sake doesn't provide any guarantee that the tasks you write are idempotent (this is similar to Ansible if you use the `shell` module). Instead, that responsibility falls on the developer and the software you use. Fortunately, a lot of software is already idempotent by default:

- `docker-compose up` will not recreate your containers unless there is a newer image and you've set your Docker compose file to use `latest`.
- `rsync` won't upload files if there hasn't been any change
- `apt` won't reinstall your package if it's already installed
- `useradd` won't create a new user if it already exists

## Dependencies

Both sake and Ansible are agentless; they don't have to be installed on the remote hosts for them to work.

Sake is delivered as a single static library that uses the ssh protocol to execute tasks on remote nodes.

Ansible requires Python to be installed on the control node and utilizes the ssh protocol to transfer your playbooks to the remote node, where it then runs the playbooks (using python) and removes them after execution. Python also has to be installed on the remote nodes if you wish to utilize Ansible modules, otherwise, you can use the `shell` module which doesn't require Python to be installed.

## Inventory

Both sake and Ansible support creating a list of hosts that you can target. In Ansible, inventory files can be defined using multiple formats. The most common one is called `INI` and looks something like this:

```
mail.example.com

[webservers]
foo.example.com
bar.example.com

[dbservers]
one.example.com
two.example.com

[prod:children]
webservers
dbservers
```

Another format is YAML which looks like this:
```yaml
all:
  hosts:
    mail.example.com:
  children:
    webservers:
      hosts:
        foo.example.com:
        bar.example.com:
    dbservers:
      hosts:
        one.example.com:
        two.example.com:
    prod:
      hosts:
        foo.example.com:
        bar.example.com:
        one.example.com:
        two.example.com:
```

To list hosts you can run:

```sh
$ ansible all --list-hosts

  hosts (5):
    mail.example.com
    foo.example.com
    bar.example.com
    one.example.com
    two.example.com
```

sake on the other hand only has one file format for all configurations, a YAML configuration file:

```yaml
servers:
  mail:
    host: mail.example.com

  webservers:
    hosts:
      - foo.example.com
      - bar.example.com
    tags: [prod]

  dbservers:
    hosts:
      - one.example.com
      - two.example.com
    tags: [prod]
```

To list all hosts you can run:

```sh
$ sake list servers

 server       | host             | tags
--------------+------------------+------
 mail         | mail.example.com |
 webservers-0 | foo.example.com  | prod
 webservers-1 | bar.example.com  | prod
 dbservers-0  | one.example.com  | prod
 dbservers-1  | two.example.com  | prod
```

You can define dynamic inventories in both Ansible and sake.

The only difference between sake and Ansible is how hosts belonging to multiple groups work. In Ansible you have to create a new group and specify the hosts belonging to it, whereas in sake you simply add a tag to hosts.

## Modules and Idempotency

A big part of Ansible, and what makes it popular, is that it provides idempotent ready-to-use modules that you can use to describe your wanted configuration state in a declarative way.

For instance, to install `htop` in Ansible, you could do something like this (using the builtin `apt` module):

```yaml
- name: install
  become: true
  tasks:
    - name: apt
      apt:
        name: htop
        state: present
```

or this (using Ansibles `shell` module):

```yaml
- name: install
  tasks:
    - shell: sudo apt-get install --no-upgrade htop -y
```

In sake you do this:

```yaml
tasks:
  install:
    cmd: sudo apt-get install htop --no-upgrade -y
```

The difference is that in Ansible, the first definition is declarative (you declare the wanted state, and Ansible takes care of the steps to get there), whereas, in the second and third definitions, you write the steps yourself to get to your desired state.

The thing is, `apt-get install --no-upgrade` is already idempotent, if the package is installed, it won't do anything, if it's not, then it will install the package. So if you already know the syntax for `apt`, there's little benefit to this particular Ansible module, other than the consistency of always using Ansible modules where possible.  In fact, The `shell` method is faster than the built-in `apt` module. I suspect the `apt` module is slower because when you run tasks in Ansible, it first has to copy over the playbook to the host and then run it (via Python), whereas, with the `shell` module, it's running the command directly.

Both methods have their pros and cons, on one hand, when you're wrapping software (as Ansible does), you provide the users with a nice looking DSL and cover most basic functionality, but on the other hand, it's up to Ansible to update their wrappers when the underlying software changes. For `apt` that is stable and I assume hasn't changed over the years, it's fine.  But when you have software like `rsync`, which has over 40 flags for modifying its behavior, it's not as pragmatic. Ansible has a `rsync` module but it doesn't wrap around all of the `rsync` flags that are available.

## Task Output

Ansibles task output is optimized for showing changes made when you run tasks, for instance:

```bash
$ ansible-playbook web_playbook.yaml -i hosts

PLAY [Setup] *******************************************************************************************************

TASK [ping] ********************************************************************************************************
changed: [172.24.2.9]

PLAY RECAP *********************************************************************************************************
172.24.2.9                 : ok=1    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0
```

To get the `pong` output, there are multiple ways:

1. Add `stdout_callback = minimal` to Ansible config
2. Add debug task and register out variable

There are probably other ways, but I think it shows that it's quite bothersome to execute singular tasks and show the output of the task without resorting to editing the configs.

Sake on the other hand optimizes for showing task output. There are multiple formats (text, tables, markdown, CSV, JSON, YAML) and it's easy to change the output format for a task inline or at the command line:

```bash
$ sake run ping

TASK [ping] ***************

172.24.2.9 | pong

$ sake run ping --output table

 Server | Ping
--------+------
 kaka   | pong

$ sake run ping --output markdown

| server | ping |
|:--- |:--- |
| kaka | pong |
```

## Host Filtering

Both sake and Ansible support host filtering at the task level. For instance, in Ansible, you'd write:

```yaml
- name: Ping
  hosts: prod
  tasks:
    - name: ping
      shell: echo pong
```

to target all `prod` hosts.

In sake it's:

```yaml
tasks:
  ping:
    target:
      tags: [prod]
    cmd: echo pong
```

If you want to override the hosts from the command line you can write:

Ansible:
```sh
ansible-playbook playbook.yml -l <host>
```

sake:
```sh
sake run ping -t dev
```

The big difference is that with sake you get auto-completion for all the hosts and tags, whereas in Ansible you have to know the group or hosts beforehand.

## Directory Structure

In Ansible you define hosts, playbooks, and optionally, roles.

We've covered hosts already, but let's look at playbooks and roles.

Playbooks are simply a collection of tasks that you define to run for a set of hosts.

Roles are a collection of tasks, files, templates, handlers, and a bunch of other stuff, which you can define to improve reusability.

What you normally do is define one or multiple playbooks that are entry points to configuring servers. The playbook then refers to a set of roles, that further define the tasks, handlers, files, etc. that you use to configure your server.

The roles have an opinionated and strict directory structure that you must adhere to (for instance, the `tasks` directory must have a `main.yaml`file for it to be picked up).

One limitation with Ansible tasks is that you can only import task files, not individual tasks from a file. What you can do instead is define a tag for that specific task and then run the playbook with the `--tags` flag. I'm not sure why Ansible doesn't allow you to import specific tasks only. Also, be aware of cyclic dependencies, it seems Ansible doesn't detect them so you can an infinitive loop.

sake on the other hand is like any other programming language, you import config files and can then reference specific tasks you intend to use. So you're free to structure your project any way you see fit.

## IDE & Auto-completion

Ansible has basic auto-completion for commands and flags, but it doesn't seem to be able to autocomplete single tasks, groups, tags, etc.

There's also autocompletion for running modules ad-hoc on the command line, however, when you run `ansible -m <TAB>`, you get a list of all modules, which is over 3000, so the initial load time is substantial.

sake supports rich autocompletion for both commands, flags, and values for hosts, tasks, tags, etc.

For instance, when I type `sake run <TAB> --server <TAB> --tags <TAB>` I get the following autocompletions:

- first `<TAB>` autocompletes all available tasks
- second `<TAB>` autocompletes all available hosts
- third `<TAB>` autocompletes all available tags

Ansible has IDE integrations for vscode, vim, and other IDEs. It seems to support key attributes (gather_facts, etc.) but doesn't provide completion for hosts or tasks.

Sake doesn't have any IDE integrations yet, but I plan on adding them at a later stage when sake reaches maturity and the API has stabilized.

## Stability & Community

Ansible has a big community and is considered stable (reached v1).

Sake is still in early development and the community is non-existent.

## Misc

Sake has a list of quality-of-life features that I've not found in Ansible:

- Edit the `sake` config file via the `sake edit` command, which opens up the config file in your preferred editor, this also works for servers, tasks, targets and specs
- `sake ssh <TAB>` gives me a list of all the hosts and the ability to ssh into them
- `sake run <task> --edit` opens up the editor and navigates to the line where the task is defined, this is quite helpful when you're debugging a task
- `sake describe/list servers --tags web` lists/describes all the servers that have the tag `web`

## Performance

Ansible is not known for its performance, in fact, just performing a simple ping to a single host over LAN takes over a second, whereas sake only takes 200 ms. And it gets worse when you increase the number of hosts, for 100 hosts it took Ansible around 3.5 seconds to complete (Pipelining enabled, gather facts off and forks set to 10), and for sake only 350 ms.
Additionally, I found sake to consume less memory and CPU.

Some more extensive benchmarks can be found [here](/performance).
