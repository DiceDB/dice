---
title: PTTL
description: The `PTTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key in milliseconds. This command is particularly useful for understanding how much longer a key will exist before it expires. If the key does not have an associated expiration, the command will return a specific value indicating this state.
---

The `PTTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key in milliseconds. This command is particularly useful for understanding how much longer a key will exist before it expires. If the key does not have an associated expiration, the command will return a specific value indicating this state.

## Syntax

```
PTTL key
```

## Parameters

| Parameter       | Description                                                              | Type    | Required |
|-----------------|--------------------------------------------------------------------------|---------|----------|
| `key`           | The key for which the remaining TTL is to be retrieved.                  | String  | Yes      |

## Return values

| Condition                                                  | Return Value      |
|------------------------------------------------------------|-------------------|
| Command is successful                                      | Returns a positive integer value representing the remaining time to live of the key in milliseconds |
| The key exists but has no associated expiration            | `-1`              |
| The key does not exist                                     | `-2`              |


## Behaviour



## Errors

1. `Wrong number of arguments`:

   - Error Message: `ERROR wrong number of arguments for 'pttl' command`
   - Occures when attempting to use this command without a key.

2. `Invalid key type`: If the key is not a string, DiceDB will return an error.

   - `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`

## Example Usage

### Example 1: Key with TTL

```plaintext
SET mykey "Hello"
EXPIRE mykey 5000
PTTL mykey
```

`Output:`

```plaintext
(integer) 5000
```

In this example, the key `mykey` is set with a value of "Hello" and an expiration of 5000 milliseconds. The `PTTL` command returns `5000`, indicating that the key will expire in 5000 milliseconds.

### Example 2: Key without TTL

```plaintext
SET mykey "Hello"
PTTL mykey
```

`Output:`

```plaintext
(integer) -1
```

In this example, the key `mykey` is set with a value of "Hello" but no expiration is set. The `PTTL` command returns `-1`, indicating that the key exists but has no associated expiration.

### Example 3: Non-existent Key

```plaintext
PTTL nonExistentKey
```

`Output:`

```plaintext
(integer) -2
```

In this example, the key `nonExistentKey` does not exist in the DiceDB database. The `PTTL` command returns `-2`, indicating that the key does not exist.
