---
title: ZRANGE.WATCH
description: The `ZRANGE.WATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying data.
sidebar:
  badge:
    text: Reactive
    variant: success
---

The `ZRANGE.WATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying data.
It allows clients to subscribe to a sorted set and receive notifications whenever the sorted set is updated.

## Protocol Support

| Protocol  | Supported |
| --------- | --------- |
| TCP-RESP  | ✅        |
| HTTP      | ❌        |
| WebSocket | ❌        |

## Syntax

```bash
ZRANGE.WATCH <key> <start> <stop>
```

## Parameters

| Parameter | Description                                             | Type    | Required |
| --------- | ------------------------------------------------------- | ------- | -------- |
| `key`     | key which the client would like to get updates on       | String  | Yes      |
| `start`   | start index of the sorted set                           | Integer | Yes      |
| `stop`    | stop index of the sorted set                            | Integer | Yes      |
| options   | Additional options to be passed to the command `ZRANGE` | String  | No       |

## Return Value

| Condition             | Return Value                             |
| --------------------- | ---------------------------------------- |
| Command is successful | returns a message similar to `subscribe` |

## Behavior

- The client establishes a subscription to the specified key.
- The initial result set based on the current data is sent to the client.
- DiceDB continuously monitors the key specified in the command.
- Whenever data changes that might affect the query result, the query is reevaluated.

## Errors

1. `Missing Key`
   - Error Message: `(error) ERROR wrong number of arguments for 'zrange.watch' command`
   - Occurs if no Key is provided.

## Example Usage

### Basic Usage

Let's explore a practical example of using the `ZRANGE.WATCH` command to create a real-time submission leaderboard for a game match.

```bash
127.0.0.1:7379> ZRANGE.WATCH match:100 0 1 REV WITHSCORES
Press Ctrl+C to exit watch mode.

```

This query does the following:

- Monitors key matching the name `match:100`

When the sorted set is updated using following set of commands from another client:

```bash
127.0.0.1:7379> ZADD match:100 1 "player1"
OK
127.0.0.1:7379> ZADD match:100 2 "player2"
OK
127.0.0.1:7379> ZADD match:100 1 "player3"
OK
127.0.0.1:7379> ZADD match:100 4 "player4"
OK
```

The client will receive a message similar to the following:

```bash
127.0.0.1:7379> ZRANGE.WATCH match:100 0 1 REV WITHSCORES
Press Ctrl+C to exit watch mode.
[{1 player1}]
[{2 player2} {1 player1}]
[{2 player2} {1 player3}]
[{4 player4} {2 player2}]
```

## Notes

Use the `ZRANGE.UNWATCH` command to unsubscribe from a sorted set. This will stop the client from receiving updates on the sorted set. Please refer to
the [ZRANGE.UNWATCH](/commands/zrangeunwatch) command documentation for more information.

## Related commands

following are the related commands to `ZRANGE.WATCH`:

- [GET.WATCH](/commands/getwatch)
- [PFCOUNT.WATCH](/commands/pfcountwatch)
- [ZRANGE.UNWATCH](/commands/zrangeunwatch)
