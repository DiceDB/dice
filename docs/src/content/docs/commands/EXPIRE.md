---
title: EXPIRE
description: The `EXPIRE` command in DiceDB is used to set a timeout on a specified key. After the timeout has expired, the key will automatically be deleted. This command is useful for implementing time-based expiration of keys, ensuring that data does not persist longer than necessary.
---

The `EXPIRE` command in DiceDB is used to set a timeout on a specified key. After the timeout has expired, the key will automatically be deleted. This command is useful for implementing time-based expiration of keys, ensuring that data does not persist longer than necessary.

## Syntax

```
EXPIRE key seconds
```

## Parameters

- `key`: The key on which the timeout is to be set. This must be an existing key in the DiceDB database.
- `seconds`: The timeout duration in seconds. This must be a positive integer.

## Return Value

The `EXPIRE` command returns an integer value:

- `1` if the timeout was successfully set.
- `0` if the key does not exist or the timeout could not be set.

## Behaviour

When the `EXPIRE` command is issued:

1. If the specified key exists, DiceDB sets a timeout on the key. The key will be automatically deleted after the specified number of seconds.
2. If the key does not exist, no timeout is set, and the command returns `0`.
3. If the key already has a timeout, the existing timeout is replaced with the new one specified by the `EXPIRE` command.

## Error Handling

The `EXPIRE` command can raise errors in the following scenarios:

- `Wrong Type Error`: If the key exists but is not a string, list, set, hash, or zset, DiceDB will return an error.
- `Syntax Error`: If the command is not used with the correct number of arguments, DiceDB will return a syntax error.

## Example Usage

### Setting a Timeout on a Key

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> EXPIRE mykey 10
```

In this example, the key `mykey` is set with the value "Hello". The `EXPIRE` command sets a timeout of 10 seconds on `mykey`. After 10 seconds, `mykey` will be automatically deleted.

### Checking if Timeout was Set

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> EXPIRE mykey 10
(integer) 1
```

The command returns `1`, indicating that the timeout was successfully set.

### Attempting to Set Timeout on a Non-Existent Key

```bash
127.0.0.1:7379> EXPIRE non_existent_key 10
(integer) 0
```

The command returns `0`, indicating that the key does not exist and no timeout was set.

### Replacing an Existing Timeout

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> EXPIRE mykey 10
127.0.0.1:7379> EXPIRE mykey 20
```

The initial `EXPIRE` command sets a timeout of 10 seconds. The subsequent `EXPIRE` command replaces the existing timeout with a new timeout of 20 seconds.

## Error Handling Examples

### Wrong Type Error

```bash
127.0.0.1:7379> LPUSH mylist "Hello"
127.0.0.1:7379> EXPIRE mylist 10
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, `mylist` is a list, and attempting to set an expiration on it will result in a `WRONGTYPE` error.

### Syntax Error

```bash
127.0.0.1:7379> EXPIRE mykey
(error) ERR wrong number of arguments for 'expire' command
```

This example shows a syntax error due to missing the `seconds` argument.

## Additional Notes

- The `EXPIRE` command is often used in conjunction with other commands like `SET`, `GET`, and `DEL` to manage the lifecycle of keys in DiceDB.
- The timeout can be removed by using the `PERSIST` command, which removes the expiration from the key.
- The `TTL` command can be used to check the remaining time to live of a key with an expiration.
