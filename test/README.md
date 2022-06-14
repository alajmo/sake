# Tests

`sake` relies on unit tests and integration tests. The integration tests require `docker` to run, since we need to mock SSH connections somehow, and this is the easiest ways to go about it, without creating mock interfaces. The integration tests uses `golden` files to verify the output of `sake` commands.

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
