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
| Parameter       | Description                                      | Type    | Required |
|-----------------|--------------------------------------------------|---------|----------|
| `key`           | The key for which you want to check the TTL.                   | String  | Yes      |

## Return values

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| The remaining TTL in seconds                         | A positive integer                                              |
| The key exists but has no associated expiration time            | `-1`                                             |
| The key does not exist.    | `-2`                                             |


## Behaviour

- If the `key` exists and has an expiration time set, the command returns the remaining time to live in seconds.
- If the `key` exists but does not have an expiration time set, the command returns `-1`.
- If the `key` does not exist, the command returns `-2`.

## Errors

The `TTL` command can raise errors in the following scenarios:

- `Syntax Error`:

   - (error) ERROR syntax error
   - Occurs when attempting to use the command with more than one argument.

## Examples

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


### Example 4: Invalid usage
```bash
127.0.0.1:7379> SET newkey "value"
127.0.0.1:7379> TTL newkey value
(error) ERROR syntax error
```

- The `TTL` command requires exactly one argument: `key`
- Since only more than one argument is provided, DiceDB returns a syntax error.
