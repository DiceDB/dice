DiceDB
===

DiceDB is a drop-in replacement of Redis with SQL-based real-time reactivity baked in.

> Note: DiceDB is still in development and it supports a subset of Redis commands. So, please do not use it in production. But, feel free to go through the [open issues](https://github.com/DiceDB/dice/issues) and contribute to help us speed up the development.

## How is it different from Redis?

1. DiceDB is multi-threaded and follows [shared-nothing architecture](https://en.wikipedia.org/wiki/Shared-nothing_architecture).
2. DiceDB supports a new command called `QWATCH` that lets clients listen to a SQL query and get notified in real-time whenever something changes.

## Get started

### Using Docker

The easiest way to get started with DiceDB is using [Docker](https://www.docker.com/) by running the following command.

```
$ docker run dicedb/dice-server
```

The above command will start the DiceDB server running locally on the port `7379` and you can connect
to it using DiceDB CLI and SDKs, or even Redis CLIs and SDKs.

> Note: Given it is a drop-in replacement of Redis, you can also use any Redis CLI and SDK to connect to DiceDB.

### Setting up DiceDB from source for development and contributions

To run DiceDB for local development or running from source, you will need

1. [Golang](https://go.dev/)
2. Any of the below supported platform environment:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows

```
$ git clone https://github.com/dicedb/dice
$ cd dice
$ go run main.go
```

### Live Development Server

DiceDB provides a hot-reloading development environment, which allows you to instantly view your code changes in a live server. This functionality is supported by [Air](https://github.com/air-verse/air)

To Install Air on your system you have following options.

1. If you're on go 1.22+
```sh
go install github.com/air-verse/air@latest
```



2. Install the Air binary
```sh
# binary will be installed at $(go env GOPATH)/bin/air
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

Once `air` is installed you can verify the installation using the command `air -v`

To run the live DiceDB server for local development:

```sh
$ git clone https://github.com/dicedb/dice
$ cd dice
$ air
```


## Setting up CLI

The best way to connect to DiceDB is using DiceDB CLI and you can install it by running the following command.

```
$ pip install dicedb-cli
```

> Because DiceDB speaks Redis dialect, you can connect to it with any Redis Client and SDK also.
> But if you are planning to use the `QWATCH` feature then you need to use the DiceDB CLI.

## Running Tests

Unit tests and integration tests are essential for ensuring correctness and in the case of DiceDB, both types of tests are available to validate its functionality.

For unit testing, you can execute individual unit tests by specifying the name of the test function using the `TEST_FUNC` environment variable and running the `make unittest-one` command. Alternatively, running `make unittest` will execute all unit tests.

### Executing one unit test

```
$ TEST_FUNC=<name of the test function> make unittest-one
$ TEST_FUNC=TestByteList make unittest-one
```

### Running all unit tests

```
$ make unittest
```

Integration tests, on the other hand, involve starting up the DiceDB server and running a series of commands to verify the expected end state and output. To execute a single integration test, you can set the `TEST_FUNC` environment variable to the name of the test function and run `make test-one`. Running `make test` will execute all integration tests.

### Executing a single integration test

```
$ TEST_FUNC=<name of the test function> make test-one
$ TEST_FUNC=TestSet make test-one
```

### Running all integration tests

```
$ make test
```

> Work to add more tests in DiceDB is in progress and we will soon port the
> test [Redis suite](https://github.com/redis/redis/tree/f60370ce28b946c1146dcea77c9c399d39601aaa) to this codebase to ensure full compatibility.

## Running Benchmark

```sh
$ go test -test.bench <pattern>
$ go test -test.bench BenchmarkListRedis -benchmem
```

## Getting Started

To get started with building and contributing to DiceDB, please refer to the [issues](https://github.com/DiceDB/dice/issues) created in this repository.

## The story

DiceDB started as a re-implementation of Redis in Golang and the idea was to - build a DB from scratch and understand the micro-nuances that come with its implementation. The database does not aim to replace Redis, instead, it will fit in and optimize itself for multi-core computations running on a single-threaded event loop.

## How to contribute

The Code Contribution Guidelines are published at [CONTRIBUTING.md](CONTRIBUTING.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

Contributors can join the [Discord Server](https://discord.gg/6r8uXWtXh7) for quick collaboration.

## Contributors

<a href = "https://github.com/dicedb/dice/graphs/contributors">
  <img src = "https://contrib.rocks/image?repo=dicedb/dice"/>
</a>

## Troubleshoot

### Forcefully killing the process

```
$ sudo netstat -atlpn | grep :7379
$ sudo kill -9 <process_id>
```
