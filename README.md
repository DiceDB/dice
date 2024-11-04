DiceDB
===

<a href="https://dicedb.io">![slatedb.io](https://img.shields.io/badge/site-dicedb.io-00A1FF?style=flat-square)</a>
<a href="https://dicedb.io/get-started/installation/">![Docs](https://img.shields.io/badge/docs-00A1FF?style=flat-square)</a>
<a target="_blank" href="https://discord.gg/6r8uXWtXh7"><img src="https://dcbadge.limes.pink/api/server/6r8uXWtXh7?style=flat" alt="discord community" /></a>

DiceDB is a redis-compliant, in-memory, real-time, and reactive database optimized for modern hardware and for building and scaling truly real-time applications. 

We are looking for Early Design Partners, so, if you want to evaluate DiceDB, [block our calendar](https://cal.com/dicedb-arpit). always up for a chat.

> [!CAUTION]
> DiceDB is under development and it supports a subset of Redis commands. So, please do not use it in production. But, feel free to go through the [open issues](https://github.com/DiceDB/dice/issues) and contribute to help us speed up the development.

## Want to contribute?

We have multiple repositories where you can contribute. So, as per your interest, you can pick one and build a deeper understanding of the project on the go.

- [dicedb/dice](https://github.com/dicedb/dice) for core database features and engine / Stack - Go
- [dicedb/dicedb-cli](https://github.com/dicedb/dicedb-cli) command line interface for DiceDB / Stack - Go
- [dicedb/playground-mono](https://github.com/dicedb/playground-mono) backend APIs for DiceDB playground / Stack - Go
- [dicedb/alloy](https://github.com/dicedb/alloy) frontend and marketplace for DiceDB playground / Stack - NextJS

## How is it different from Redis?

Although DiceDB is a drop-in replacement of Redis, which means almost no learning curve and switching does not require any code change, it still differs in two key aspects and they are

1. DiceDB is multithreaded and follows [shared-nothing architecture](https://en.wikipedia.org/wiki/Shared-nothing_architecture).
2. DiceDB supports `.WATCH` commands like `GET.WATCH`, `ZRANGE.WATCH`, etc. that lets clients listen to data changes and get the result set in real-time whenever something changes.

`.WATCH` commands are pretty handy when it comes to building truly real-time applications like [Leaderboard](https://github.com/arpitbbhayani/leaderboard-go-dicedb).

## Get started

### Setting up DiceDB with Dockerc

The easiest way to get started with DiceDB is using [Docker](https://www.docker.com/) by running the following command.

```bash
docker run -p 7379:7379 dicedb/dicedb --enable-multithreading --enable-watch
```

The above command will start the DiceDB server running locally on the port `7379` and you can connect
to it using [DiceDB CLI](https://github.com/DiceDB/dicedb-cli) and SDKs.

> [!TIP]
> Since DiceDB is a drop-in replacement for Redis, you can also use any Redis CLI and SDK to connect to DiceDB.


### Setting up DiceDB from source for development and contributions

To run DiceDB for local development or running from source, you will need

1. [Golang](https://go.dev/)
2. Any of the below supported platform environments:
    1. [Linux based environment](https://en.wikipedia.org/wiki/Comparison_of_Linux_distributions)
    2. [OSX (Darwin) based environment](https://en.wikipedia.org/wiki/MacOS)
    3. WSL under Windows

```bash
git clone https://github.com/dicedb/dice
cd dice
go run main.go --enable-multithreading --enable-watch
```

You can skip passing the two flags if you are not working with multi-threading or `.WATCH` features.

1. Install GoLangCI

```bash
sudo su
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /bin v1.60.1
```

### Live Development Server

DiceDB provides a hot-reloading development environment, which allows you to instantly view your code changes in a live server. This functionality is supported by [Air](https://github.com/air-verse/air)

To Install Air on your system you have the following options.

1. If you're on go 1.22+
```bash
go install github.com/air-verse/air@latest
```

1. Install the Air binary
```bash
# binary will be installed at $(go env GOPATH)/bin/air
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

Once `air` is installed you can verify the installation using the command `air -v`

To run the live DiceDB server for local development:

```bash
git clone https://github.com/dicedb/dice
cd dice
air
```

### Local Setup with Custom Config

By default, DiceDB will look for the configuration file at `/etc/dice/config.toml`. (Linux, Darwin, and WSL)

```bash
# set up configuration file # (optional but recommended)
sudo mkdir -p /etc/dice
sudo chown root:$USER /etc/dice
sudo chmod 775 /etc/dice # or 777 if you are the only user
git clone https://github.com/DiceDB/dice.git
cd dice
go run main.go -init-config

```

> [!NOTE]
> #### For Windows Users:
> If you're using Windows, it is recommended to use Windows Subsystem for Linux (WSL) or WSL 2 to run the above commands seamlessly in a Linux-like environment.

#### Additional Configuration Options:

If you'd like to use a different location, you can specify a custom configuration file path with the `-c flag`:

```bash
go run main.go -c /path/to/config.toml
```
If you'd like to output the configuration file to a specific location, you can specify a custom output path with the `-o flag`:

```bash
go run main.go -o /path/of/output/dir
```


### Setting up CLI

The best way to connect to DiceDB is using DiceDB CLI and you can install it by running the following command

```bash
sudo su
curl -sL https://raw.githubusercontent.com/DiceDB/dicedb-cli/refs/heads/master/install.sh | sh
```

### Client Compatibility

DiceDB is fully compatible with Redis protocol, allowing you to connect using any existing Redis client or SDK.

> [!NOTE]
> The `.WATCH` feature is only accessible through the DiceDB CLI.
> If you are working on unsupported OS (as per above script), you can always follow the installation instructions mentioned in the [dicedb/cli](https://github.com/DiceDB/dicedb-cli) repository.

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

```bash
TEST_FUNC=<name of the test function> make test-one
TEST_FUNC=TestSet make test-one
```

### Running all integration tests

```bash
make test
```

> Work to add more tests in DiceDB is in progress, and we will soon port the
> test [Redis suite](https://github.com/redis/redis/tree/f60370ce28b946c1146dcea77c9c399d39601aaa) to this codebase to ensure full compatibility.

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

## The Story

DiceDB started as a re-implementation of Redis in Golang with the idea of building a DB from scratch to understand the micro-nuances that come with its implementation. DiceDB isn’t just another database; it’s a platform purpose-built for the real-time era. As real-time systems become increasingly prevalent in modern applications, DiceDB’s hyper-optimized architecture is positioned to power the next generation of user experiences.

## How to contribute

The Code Contribution Guidelines are published at [CONTRIBUTING/README.md](CONTRIBUTING/README.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

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