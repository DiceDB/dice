---
title: TTL
description: The `TTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key that has an expiration set. The TTL is returned in seconds. This command is useful for understanding how much longer a key will exist before it is automatically deleted by DiceDB.
---

The `TTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key that has an expiration set. The TTL is returned in seconds. This command is useful for understanding how much longer a key will exist before it is automatically deleted by DiceDB.

## Syntax

```bash
TTL key
```

## Parameters

| Parameter | Description                                  | Type   | Required |
| --------- | -------------------------------------------- | ------ | -------- |
| `key`     | The key for which you want to check the TTL. | String | Yes      |

## Return values

| Condition                                            | Return Value       |
| ---------------------------------------------------- | ------------------ |
| The remaining TTL in seconds                         | A positive integer |
| The key exists but has no associated expiration time | `-1`               |
| The key does not exist.                              | `-2`               |

## Behaviour

- If the `key` exists and has an expiration time set, the command returns the remaining time to live in seconds.
- If the `key` exists but does not have an expiration time set, the command returns `-1`.
- If the `key` does not exist, the command returns `-2`.

## Errors

1. `Wrong number of arguments`:
   - Error Message: `(error) ERR wrong number of arguments for 'ttl' command`
   - Occurs when attempting to use the command with incorrect number of arguments

## Example Usage

### Check TTL for key with expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> EXPIRE mykey 10
127.0.0.1:7379> TTL mykey
(integer) 10
```

In this example, a key `mykey` is created with a value "Hello". Then, an expiration time of 10 seconds is set using the `EXPIRE` command. When `TTL` is called on `mykey`, it returns 10, indicating that the key will expire in 10 seconds.

### Check TTL for key without expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> TTL mykey
(integer) -1
```

In this example, the key `mykey` is set with a value of "Hello" but no expiration is set. The `TTL` command returns `-1`, indicating that the key exists but has no associated expiration.

### Check TTL for non-existent key

```bash
127.0.0.1:7379> TTL non_existent_key
(integer) -2
```

In this example, the key `non_existent_key` does not exist in the DiceDB database. The `TTL` command returns `-2`, indicating that the key does not exist.

### Invalid usage

```bash
127.0.0.1:7379> SET newkey "value"
127.0.0.1:7379> TTL newkey value
(error) ERR wrong number of arguments for 'ttl' command
```

In this example, the `TTL` command is used with an extra argument. This results in an error, as the `TTL` command accepts only one argument.

## Best Practices

- Use `TTL` in conjunction with [`EXPIRE`](/commands/expire) or [`EXPIREAT`](/commands/expireat) commands to manage key expiration effectively

## Alternatives

- [`PTTL`](/commands/pttl): Similar to `TTL` but returns the time-to-live in milliseconds instead of seconds