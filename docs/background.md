# Background

> To configure servers, all you need is bash scripting, environment variables, and some know-how.

`sake` came about because I needed a simple tool to run tasks on remote hosts. There's tons of software in this category, Ansible, pyinfra, Puppet, Chef, Salt, Sup, and probably many more. However, some of them can be quite complex to master, have complicated DSLs and for simple situations are not quite ergonomic. So, `sake` was born, an ergonomic utility CLI tool to run tasks on servers.

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

With these principles in mind, I've elected not to create a complex DSL (see Terraform, Puppet, Ansible, etc.), but instead, just add a few primitives and let the user leverage their existing sysadmin knowledge to create their ideal setup. This results in not forcing users to learn yet another DSL, avoiding continuous lookup of `sake` commands, and updating `sake` to get new features.

It does, however, push complexity onto the user, for instance, there's no built-in primitive to download files and the user must define a task to do so. It would be foolish to aim for feature parity with 3rd party software like rsync for downloading files (there are over 150 flags for rsync).

So what config format is best suited for this purpose? In my opinion, YAML is a suitable candidate. While it has its issues, I think its purpose as a human-readable config/state file works well. It has all the primitives you'd need in a config language, simple key/value entries, dictionaries, and lists, as well as supporting comments (something which JSON doesn't). We could create a custom format, but then users would have to learn that syntax, so in this case, YAML has a major advantage, many developers are familiar with it.

I don't intend to introduce any templating capability via Jinja templates or similar, instead, the user will have to leverage shell or any other programming to compose more complex tasks.

In this sense, `sake` follows the Unix principle:

- Write programs that do one thing and do it well - run tasks on multiple servers
- Write programs to work together - invoke other programs to do your bidding (rsync)
- Write programs to handle text streams - output of all `sake` commands is just text
