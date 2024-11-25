---
title: PFCOUNT.UNWATCH
description: The `PFCOUNT.UNWATCH` command is a feature to stop receiving updates on a HyperLogLog.
sidebar:
  badge:
    text: Reactive
    variant: success
---

The `PFCOUNT.UNWATCH` command is a feature to stop receiving updates on a HyperLogLog.

## Protocol Support

| Protocol  | Supported |
| --------- | --------- |
| TCP-RESP  | ✅        |
| HTTP      | ❌        |
| WebSocket | ❌        |

## Syntax

```bash
PFCOUNT.UNWATCH <fingerprint>
```

## Parameters

| Parameter     | Description                                               | Type   | Required |
| ------------- | --------------------------------------------------------- | ------ | -------- |
| `fingerprint` | Fingerprint returned as part of the `PFCOUNT.WATCH` query | String | Yes      |

## Return Value

| Condition             | Return Value |
| --------------------- | ------------ |
| Command is successful | `OK`         |

## Behavior

- The client unsubscribes from the HyperLogLog key specified by the fingerprint.

## Errors

1. `Missing fingerprint`
   - Error Message: `(error) ERROR wrong number of arguments for 'pfcount.unwatch' command`
   - Occurs if no fingerprint is provided.

## Example Usage

### Basic Usage

Here’s an example of using the `PFCOUNT.WATCH` and `PFCOUNT.UNWATCH` commands to monitor and stop monitoring the cardinality of a HyperLogLog.

```bash
127.0.0.1:7379> PFCOUNT.WATCH users:hll
Press Ctrl+C to exit watch mode.

```

When the HyperLogLog is updated using the following commands from another client:

```bash
127.0.0.1:7379> PFADD users:hll "user1"
OK
127.0.0.1:7379> PFADD users:hll "user2"
OK
127.0.0.1:7379> PFADD users:hll "user3"
OK
```

The subscribing client will receive messages similar to the following:

```bash
127.0.0.1:7379> PFCOUNT.WATCH users:hll
Press Ctrl+C to exit watch mode.
1
2
3
```

To stop receiving updates on the key, use the `PFCOUNT.UNWATCH` command:

```bash
127.0.0.1:7379> PFCOUNT.UNWATCH 1298365423
OK
```

## Notes

Use the `PFCOUNT.WATCH` command to subscribe to a HyperLogLog key and receive real-time updates. Refer to the [PFCOUNT.WATCH](/commands/pfcountwatch) command documentation for more details.

## Related Commands

The following commands are related to `PFCOUNT.UNWATCH`:

- [ZRANGE.WATCH](/commands/zrangewatch)
- [PFCOUNT.WATCH](/commands/pfcountwatch)
- [GET.UNWATCH](/commands/getunwatch)
