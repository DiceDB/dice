---
title: GET.UNWATCH
description: The `GET.UNWATCH` command is a feature to stop receiving updates on a key.
sidebar:
  badge:
    text: Reactive
    variant: success
---

The `GET.UNWATCH` command is a feature to stop receiving updates on a key.

## Protocol Support

| Protocol  | Supported |
| --------- | --------- |
| TCP-RESP  | ✅        |
| HTTP      | ❌        |
| WebSocket | ❌        |

## Syntax

```bash
GET.UNWATCH <fingerprint>
```

## Parameters

| Parameter     | Description                                         | Type   | Required |
| ------------- | --------------------------------------------------- | ------ | -------- |
| `fingerprint` | Fingerprint returned as part of the get.watch query | String | Yes      |

## Return Value

| Condition             | Return Value |
| --------------------- | ------------ |
| Command is successful | `OK`         |

## Behavior

- The client unsubscribes from the specified key.

## Errors

1. `Missing fingerprint`
   - Error Message: `(error) ERROR wrong number of arguments for 'get.watch' command`
   - Occurs if no fingerprint is provided.

## Example Usage

### Basic Usage

Let's explore a practical example of using the `GET.WATCH` command to create a real-time journal which takes backup periodically.

```bash
127.0.0.1:7379> GET.WATCH journal:user:0
GET.WATCH
Command: GET
Fingerprint: 4016579015
Data: Hello World
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
Command: GET
Fingerprint: 4016579015
Data: Hello World, I am user 0 of dice db
Command: GET
Fingerprint: 4016579015
Data: Hello World, I am user 0 of dice db, and i am going to demonstrate the use of watch commands
Command: GET
Fingerprint: 4016579015
Data: Hello World, I am user 0 of dice db, and i am going to demonstrate the use of watch and unwatch commands.
```

To stop receiving updates on the key, use the `GET.UNWATCH` command.

```bash
127.0.0.1:7379> GET.UNWATCH 4016579015
OK
```

## Notes

Use the `GET.WATCH` command to subscribe to a key. This will allow the client to receive updates on the key. Please refer to
the [GET.WATCH](/commands/getwatch) command documentation for more information.

## Related commands

following are the related commands to `GET.UNWATCH`:

- [ZRANGE.UNWATCH](/commands/zrangeunwatch)
- [PFCOUNT.UNWATCH](/commands/pfcountunwatch)
