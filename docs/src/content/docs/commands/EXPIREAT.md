---
title: EXPIREAT
description: The `EXPIREAT` command is used to set the expiration time of a key in DiceDB. Unlike the `EXPIRE` command, which sets the expiration time in seconds from the current time, `EXPIREAT` sets the expiration time as an absolute Unix timestamp (in seconds). This allows for more precise control over when a key should expire.
---

The `EXPIREAT` command is used to set the expiration time of a key in DiceDB. Unlike the `EXPIRE` command, which sets the expiration time in seconds from the current time, `EXPIREAT` sets the expiration time as an absolute Unix timestamp (in seconds). This allows for more precise control over when a key should expire.

## Syntax

```bash
EXPIREAT key timestamp [NX|XX|GT|LT]
```

## Parameters

| Parameter   | Description                                                                                     | Type    | Required |
| ----------- | ----------------------------------------------------------------------------------------------- | ------- | -------- |
| `key`       | The name of the key to be set.                                                                  | String  | Yes      |
| `timestamp` | The unix-time-seconds timestamp at which the key should expire. This is an integer.             | Integer | Yes      |
| `NX`        | Set the expiration only if the key does not already have an expiration time.                    | None    | No       |
| `XX`        | Set the expiration only if the key already has an expiration time.                              | None    | No       |
| `GT`        | Set the expiration only if the new expiration time is greater than or equal to the current one. | None    | No       |
| `LT`        | Set the expiration only if the new expiration time is less than the current one.                | None    | No       |

## Return Values

| Condition                                                              | Return Value |
| ---------------------------------------------------------------------- | ------------ |
| Timeout was successfully set.                                          | `1`          |
| Timeout was not set (e.g., key does not exist, or conditions not met). | `0`          |

## Behaviour

- When the `EXPIREAT` command is executed, DiceDB will set the expiration time of the specified key to the given Unix timestamp.
- If the key already has an expiration time, it will be overwritten with the new timestamp.
- If the key does not exist, no timeout is set, and the command returns `0`.
- Conditional flags (NX, XX, GT, LT) control when the expiry can be set based on existing timeouts.

## Errors

1. `Syntax Error`:

   - Error Message: `(error) ERROR wrong number of arguments for 'expireat' command`
   - Returned if the command is issued with an incorrect number of arguments.

2. `Invalid Timestamp`:

   - Error Message: `(error) ERROR value is not an integer or out of range`
   - Returned if the timestamp is not a valid integer.

3. `Invalid Unix Timestamp Format`:
   - Error Message: `(error) ERROR invalid expire time in 'EXPIREAT' command`
   - Returned if the timestamp is outside the supported range.

## Example Usage

### Basic Usage

This example demonstrates setting a key to expire at a specific Unix timestamp.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIREAT mykey 1728212687
(integer) 1
127.0.0.1:7379> TTL mykey
(integer) 15553913293
```

### Key does not exist

Trying to set an expiration time for a non-existing key.

```bash
127.0.0.1:7379> EXPIREAT nonexistingkey 1728212687
(integer) 0
```

### Setting an EXPIRYTIME with NX option

Here, the `NX` option is used to set the expiration time only if the key does not already have an expiration time.

```bash
127.0.0.1:7379> SET key value
OK
127.0.0.1:7379> EXPIREAT key 1728212987222 NX
(integer) 1
127.0.0.1:7379> EXPIREAT key 1728212987222 NX
(integer) 0
```

### Setting an EXPIRYTIME with XX option

Here, the `XX` option is used to set the expiration time only if the key already has an expiration time.

```bash
127.0.0.1:7379> SET key value
OK
127.0.0.1:7379> EXPIREAT key 12345677777 XX
(integer) 0
127.0.0.1:7379> EXPIREAT key 12345677777
(integer) 1
127.0.0.1:7379> EXPIREAT key 123456777722 XX
(integer) 1
```

### Setting an EXPIRYTIME with GT option

The `GT` option is used to set the expiration time only if the new expiration time is greater than or equal to the current one.

```bash
127.0.0.1:7379> SET key value
OK
127.0.0.1:7379> EXPIREAT key 12334444444
(integer) 1
127.0.0.1:7379> EXPIREAT key 12334444424 GT
(integer) 0
127.0.0.1:7379> EXPIREAT key 12334444524 GT
(integer) 1
```

### Setting an EXPIRYTIME with LT option

Similar to the `GT` option, the `LT` option is used to set the expiration time only if the new expiration time is less than (or equal to) the current one.

```bash
127.0.0.1:7379> SET key value
OK
127.0.0.1:7379> EXPIREAT key 12334444444
(integer) 1
127.0.0.1:7379> EXPIREAT key 12334444445 LT
(integer) 0
127.0.0.1:7379> EXPIREAT key 12334444442 LT
(integer) 1
```

## Best Practices

- Use [`TTL`](/commands/ttl) command to check remaining time before expiration
- Consider using [`PERSIST`](/commands/persist) command to remove expiration if needed
- Choose appropriate conditional flags (NX, XX, GT, LT) based on your use case
- Ensure Unix timestamps are in seconds, not milliseconds
- Be aware of the timestamp limit of [9223372036854775]

## Alternatives

- Use [`EXPIRE`](/commands/expire) command for simpler expiration control based on relative time