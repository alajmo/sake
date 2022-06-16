# Tests

`sake` relies on unit tests and integration tests. The integration tests require `docker` to run, since we need to mock SSH connections somehow, and this is the easiest ways to go about it, without creating mock interfaces. The integration tests uses `golden` files to verify the output of `sake` commands.

The Docker Compose files has multiple servers with withdoc

## Unit Tests

To run unit tests run the `make unit-test`.

## Integration

For local testing first start docker compose in the background so we can have some servers to test against:

```bash
make mock-ssh
```

Then afterward run the tests:

```bash
make integration-test
```

To update the `golden` files run:

```bash
make update-golden-files
```

### Generating Keys

We generate multiple keys for testing, both encrypted and unencrypted identity keys. The key types tested are **rsa** and **ed25519**, and the key formats used are **RFC 4716** and **PEM**.

The user credentials are:

- user: test
- password: testing

The public keys are then mounted to `/home/test/.ssh/authorized_keys` via the local file `test/authorized_keys` by using Docker volume.


Stand in the tests/keys directory and run these twice (once for encrypted and once for unencrypted). For the unencrypted keys, add suffix `_no`.

```bash
# Encrypted + RSA + RFC -> id_rsa_rfc
ssh-keygen -t rsa

# Encrypted + RSA + PEM -> id_rsa_pem
ssh-keygen -t rsa -m PEM

# Encrypted + ed25519 + RFC -> id_ed25519_rfc
ssh-keygen -t ed25519

# Encrypted + ed25519 + PEM -> id_ed25519_pem
ssh-keygen -t ed25519 -m PEM
```
