DiceDB
===

Dice 🎲 is a drop-in replacement of Redis with SQL-based realtime reactivity baked in.

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

### Setting up

To run DiceDB for local development or running from source, you will need

1. [Golang](https://go.dev/)
2. Any of the below supported platform environment:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)

```
$ git clone https://github.com/dicedb/dice
$ cd dice
$ go run main.go
```

## Dice in action

Because Dice speaks Redis' dialect, you can connect to it with any Redis Client and the simplest way it to use a [Redis CLI](https://redis.io/docs/manual/cli/). Programmatically, depending on the language you prefer, you can use your favourite Redis library to connect.

But if you are planning to use `QWATCH` feature then you need to use the DiceDB CLI that you can download from [PyPI](https://pypi.org/project/dicedb-cli/) by running the following command. The codebase for the same can be found at [dicedb/cli](https://github.com/DiceDB/cli/).

```
$ pip install dicedb-cli
```

## Running Tests

To run all the unit tests fire the following command

```sh
$ go test ./...
```

### Running a single test

```sh
$ go test -timeout 30s -run <pattern> <package path>
$ go test -timeout 30s -run ^TestByteList$ ./...
```

## Running Benchmark

```sh
$ go test -test.bench <pattern>
$ go test -test.bench BenchmarkListRedis
```

## Getting Started

To get started with building and contributing to DiceDB, please refer to the [issues](https://github.com/DiceDB/dice/issues) created in this repository.

## The story

DiceDB started as a re-implementation of Redis in Golang and the idea was to - build a DB from scratch and understand the micro-nuances that comes with its implementation. The database does not aim to replace Redis, instead it will fit in and optimize itself for multi-core computations running on a single-threaded event loop.

## How to contribute

The Code Contribution Guidelines are published at [CONTRIBUTING.md](CONTRIBUTING.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

Contributors can join the [Discord Server](https://discord.gg/6r8uXWtXh7) for quick collaboration.

## Contributors

<a href = "https://github.com/dicedb/dice/graphs/contributors">
  <img src = "https://contrib.rocks/image?repo=dicedb/dice"/>
</a>
