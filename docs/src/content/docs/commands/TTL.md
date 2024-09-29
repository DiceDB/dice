---
title: TTL
description: The `TTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key that has an expiration set. The TTL is returned in seconds. This command is useful for understanding how much longer a key will exist before it is automatically deleted by DiceDB.
---

The `TTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key that has an expiration set. The TTL is returned in seconds. This command is useful for understanding how much longer a key will exist before it is automatically deleted by DiceDB.

## Parameters

- `key`: The key for which you want to check the TTL. This parameter is mandatory.

## Return Value

The `TTL` command returns an integer value representing the remaining time to live of the key in seconds. The possible return values are:

- A positive integer: The remaining TTL in seconds.
- `-1`: The key exists but has no associated expiration time.
- `-2`: The key does not exist.

## Behaviour

When the `TTL` command is executed:

1. If the key exists and has an expiration time set, the command returns the remaining time to live in seconds.
2. If the key exists but does not have an expiration time set, the command returns `-1`.
3. If the key does not exist, the command returns `-2`.

## Error Handling

The `TTL` command can raise errors in the following scenarios:

- `Wrong Type Error`: If the key exists but is not a string, list, set, hash, or sorted set, an error will be raised. DiceDB will return an error message similar to `(error) WRONGTYPE Operation against a key holding the wrong kind of value`.

## Example Usage

### Example 1: Key with Expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> EXPIRE mykey 10
127.0.0.1:7379> TTL mykey
(integer) 10
```

In this example, the key `mykey` is set with a value of "Hello" and an expiration time of 10 seconds. The `TTL` command returns `10`, indicating that the key will expire in 10 seconds.

### Example 2: Key without Expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> TTL mykey
(integer) -1
```

Here, the key `mykey` is set with a value of "Hello" but no expiration time is set. The `TTL` command returns `-1`, indicating that the key has no expiration.

### Example 3: Non-Existent Key

```bash
127.0.0.1:7379> TTL non_existent_key
(integer) -2
```

In this example, the key `non_existent_key` does not exist in the database. The `TTL` command returns `-2`, indicating that the key does not exist.

## Error Handling Example

### Example 4: Wrong Type Error

```bash
127.0.0.1:7379> HSET myhash field1 "value1"
127.0.0.1:7379> TTL myhash
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, `myhash` is a hash, not a string, list, set, or sorted set. Attempting to use the `TTL` command on a hash results in a `WRONGTYPE` error.
