DiceDB
===

<a href="https://dicedb.io">![slatedb.io](https://img.shields.io/badge/site-dicedb.io-00A1FF?style=flat-square)</a>
<a href="https://dicedb.io/get-started/installation/">![Docs](https://img.shields.io/badge/docs-00A1FF?style=flat-square)</a>
<a target="_blank" href="https://discord.gg/6r8uXWtXh7"><img src="https://dcbadge.limes.pink/api/server/6r8uXWtXh7?style=flat" alt="discord community" /></a>
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)
![GitHub Sponsor](https://img.shields.io/github/sponsors/arpitbbhayani?label=Sponsor&logo=GitHub)

### What is DiceDB?

DiceDB is an open-source, fast, reactive, in-memory database optimized for modern hardware. Commonly used as a cache, it offers a familiar interface while enabling real-time data updates through query subscriptions. It delivers higher throughput and lower median latencies, making it ideal for modern workloads.

## Get started

### Setting up DiceDB with Docker

The easiest way to get started with DiceDB is using [Docker](https://www.docker.com/) by running the following command.

```bash
$ docker run -p 7379:7379 dicedb/dicedb:latest
```

The above command will start the DiceDB server running locally on the port `7379` and you can connect
to it using [DiceDB CLI](https://github.com/DiceDB/dicedb-cli) and SDKs.

> [!NOTE]
> If you are looking to setup DiceDB for development or want to setup from source, refer
> our [CONTRIBUTING/README.md](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/README.md) guide.

## Setting up CLI

### Using cURL

The best way to connect to DiceDB is using [DiceDB CLI](https://github.com/DiceDB/dicedb-cli) and you can install it by running the following command

```bash
$ sudo su
$ curl -sL https://raw.githubusercontent.com/DiceDB/dicedb-cli/refs/heads/master/install.sh | sh
```

If you are working on unsupported OS (as per above script), you can always follow the installation instructions mentioned in the [dicedb/cli](https://github.com/DiceDB/dicedb-cli) repository.

> [!NOTE]
> If you are looking to setup DiceDB for development or want to setup from source, refer
> our [CONTRIBUTING/README.md](https://github.com/DiceDB/dice/blob/master/CONTRIBUTING/README.md) guide.

## Want to contribute?

The Code Contribution Guidelines are published at [CONTRIBUTING/README.md](CONTRIBUTING/README.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

Contributors can join the [Discord Server](https://discord.gg/6r8uXWtXh7) for quick collaboration.

## Sponsors

We are incredibly grateful to our sponsors for their generous support, which makes the development of DiceDB possible.

[![CodeRabbit](https://www.coderabbit.ai/images/logo-white.svg)](https://www.coderabbit.ai/?utm_source=github&utm_medium=social&utm_campaign=sponsor&utm_term=dicedb)

## Support and Sponsor Us

DiceDB is a project with a very strong vision and [roadmap](https://dicedb.io/roadmap/). If you like what
we do and find DiceDB useful, please consider supporting and [sponsoring us on GitHub](https://github.com/sponsors/arpitbbhayani).

![GitHub Sponsor](https://img.shields.io/github/sponsors/arpitbbhayani?label=Sponsor&logo=GitHub)

## Contributors

<a href = "https://github.com/dicedb/dice/graphs/contributors">
  <img src = "https://contrib.rocks/image?repo=dicedb/dice"/>
</a>

## License

This project is licensed under the BSD 3-Clause License. See the [LICENSE](LICENSE) file for details.
