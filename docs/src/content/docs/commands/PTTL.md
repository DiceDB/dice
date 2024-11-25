---
title: PTTL
description: The `PTTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key in milliseconds. This command is particularly useful for understanding how much longer a key will exist before it expires. If the key does not have an associated expiration, the command will return a specific value indicating this state.
---

The `PTTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key in milliseconds. This command is particularly useful for understanding how much longer a key will exist before it expires. If the key does not have an associated expiration, the command will return a specific value indicating this state.

## Syntax

```bash
PTTL key
```

## Parameters

| Parameter | Description                                            | Type   | Required |
| --------- | ------------------------------------------------------ | ------ | -------- |
| `key`     | The key for which the remaining TTL is to be retrieved | String | Yes      |

## Return values

| Condition                                   | Return Value                                                                        |
| ------------------------------------------- | ----------------------------------------------------------------------------------- |
| Command is successful                       | Positive integer representing the remaining time to live of the key in milliseconds |
| Key exists but has no associated expiration | `-1`                                                                                |
| Key does not exist                          | `-2`                                                                                |

## Behaviour

- The command is non-destructive and does not modify the key or its expiration in any way.
- If a key's TTL is modified (e.g., by `EXPIRE` or `PEXPIRE` commands), subsequent `PTTL` calls will reflect the updated remaining time.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for 'pttl' command`
   - Occurs when attempting to use this command without any arguments or with more than one argument.

## Example Usage

### Key with Expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> EXPIRE mykey 10
127.0.0.1:7379> PTTL mykey
(integer) 10000
```

In this example, the key `mykey` is set with a value of "Hello" and an expiration of 10 seconds. The `PTTL` command returns `10000`, indicating that the key will expire in 10000 milliseconds.

### Key without Expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> PTTL mykey
(integer) -1
```

In this example, the key `mykey` is set with a value of "Hello" but no expiration is set. The `PTTL` command returns `-1`, indicating that the key exists but has no associated expiration.

### Non-existent Key

```bash
127.0.0.1:7379> PTTL nonExistentKey
(integer) -2
```

In this example, the key `nonExistentKey` does not exist in the DiceDB database. The `PTTL` command returns `-2`, indicating that the key does not exist.

### Invalid usage

```bash
127.0.0.1:7379> SET newkey "value"
127.0.0.1:7379> PTTL newkey value
(error) ERR wrong number of arguments for 'pttl' command
```

In this example, the `PTTL` command is used with an extra argument. This results in an error, as the `PTTL` command accepts only one argument.

## Best Practices

- Use `PPTL` in conjunction with [`EXPIRE`](/commands/expire) or `PEXPIRE` commands to manage key expiration effectively

## Alternatives

- [`TTL`](/commands/ttl): Similar to `PTTL` but returns the time-to-live in seconds instead of milliseconds