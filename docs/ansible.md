# Ansible

Ansible is by far the most well-recognized software when it comes to configuring servers and/or running tasks and as such, warrants an in-depth comparison.

Ansible and sake overlap in one segment, namely ad-hoc task execution over remote hosts using ssh. They both allow you to write commands in YAML and target multiple hosts in parallel.

Where they notably differ is that Ansible has modules which allows you to easily write idempotent tasks for common use-cases , whereas with sake it's up to the developers to write tasks.

## Comparison Table

|  | sake | Ansible
---|---|---
 Dependencies | Single Portable Binary | Python runtime on control node and hosts
 Host Filtering | sake can filter on name, range, tags, and regex via config or CLI | Ansible can filter on name, group, range, regex via config or CLI
 Tasks | User writes tasks | Provides declarative idempotent modules
 Real-time output | Yes | No, however, there's some workarounds with async actions/polling
 Templating | No built-in templating | Jinja templates
 Auto-completion | Rich | Limited
 Config Format | YAML, same file for sake config, tasks, and inventory | YAML/Jinja + INI inventory, separate files for Ansible config, tasks, and inventory
 Community | Non-existent community | Big community
 Stable | No, and 1 core developer | Yes, and a lot of core developers
 Opionated | No | Yes
 Learning curve | sake has less concepts to learn but puts more work on the user | While not complex, it's safe to say you'll be revisiting Ansible docs
 Documentation | Yes | Yes, Ansible has a lot more Stackoverflow posts, blogs, etc.
 Performance | [sake is around 4-18 times faster than Ansible depending on the number of hosts ](/performance) | Lacking, however there's a lot of optimization configurations you can do to increase performance

## Should I use Ansible or Sake?

You can use Ansible and sake side by side, perhaps using Ansible playbooks/roles for configuring your servers, and leveraging sakes performance for running ad-hoc tasks.

That said, if you wish to use only one:

- Choose **sake** when:
  - you prefer or mostly rely on shell scripts
  - you want real-time output
  - you want the most performant task runner
  - you prefer sakes simplicity

- Choose **Ansible** when:
  - you want to use idempotent and declarative configurations
  - you are already familiar with Ansible, and don't find any of sakes features particularly useful
  - you want to use established software and relevant industry skills
  - you manage complex infrastructure

## Background

### What is Ansible

> Ansible is a radically simple IT automation system. It handles configuration management, application deployment, cloud provisioning, ad-hoc task execution, network automation, and multi-node orchestration

Ansible provides over 3000 ready-to-use modules, such as `shell`, `apt`, `file`, `copy`, etc. which you compose as playbooks and roles that you then execute for many hosts. The configuration files are written in a declarative DSL (YAML + Jinja templates).

Almost all Ansible modules are idempotent, meaning that if you run the same playbook multiple times, you will get the same results. Furthermore, Ansible is quite opinionated in terms of how you set up your roles:

> An Ansible role has a defined directory structure with eight main standard directories. You must include at least one of these directories in each role. You can omit any directories the role does not use
>
> [docs.ansible.com](https://docs.ansible.com/ansible/latest/user_guide/playbooks_reuse_roles.html#id2)

This makes sense for multiple reasons. First, it gives developers a common language to configure servers and removes yak-shaving because there are enforced decisions on how you're supposed to structure projects, and secondly, you can easily switch between projects that use Ansible.

There are some drawbacks though, the user has to become familiar with the opinionated and implicit file structure of Ansible, and there's a lot of boilerplate (more than 8 directories/files per role that you can use to customize roles: tasks, handlers, templates, files, vars, defaults, meta, module_utils, lookup_plugins, library, etc.).

### What is sake

> sake is a task runner for remote and local hosts

That's it basically. Its focus is the `ad-hoc task execution` segment that Ansible also does, but it tries to do it a bit better and more performant. You can use it to configure, deploy, and automate tasks, just like Ansible, but it's not as capable or robust as Ansible in that area.

Sake provides only one module, namely the `shell` module. The `shell` module allows you to write commands in your preferred language that you then execute for many hosts.

Similar to Ansible, the configuration files in sake are written in YAML but without any support for Jinja templates. Instead, all conditional logic is handled inside the task definitions.

So, as you can see, sake doesn't have as valiant a goal as Ansible, since to accomplish what Ansible has, you'd have to write over 3387 (and counting) modules! The benefit is that if you already have basic shell knowledge, you don't have to learn another software to set up your infrastructure, except of course for how to write tasks in sake, but that's quite simple.

Furthermore, there is no best-practice setup of how you organize your configuration files, that's all up to the user.

When it comes to idempotency, sake doesn't provide any guarantee that the tasks you write are idempotent (this is similar to Ansible if you use the `shell` module). Instead, that responsibility falls on the developer and the software you use. Fortunately, a lot of software is already idempotent by default:

- `docker-compose up` will not recreate your containers unless there is a newer image and you've set your Docker compose file to use `latest`.
- `rsync` won't upload files if there hasn't been any change
- `apt` won't reinstall your package if it's already installed
- `useradd` won't create a new user if it already exists

## Comparisons

### Dependencies

Both sake and Ansible are agentless; they don't have to be installed on the remote hosts for them to work.

Sake is delivered as a single static library that uses the ssh protocol to execute tasks on remote nodes. The remote nodes don't require `sake` to be installed.

Ansible requires Python to be installed on the control node and utilizes the ssh protocol to execute tasks on the remote nodes.
If you use Ansible modules then the Python runtime has to be installed on the remote nodes as well. The only Ansible module which doesn't require Python to be installed on the remote nodes is the `shell` module.

### Inventory

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

The only difference between sake and Ansible is how hosts belonging to multiple groups work. In Ansible you have to create a new group and specify the hosts belonging to it, whereas in sake you simply add a tag to the hosts.

### Modules and Idempotency

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

### Task Output

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

1. Add `stdout_callback = minimal` to Ansible config or define a custom callback
2. Add debug task and register out variable
3. Using debug flags `-v`

There are probably other ways, but I think it shows that it's quite bothersome to execute singular tasks and show the output of the task without resorting to editing the configs.

Sake on the other hand optimizes for showing task output. There are multiple formats (text, tables, markdown, CSV, JSON, YAML) and it's easy to change the output format for a task inline or at the command line:

```bash
$ sake run ping

TASK [ping] ***************

172.24.2.9 | pong

$ sake run ping --output table

 Server | Ping
--------+------
 172.24.2.9 | pong

$ sake run ping --output markdown

| server | ping |
|:--- |:--- |
| 172.24.2.9 | pong |
```

### Host Filtering

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

### Directory Structure

In Ansible you define hosts, playbooks, and optionally, roles. We've covered hosts already, but let's look at playbooks and roles:

- Playbooks are simply a collection of tasks that you define to run for a set of hosts.
- Roles are a collection of tasks, files, templates, handlers, and a bunch of other stuff, which you can define to improve reusability.

What you normally do is define one or multiple playbooks that are entry points to configuring servers. The playbook then refers to a set of roles, that further define the tasks, handlers, files, etc. that you use to configure your server.

The roles have an opinionated and strict directory structure that you must adhere to (for instance, the `tasks` directory must have a `main.yaml`file for it to be picked up).

One limitation with Ansible tasks is that you can only import task files, not individual tasks from a file. What you can do instead is define a tag for that specific task and then run the playbook with the `--tags` flag. I'm not sure why Ansible doesn't allow you to import specific tasks only. Also, be aware of cyclic dependencies, Ansible doesn't detect them so you can end up with infinitive loops.

sake on the other hand is like any other programming language, you import config files and can then reference specific tasks you intend to use. So you're free to structure your project any way you see fit.

### IDE & Auto-completion

Ansible has basic auto-completion for commands and flags, but it doesn't seem to be able to autocomplete single tasks, groups, tags, etc.

There's also autocompletion for running modules ad-hoc on the command line, however, when you run `ansible -m <TAB>`, you get a list of all modules, which is over 3000, so the initial load time is substantial.

sake supports rich autocompletion for both commands, flags, and values for hosts, tasks, tags, etc.

For instance, when you type `sake run <TAB> --server <TAB> --tags <TAB>`, you get the following autocompletions:

- first `<TAB>` autocompletes all available tasks
- second `<TAB>` autocompletes all available hosts
- third `<TAB>` autocompletes all available tags

Ansible has IDE integrations for vscode, vim, and other IDEs. It seems to support key attributes (gather_facts, etc.) but doesn't provide completion for hosts or tasks.

Sake doesn't have any IDE integrations yet, but I plan on adding them at a later stage when sake reaches maturity and the API has stabilized.

### Stability & Community

Ansible has a big community and is considered stable (reached v1).

Sake is still in early development and the community is non-existent.

### Misc

Sake has a list of quality-of-life features that I've not found in Ansible:

- Edit the `sake` config file via the `sake edit` command, which opens up the config file in your preferred editor, this also works for servers, tasks, targets and specs
- `sake ssh <TAB>` gives me a list of all the hosts and the ability to ssh into them
- `sake run <task> --edit` opens up the editor and navigates to the line where the task is defined, this is quite helpful when you're debugging a task
- `sake describe/list servers --tags web` lists/describes all the servers that have the tag `web`
- `sake run <task> --attach <server>` will run the task, then ssh into the server and give you a tty
- `sake run <task> --tty <server> container=my-container`, enables you to replace the current process and attach to a Docker container running on a remote server. The `<task>` can be defined as `ssh -t $S_USER@$S_HOST "docker exec -it $container"`

### Performance

Ansible is not known for its performance, in fact, just performing a simple ping to a single host over LAN takes over a second, whereas sake only takes 200 ms. And it gets worse when you increase the number of hosts, for instance, running ping on 100 hosts takes Ansible around 3.5 seconds to complete (Pipelining enabled, gather facts off and forks set to 10), and for sake only 350 ms.
Additionally, I found sake to consume less memory and CPU.

Some more extensive benchmarks can be found [here](/performance).

## Example

Now let's compare Ansible and sake with a simple demo. We'll begin with the problem definition and then look at a possible solution, using Ansible and sake:

1. Connect to a machine
2. Install `htop`
3. Start a docker container
4. Read file content
5. Upload file
6. ssh and run `htop`

### Ansible

We need to create 3 files:

```toml title=ansible.cfg
[defaults]
host_key_checking = false
```

```toml title=hosts
172.24.2.2
```

```yaml title=playbook.yaml
- hosts: all
  become: true
  vars:
   contents: "{{ lookup('file','ansible-upload.txt') }}"

  tasks:
    - name: Install htop
      ansible.builtin.apt:
        name: htop
        state: present
        update_cache: no

    - name: docker-compose up
      community.docker.docker_compose:
        project_src: /opt
        build: false

    - name: print file
      ansible.builtin.debug:
      msg: "{{ contents }}"

    - name: Upload file
      ansible.builtin.copy:
        src: ./file.txt
        dest: /home/test/
```

To run all tasks:

```bash
$ ansible-playbook -i hosts playbook.yaml

# Run htop on host
$ ssh test@172.24.2.2 -t 'htop'
```

### Sake

With sake we create 1 file containing the config, inventory, and tasks:

```yaml title=sake.yaml
disable_verify_host: true

servers:
  server-1:
    host: test@172.24.2.2:22

tasks:
  many-tasks:
    tasks:
      - name: Install htop
        cmd: sudo apt-get install htop --no-upgrade -y

      - name: docker-compose up
        cmd: docker-compose up -d

      - name: Print file
        cmd: cat file.txt

      - name: Upload file
        local: true
        env:
          file: file.txt
        cmd: scp -P "$S_PORT" "$file" $S_USER@$S_HOST:/home/test

      - name: Run htop
        tty: true
        cmd: ssh $S_USER@$S_HOST -t "htop"
```

To run all tasks:

```bash
$ sake run many-tasks -a
```

Note, you can create a generic task for the commands, and then reference them like this:

```yaml
tasks:
  upload:
    local: true
    cmd: scp -P "$S_PORT" "$from" $S_USER@$S_HOST:"$to"
    env:
      from:
      to: /home/test

  upload-file:
    task: upload
    env:
      from: file.txt
      to: /home/test
```

### Recap

As we can see, with Ansible we had to define 3 files:

- Ansible configuration file `ansible.cfg`
- Ansible inventory `hosts`
- Ansible playbook `playbook.yaml`

With sake, we only had to define 1 file. You could split it up into 3 separate files if you wanted to, but it's not required.

The main difference is that for Ansible, you have to read Ansible documentation for all the modules to know how to configure them:

- apt module
- docker-compose module
- debug bultin
- file upload module

Additionally, you have to run the `ssh` commands at the end to run `htop`, something which you don't have to do in sake.

So, if you already have basic unix knowledge, it's a lot quicker to get started with sake.
