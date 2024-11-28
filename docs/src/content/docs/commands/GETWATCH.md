---
title: GET.WATCH
description: The `GET.WATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying data.
sidebar:
  badge:
    text: Reactive
    variant: success
---

The `GET.WATCH` command is a novel feature designed to provide real-time updates to clients based on changes in underlying data.
It allows clients to subscribe to a key and receive notifications whenever the key is updated.

This command is what makes DiceDB different from Redis and uniquely positions it as the easiest and most intuitive way
to build real-time reactive applications like leaderboards, documents etc.

## Protocol Support

| Protocol  | Supported |
| --------- | --------- |
| TCP-RESP  | ✅        |
| HTTP      | ❌        |
| WebSocket | ❌        |

## Syntax

```bash
GET.WATCH <key>
```

## Parameters

| Parameter | Description                                       | Type   | Required |
| --------- | ------------------------------------------------- | ------ | -------- |
| `key`     | key which the client would like to get updates on | String | Yes      |

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
   - Error Message: `(error) ERROR wrong number of arguments for 'get.watch' command`
   - Occurs if no Key is provided.

## Example Usage

### Basic Usage

Let's explore a practical example of using the `GET.WATCH` command to create a real-time journal which takes backup periodically.

```bash
127.0.0.1:7379> GET.WATCH journal:user:0
Press Ctrl+C to exit watch mode.
"Hello World"
```

This query does the following:

- Monitors key matching the name `journal:user:0`

When the key is updated using following set of commands from another client:

```bash
127.0.0.1:7379> set journal:user:0 "Hello World, I am user 0 of dice db"
OK
127.0.0.1:7379> set journal:user:0 "Hello World, I am user 0 of dice db, and i am going to demonstrate the use of watch commands"
OK
127.0.0.1:7379> set journal:user:0 "Hello World, I am user 0 of dice db, and i am going to demonstrate the use of watch and unwatch commands."
OK
```

The client will receive a message similar to the following:

```bash
127.0.0.1:7379> GET.WATCH journal:user:0
Press Ctrl+C to exit watch mode.
"Hello World"
"Hello World, I am user 0 of dice db"
"Hello World, I am user 0 of dice db, and i am going to demonstrate the use of watch commands"
"Hello World, I am user 0 of dice db, and i am going to demonstrate the use of watch and unwatch commands."
```

## Notes

Use the `GET.UNWATCH` command to unsubscribe from a key. This will stop the client from receiving updates on the key. Please refer to
the [GET.UNWATCH](/commands/getunwatch) command documentation for more information.

## Related commands

following are the related commands to `GET.WATCH`:

- [GET.UNWATCH](/commands/getunwatch)
- [ZRANGE.WATCH](/commands/zrangewatch)
- [PFCOUNT.WATCH](/commands/pfcountwatch)
