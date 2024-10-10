---
title: EXPIREAT
description: The `EXPIREAT` command is used to set the expiration time of a key in DiceDB. Unlike the `EXPIRE` command, which sets the expiration time in seconds from the current time, `EXPIREAT` sets the expiration time as an absolute Unix timestamp (in seconds). This allows for more precise control over when a key should expire.
---

The `EXPIREAT` command is used to set the expiration time of a key in DiceDB. Unlike the `EXPIRE` command, which sets the expiration time in seconds from the current time, `EXPIREAT` sets the expiration time as an absolute Unix timestamp (in seconds). This allows for more precise control over when a key should expire.

## Syntax

```plaintext
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

## Return values

| Condition                                          | Return Value |
| -------------------------------------------------- | ------------ |
| Command is successful                              | `1`          |
| Key does not exist or the timeout could not be set | `0`          |
| Syntax or specified constraints are invalid        | error        |

## Behaviour

- When the `EXPIREAT` command is executed, DiceDB will set the expiration time of the specified key to the given Unix timestamp.
- If the key already has an expiration time, it will be overwritten with the new timestamp.
- If the key does not exist, the command will return `0` and no expiration time will be set.

## Errors

### Wrong number of arguments

When the `EXPIREAT` command is called with the wrong number of arguments, an error is returned.

```bash
127.0.0.1:7379> EXPIREAT testkey1
(error) ERROR wrong number of arguments for 'EXPIREAT' command
```

### Invalid timestamp

When the provided timestamp is not a valid integer, an error is returned.

```bash
127.0.0.1:7379> EXPIREAT testkey1 17282112781a
(error) ERROR value is not an integer or out of range
```

### Invalid format of Unix Timestamp

When the provided timestamp is not a valid Unix timestamp, or is outside the supported range, an error is returned.

```bash
127.0.0.1:7379> EXPIREAT testkey1 11111111111111111
(error) ERROR invalid expire time in 'EXPIREAT' command
```

## Example Usage

### Setting an Expiration Time

Setting a key `mykey` to expire at the Unix timestamp `17282126871`.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIREAT mykey 17282126871
(integer) 1
```

### Checking the expiration time

Checking the remaining time to live of a key (in seconds).

```bash
127.0.0.1:7379> EXPIREAT mykey 17282126871
(integer) 1
127.0.0.1:7379> TTL mykey
(integer) 15553913293
```

### Key does not exist

Trying to set an expiration time for a non-existing key.

```bash
127.0.0.1:7379> EXPIREAT nonexistingkey 17282126871
(integer) 0
```

### Setting an EXPIRYTIME only if not exists

Here, the `NX` option is used to set the expiration time only if the key does not already have an expiration time.

```bash
127.0.0.1:7379> SET key value
OK
127.0.0.1:7379> EXPIREAT key 1728212987222 NX
(integer) 1
127.0.0.1:7379> EXPIREAT key 1728212987222 NX
(integer) 0
```

### Setting an EXPIRYTIME only if it already has one

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

### Setting an EXPIRYTIME only if the new expiry time is greater than or equal to the current one

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

### Setting an EXPIRYTIME only if the new expiry time is less than or equal to the current one

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

## Additional notes

- The `EXPIREAT` command is useful when you need to synchronize the expiration of keys across multiple DiceDB instances or when you need to set an expiration time based on an external event that provides a Unix timestamp.
- The timestamp should be in seconds. If you have a timestamp in milliseconds, you need to convert it to seconds before using it with `EXPIREAT`.
- There is an arbritrary limit to the size of the `unix-time-seconds` of [9223372036854775](https://github.com/DiceDB/dice/blob/b74dc8ffd5e518eaa9b82020d2b25a592c6472d4/internal/eval/eval.go#L69).

## Related commands

- `EXPIRE`: Sets the expiration time of a key in seconds from the current time.
- `PEXPIREAT`: Sets the expiration time of a key as an absolute Unix timestamp in milliseconds.
- `TTL`: Returns the remaining time to 127.0.0.1:7379 of a key in seconds.
- `PTTL`: Returns the remaining time to live of a key in milliseconds.

By understanding and using the `EXPIREAT` command, you can effectively manage the lifecycle of keys in your DiceDB database, ensuring that data is available only as long as it is needed.
