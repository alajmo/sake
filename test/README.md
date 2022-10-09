# Tests

`sake` relies on unit tests and integration tests. The integration tests require `docker` to run, since we need to mock SSH connections somehow, and this is the easiest ways to go about it, without creating mock interfaces. The integration tests uses `golden` files to verify the output of `sake` commands.
`sake` also has a few profile tests which measures execution time of various commands.

The Docker Compose files has multiple servers with different auths (pem, rfc, ed, rsa, encrypted, unencrypted, etc.) and ip addresses (ipv4/6).

## Profiling

To run benchmark run `make benchmark`, and to save benchmarks run `make benchmark-save`.

Additionally, if you wish to dive into cpu, heap, and goroutine profiling, replace the `main.go` with the contents of `debug.go`, cd into test directory and run:

```sh
go run ../main.go run ping -t reachable
go tool pprof -http="localhost:8000" ./benchmarks/cpu.prof
go tool pprof -http="localhost:8000" ./benchmarks/heap.prof
go tool pprof -http="localhost:8000" ./benchmarks/goroutine.prof
```

## Unit Tests

To run unit tests run `make unit-test`.

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
We also have servers accepting only user and passwords.

The user credentials for private keys are:

- user: test
- password: testing

The user credentials for password auth is:

- user: test
- password: test

The public keys are then mounted to `/home/test/.ssh/authorized_keys` via the local file `test/authorized_keys` by using Docker volume.

Stand in the keys directory and run the following commands:

```bash
# Encrypted + RSA + RFC
ssh-keygen -t rsa -f id_rsa_rfc -C "test@test.test" -N "testing"

# Encrypted + RSA + PEM
ssh-keygen -t rsa -f id_rsa_pem -C "test@test.test" -N "testing" -m PEM

# Encrypted + ed25519 + RFC
ssh-keygen -t ed25519 -f id_ed25519_rfc -C "test@test.test" -N "testing"

# Encrypted + ed25519 + PEM
ssh-keygen -t ed25519 -f id_ed25519_pem -C "test@test.test" -N "testing" -m PEM

# Unencrypted + RSA + RFC
ssh-keygen -t rsa -f id_rsa_rfc_no -C "test@test.test" -N ""

# Unencrypted + RSA + PEM
ssh-keygen -t rsa -f id_rsa_pem_no -C "test@test.test" -N "" -m PEM

# Unencrypted + ed25519 + RFC
ssh-keygen -t ed25519 -f id_ed25519_rfc_no -C "test@test.test" -N ""

# Unencrypted + ed25519 + PEM
ssh-keygen -t ed25519 -f id_ed25519_pem_no -C "test@test.test" -N "" -m PEM
```
