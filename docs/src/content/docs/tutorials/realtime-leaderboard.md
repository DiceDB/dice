---
title: "Realtime Leaderboard"
description: "Create a gaming leaderboard in the easiest way possible."
---

In the world of gaming, leaderboards are essential for tracking player
rankings and improving engagement. DiceDB is an in-memory, real-time, and reactive database with Redis and SQL support optimized for modern hardware and building real-time applications. In this tutorial, we'll walk through the process of creating a gaming leaderboard using DiceDB CLI commands, without relying on any SDK.

Prerequisites:

- [DiceDB](https://github.com/dicedb/dice) installed on your system
- Basic familiarity with [DiceDB and its CLI](https://github.com/dicedb/dice?tab=readme-ov-file#dice-in-action)

## Setup DiceDB - Server and CLI

If you haven't installed DiceDB yet, follow the instructions in the [official repository](https://github.com/dicedb/dice)
to set up DiceDB on your system and start the server.

The easiest way to get started with DiceDB is using [Docker](https://www.docker.com/) by running the following command.

```
docker run dicedb/dicedb
```

The above command will start the DiceDB server running locally on the port `7379` and you can connect
to it using DiceDB CLI and SDKs.

Now that the server is running, install the DiceDB CLI
to seamlessly talk to the database. To install DiceDB CLI,
just run the following command on your machine.

```
pip install dicedb-cli
```

## Connecting to DiceDB

First, open your terminal and connect to your DiceDB instance:

```
dice-cli
```

To test that the connection is well established fire the `PING` command, and you should get `PONG` in return.

## Ingesting the stats

Let's start by ingesting some sample data into DiceDB. We'll use the [`JSON.SET`](/commands/jsonset) command to store player scores.
Once the CLI connects to the database, fire the following commands to ingest player scores.

```
JSON.SET match:1:player:1 $ {"name": "aquaman", "score": 5}
JSON.SET match:1:player:2 $ {"name": "batman", "score": 5}
JSON.SET match:1:player:3 $ {"name": "cyclops", "score": 2}
JSON.SET match:1:player:4 $ {"name": "deadpool", "score": 9}
```

Note: If you were using Redis, then you'd need to use Sorted Set to build the leaderboard
but with DiceDB, you can use the [`JSON.SET`](/commands/jsonset) command to store the metadata and score against the player as a top-level key, value pair.

## Querying the data

Open another terminal window, connect to the DiceDB with CLI.
To get the leaderboard, you need to simply query the DiceDB using the [`QWATCH`](/commands/qwatch) command.

```
QWATCH "SELECT $key, $value FROM 'match_1:%' ORDER BY $value.score DESC"
```

This command will keep emitting the list of key, value pairs in the descending order of the value (score).

## Updating Player Scores

As players complete games, you'll need to update their scores.
Use the [`JSON.SET`](/commands/jsonset) command again to update the score of the player and set it to the new value.

Say, we increase the score of `cyclops` to a `10`, then the command would be

```
JSON.SET match:1:player:3 $.score 10
```

## Realtime Leaderboard

Given that we have used [`QWATCH`](/commands/qwatch) command to get the leaderboard,
any time the data is updated, it will re-evaluate the query and
emitting the list of key, value pairs in the descending order of the value (score).

Thus, your client will get the leaderboard in realtime, without having to poll or query the data periodically.

## Conclusion

DiceDB provides a powerful and efficient solution for implementing gaming leaderboards.
By using DiceDB CLI commands, you can create fast, scalable, and feature-rich leaderboards for your games,
without having to

1. periodically poll for the data, or
2. knowing the internal data structures like Sorted Set.
