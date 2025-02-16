---
name: Fix integration tests for a command
about: This template is supposed to be used by maintainers tracking the port of a integration test to IronHawk
title: 'IronHawk Integration Tests: '
labels: ''
assignees: ''
---

We reimplemented the core of the DiceDB engine and re-wrote - the wire protocol, execution engine, and config management. This rewrite helped us gain 32% throughput over our existing benchmark.

To make DiceDB stable even after 100s of changes, we have to put effort
into making sure we have a comprehensive integration test suite for each command. This issue will be used to cover the port and fixes for integration tests related to the command `CMD`.

Here are the pre-requisite

1. setup DiceDB server locally from the source - [instructions](https://github.com/dicedb/dice)
2. setup DiceDB Go SDK locally from the source - [instructions](https://github.com/dicedb/dicedb-go)
3. refer to the `Pointing to local checked-out `dicedb-go` section in `README`.

### Start the DiceDB server with IronHawk engine

```
$ go run main.go --engine ironhawk --log-level debug
```

## Setting up Integration Tests

0. Make sure the DiceDB server is running. This is essential for you to run the tests
1. Integration tests of all the commands can be found under `tests/commands/ironhawk` with the name `cmd_test.go`.
2. For the command `cmd` find the tests
3. Run the test function using the following command

```
$ TEST_FUNC=^TestSet$ make test-one
```

Replace the name of the function with whatever the name is in your `cmd_test.go` file.
Note the `^` and `$` in the `TEST_FUNC` variable. It is a regex and this way, the command
executes only one function which is `TestSet`.

## Things need to be done

1. Fix any dependency error (refer `set_test.go` file)
2. Fix any execution error (refer `set_test.go` file)
3. There are some utility functions written, use them, but as per the `set_test.go` file.

Ideally, all the tests should pass. If some are failing

1. either fix them (if you think it is a bug in the tests)
2. or raise a bug if you think there is an implementation mistake

Eventually, we need 100% integration test coverage for all the commands to
prove that DiceDB is stable and production-ready.

If you find any other bug while you are implementing it, you can either

1. fix it yourself and submit it in a new PR
2. raise a [GitHub issue](https://github.com/DiceDB/dice/issues)

## Follow the contribution guidelines

These are general guidelines to follow before you submit a patch. Please mark them as done
once you complete them

- [ ] please go through the [CONTRIBUTING](https://github.com/DiceDB/dice/tree/master/CONTRIBUTING) guide
- [ ] follow [LOGGING best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/logging.md)
- [ ] follow [Golang best practices](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/go.md)
- [ ] run `make lint` on your local copy of the codebase
