DiceDB
===

<a href="https://dicedb.io">![slatedb.io](https://img.shields.io/badge/site-dicedb.io-00A1FF?style=flat-square)</a>
<a href="https://dicedb.io/get-started/installation/">![Docs](https://img.shields.io/badge/docs-00A1FF?style=flat-square)</a>
<a target="_blank" href="https://discord.gg/6r8uXWtXh7"><img src="https://dcbadge.limes.pink/api/server/6r8uXWtXh7?style=flat" alt="discord community" /></a>
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)

DiceDB is an open-source in-memory database with query subscriptions.

> [!CAUTION]
> DiceDB is under active development and supports a wide [range of commands](https://dicedb.io/commands/get/). Inconsistent behavior is expected. Feel free to go through the [open issues](https://github.com/DiceDB/dice/issues) and contribute to help us speed up the development.

## Want to contribute?

We have multiple repositories where you can contribute. So, as per your interest, you can pick one and build a deeper understanding of the project on the go.

- [dicedb/dice](https://github.com/dicedb/dice) for core database features and engine / Stack - Go
- [dicedb/dicedb-cli](https://github.com/dicedb/dicedb-cli) command line interface for DiceDB / Stack - Go
- [dicedb/playground-mono](https://github.com/dicedb/playground-mono) backend APIs for DiceDB playground / Stack - Go
- [dicedb/alloy](https://github.com/dicedb/alloy) frontend and marketplace for DiceDB playground / Stack - NextJS

## Get started

### Setting up DiceDB with Docker

The easiest way to get started with DiceDB is using [Docker](https://www.docker.com/) by running the following command.

```bash
docker run -p 7379:7379 dicedb/dicedb --enable-watch
```

The above command will start the DiceDB server running locally on the port `7379` and you can connect
to it using [DiceDB CLI](https://github.com/DiceDB/dicedb-cli) and SDKs.

### Build from source

To build DiceDB from source, you need to have the following

1. [Golang](https://go.dev/)
2. Any of the below supported platform environments:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows

```sh
git clone https://github.com/dicedb/dice
cd dice
make build
```

The above command will create a binary `dicedb`. Execute the binary and that will
start the DiceDB server., or, you can run the following command to run like a usual
Go program

```sh
go run main.go --enable-watch
```

You can skip passing the flag if you are not working with `.WATCH` feature.

## Setting up CLI

### Using cURL

The best way to connect to DiceDB is using [DiceDB CLI](https://github.com/DiceDB/dicedb-cli) and you can install it by running the following command

```bash
sudo su
curl -sL https://raw.githubusercontent.com/DiceDB/dicedb-cli/refs/heads/master/install.sh | sh
```

If you are working on unsupported OS (as per above script), you can always follow the installation instructions mentioned in the [dicedb/cli](https://github.com/DiceDB/dicedb-cli) repository.

### Building from source

```sh
git clone https://github.com/DiceDB/dicedb-cli
cd dicedb-cli
make build
```

The above command will create a binary `dicedb-cli`. Execute the binary will
start the CLI and will try to connect to the DiceDB server.

## Essentials for Development

### Install GoLangCI

```bash
sudo su
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /bin v1.60.1
```

### Local Setup with Custom Config

Follow these steps to generate and customize your dicedb configuration in a local setup:

```bash
go run main.go config-init
```

This will generate configuration file (`dicedb.yaml`) in metadata directory.
Metadata directory is OS-specific,

 - macOS: `/usr/local/etc/dicedb/dicedb.yaml`
 - Linux: `/etc/dicedb/dicedb.yaml`

If you run with a `sudo` privileges, then these directories are used, otherwise
the current working directory is used as the metada directory.

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
make test
```

## Running Benchmark

```bash
make run_benchmark
```

## Getting Started

To get started with building and contributing to DiceDB, please refer to the [issues](https://github.com/DiceDB/dice/issues) created in this repository.

## Docs

We use [Astro](https://astro.build/) framework to power the [dicedb.io website](https://dicedb.io) and [Starlight](https://starlight.astro.build/) to power the docs. Once you have NodeJS installed, fire the following commands to get your local version of [dicedb.io](https://dicedb.io) running.

```bash
cd docs
npm install
npm run dev
```

Once the server starts, visit http://localhost:4321/ in your favourite browser. This runs with a hot reload which means any changes you make in the website and the documentation can be instantly viewed on the browser.

### Docs directory structure

1. `docs/src/content/docs/commands` is where all the commands are documented
2. `docs/src/content/docs/tutorials` is where all the tutorials are documented

## How to contribute

The Code Contribution Guidelines are published at [CONTRIBUTING/README.md](CONTRIBUTING/README.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

Contributors can join the [Discord Server](https://discord.gg/6r8uXWtXh7) for quick collaboration.

## Contributors

<a href = "https://github.com/dicedb/dice/graphs/contributors">
  <img src = "https://contrib.rocks/image?repo=dicedb/dice"/>
</a>

## License

This project is licensed under the BSD 3-Clause License. See the [LICENSE](LICENSE) file for details.

## Troubleshoot

### Forcefully killing the process

```
$ sudo netstat -atlpn | grep :7379
$ sudo kill -9 <process_id>
```
