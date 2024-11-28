---
title: ZRANGE.UNWATCH
description: The `ZRANGE.UNWATCH` command is a feature to stop receiving updates on a sorted set
sidebar:
  badge:
    text: Reactive
    variant: success
---

The `ZRANGE.UNWATCH` command is a feature to stop receiving updates on a sorted set.

## Protocol Support

| Protocol  | Supported |
| --------- | --------- |
| TCP-RESP  | ✅        |
| HTTP      | ❌        |
| WebSocket | ❌        |

## Syntax

```bash
ZRANGE.UNWATCH <fingerprint>
```

## Parameters

| Parameter     | Description                                            | Type   | Required |
| ------------- | ------------------------------------------------------ | ------ | -------- |
| `fingerprint` | Fingerprint returned as part of the zrange.watch query | String | Yes      |

## Return Value

| Condition             | Return Value |
| --------------------- | ------------ |
| Command is successful | `OK`         |

## Behavior

- The client unsubscribes from the sorted set specified with the key.

## Errors

1. `Missing fingerprint`
   - Error Message: `(error) ERROR wrong number of arguments for 'zrange.unwatch' command`
   - Occurs if no fingerprint is provided.

## Example Usage

### Basic Usage

Let's explore a practical example of using the `ZRANGE.WATCH` command to create a real-time submission leaderboard for a game match.

```bash
127.0.0.1:7379> ZRANGE.WATCH match:100 0 1 REV WITHSCORES
ZRANGE.WATCH
Command: ZRANGE
Fingerprint: 4016579015
Data:
```

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
Command: ZRANGE
Fingerprint: 4016579015
Data: [{1 player1}]
Command: ZRANGE
Fingerprint: 4016579015
Data: [{2 player2}]
Command: ZRANGE
Fingerprint: 4016579015
Data: [{2 player2}]
Command: ZRANGE
Fingerprint: 4016579015
Data: [{4 player4}]
```

To stop receiving updates on the key, use the `ZRANGE.UNWATCH` command.

```bash
127.0.0.1:7379> ZRANGE.UNWATCH 4016579015
OK
```

## Notes

Use the `ZRANGE.WATCH` command to subscribe to a key. This will allow the client to receive updates on the sorted set. Please refer to
the [ZRANGE.WATCH](/commands/zrangewatch) command documentation for more information.

## Related commands

following are the related commands to `ZRANGE.UNWATCH`:

- [ZRANGE.WATCH](/commands/zrangewatch)
- [GET.UNWATCH](/commands/getunwatch)
