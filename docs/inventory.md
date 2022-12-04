# Inventory

Inventory is a collection of hosts that you can execute tasks on. There are 4 ways to specify hosts:

```yaml
servers:
  # Single Host
  single-1:
    host: 192.168.1.1
    user: test
    port: 33
    bastion: test@192.168.0.1:22
  tags: [web]

  # Multipe hosts using a list
  many-1:
    hosts:
      - test@192.168.1.1:22
      - test@192.168.1.2:22
    tags: [web, prod]

  # Multiple hosts using a ranges
  many-2:
    hosts: test@192.168.1.[1:2]:22
    tags: [web, prod]

  # Multiple hosts by invoking a shell command
  many-3:
    inventory: echo test@192.168.1.1:22 test@192.168.1.2:22
    tags: [web, prod]
```

To target the hosts in a task there's multiple ways:

- **all**: target
- **servers**: a list of single hosts or group of hosts
  - supports range as well, for instance `--server "list[0:1]"`, select first and second host
- **tags**: target hosts that have a specific tag
- **regex**: target hosts on host regex
- **invert**: invert matching on hosts

Furthermore, to limit the number of targetted servers, you can use one of following properties:

- **limit**: limit the number of targetted hosts
- **limit_p**: limit the number of targetted hosts in percentage

## Provide Identity and Password Credentials

By default `sake` will attempt to load identity keys from an SSH agent if it's running in the background. However, if you wish to provide credentials manually, you can do so by (first takes precedence):

1. setting `--identity-file` and/or `--password` flags
2. specifying it in the server definition

The type of auth used is determined by:

- if `identity-file` and `password` are provided, then it assumes password protected identity key
- if only `identity-file` is provided, then it first tries without passphrase, if file is encrypted, it will prompt for passphrase
- if only `password` is provided, then it assumes password protected auth

```yaml
servers:
  server-1:
    host: server-1.lan
    identity_file: id_rsa
    password: $(echo $MY_SECRET_PASSWORD)
```

You can also define entries in your `~/.ssh/config` file and `sake` will try to resolve them.

## Known Hosts

By default a `known_hosts` file is used to verify host connections. If you wish to disable verification, set the global property `disable_verify_host` to true:

```yaml
disable_verify_host: true
```

The default location of the known hosts file is `$HOME/.ssh/known_hosts`. If you wish change this to another file, then set the global property `known_hosts_file` to your desired filepath:

```yaml
known_hosts_file: ./known_hosts
```
