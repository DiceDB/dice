# Development Setup for Engine

## Setting up DiceDB

### Build from source

To build DiceDB from source, you need to have the following

1. [Golang](https://go.dev/)
2. Any of the below supported platform environments:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows

```sh
$ git clone https://github.com/dicedb/dice
$ cd dice
$ make build
```

The above command will create a binary `dicedb`. Execute the binary and that will
start the DiceDB server., or, you can run the following command to run like a usual
Go program

```sh
$ go run main.go
```

## Setting up DiceDB CLI

### Building from source

```sh
$ git clone https://github.com/DiceDB/dicedb-cli
$ cd dicedb-cli
$ make build
```

The above command will create a binary `dicedb-cli`. Execute the binary will
start the CLI and will try to connect to the DiceDB server.

## Essentials for Development

### Pointing to local checked-out `dicedb-go`

It is advised to checkout [dicedb-go](https://github.com/DiceDB/dicedb-go) repository also because `dice` takes
a strong dependency on it. To point to the local copy add the following line
at the end of the `go.mod` file.

```
replace github.com/dicedb/dicedb-go => ../dicedb-go
```

Note: this is the literal line that needs to be added at the end of the go.mod file.
Refer to [this article](https://thewebivore.com/using-replace-in-go-mod-to-point-to-your-local-module/), to understand what it is and why it is needed.

Do not check-in the `go.mod` file with this change.

### Install GoLangCI

```bash
$ sudo su
$ sudo curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /bin v1.64.6
```

### Local Setup with Custom Config

Follow these steps to generate and customize your dicedb configuration in a local setup:

```bash
$ go run main.go config-init
```

This will generate configuration file (`dicedb.yaml`) in metadata directory.
Metadata directory is OS-specific,

 - macOS: `/usr/local/etc/dicedb/dicedb.yaml`
 - Linux: `/etc/dicedb/dicedb.yaml`

If you run with a `sudo` privileges, then these directories are used, otherwise
the current working directory is used as the metadata directory.

### Running Tests

Unit tests and integration tests are essential for ensuring correctness and in the case of DiceDB, both types of tests are available to validate its functionality.

For unit testing, you can execute individual unit tests by specifying the name of the test function using the `TEST_FUNC` environment variable and running the `make unittest-one` command. Alternatively, running `make unittest` will execute all unit tests.

### Executing one unit test

```bash
TEST_FUNC=<name of the test function> make unittest-one
TEST_FUNC=TestByteList make unittest-one
```

### Running all unit tests

```bash
make unittest
```

Integration tests, on the other hand, involve starting up the DiceDB server and running a series of commands to verify the expected end state and output. To execute a single integration test, you can set the `TEST_FUNC` environment variable to the name of the test function and run `make test-one`. Running `make test` will execute all integration tests.

### Executing a single integration test

Make sure you have DiceDB running before you run the following commands.
By default it connects to the local instance of DiceDB running on port `7379`.

```bash
TEST_FUNC=<name of the test function> make test-one
TEST_FUNC=^TestSet$ make test-one
```

### Running all integration tests

```bash
$ make test
```
