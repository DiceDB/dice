---
title: PFCOUNT.WATCH
description: The `PFCOUNT.WATCH` command is a novel feature designed to provide real-time updates to clients whenever the cardinality of a HyperLogLog changes.
sidebar:
  badge:
    text: Reactive
    variant: success
---

The `PFCOUNT.WATCH` command is a novel feature designed to provide real-time updates to clients whenever the cardinality of a HyperLogLog changes. This enables clients to track approximate set cardinality in real-time as new data is added or merged.

## Protocol Support

| Protocol  | Supported |
| --------- | --------- |
| TCP-RESP  | ✅        |
| HTTP      | ❌        |
| WebSocket | ❌        |

## Syntax

```bash
PFCOUNT.WATCH <key>
```

## Parameters

| Parameter | Description                                            | Type   | Required |
| --------- | ------------------------------------------------------ | ------ | -------- |
| `key`     | The key of the HyperLogLog the client wants to monitor | String | Yes      |

## Return Value

| Condition             | Return Value                                                      |
| --------------------- | ----------------------------------------------------------------- |
| Command is successful | Returns a message indicating subscription, similar to `subscribe` |

## Behavior

- The client establishes a subscription to the specified HyperLogLog key.
- The current cardinality of the HyperLogLog is sent to the client.
- DiceDB continuously monitors the specified key for changes in its cardinality.
- Updates are triggered by operations such as `PFADD` and `PFMERGE` that affect the cardinality.
- Whenever the cardinality of the HyperLogLog changes, the updated value is sent to the client.

## Errors

1. `Missing Key`
   - Error Message: `(error) ERROR wrong number of arguments for 'pfcount.watch' command`
   - Occurs if no key is provided.

## Example Usage

### Basic Usage

Here's an example of using the `PFCOUNT.WATCH` command to monitor the approximate cardinality of a HyperLogLog in real-time.

```bash
127.0.0.1:7379> PFCOUNT.WATCH users:hll
Press Ctrl+C to exit watch mode.
0
```

This query does the following:

- Monitors the HyperLogLog key `users:hll`.

When the HyperLogLog is updated using the following commands from another client:

```bash
127.0.0.1:7379> PFADD users:hll "user1"
OK
127.0.0.1:7379> PFADD users:hll "user2" "user3"
OK
127.0.0.1:7379> PFADD users:hll "user4"
OK
127.0.0.1:7379> PFADD other:hll "user5"
OK
127.0.0.1:7379> PFMERGE users:hll users:hll other:hll
OK
```

The subscribing client will receive messages similar to the following:

```bash
127.0.0.1:7379> PFCOUNT.WATCH users:hll
Press Ctrl+C to exit watch mode.
0
1
3
4
5
```

- The `PFADD` commands add elements to the `users:hll` key, incrementally increasing the cardinality.
- The `PFMERGE` command merges `users:hll` with `other:hll`, updating the cardinality of `users:hll` based on the union of both HyperLogLogs.

## Notes

Use the `PFCOUNT.UNWATCH` command to unsubscribe from a HyperLogLog key. This will stop the client from receiving updates on the cardinality. Refer to the [PFCOUNT.UNWATCH](/commands/pfcountunwatch) command documentation for more details.

## Related Commands

The following commands are related to `PFCOUNT.WATCH`:

- [PFCOUNT.UNWATCH](/commands/pfcountunwatch)
- [ZRANGE.WATCH](/commands/zrangewatch)
- [GET.WATCH](/commands/getwatch)
- [PFMERGE](/commands/pfmerge)
