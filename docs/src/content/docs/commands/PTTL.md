---
title: PTTL
description: Documentation for the DiceDB command PTTL
---

The `PTTL` command in DiceDB is used to retrieve the remaining time to live (TTL) of a key in milliseconds. This command is particularly useful for understanding how much longer a key will exist before it expires. If the key does not have an associated expiration, the command will return a specific value indicating this state.

## Syntax

```plaintext
PTTL key
```

## Parameters

- `key`: The key for which the remaining TTL is to be retrieved. This parameter is mandatory and must be a valid key in the DiceDB database.

## Return Value

The `PTTL` command returns an integer value representing the remaining time to live of the key in milliseconds. The possible return values are:

- A positive integer: The remaining TTL in milliseconds.
- `-1`: The key exists but has no associated expiration.
- `-2`: The key does not exist.

## Behaviour

When the `PTTL` command is executed, DiceDB checks the specified key to determine its remaining TTL. The command will return the TTL in milliseconds if the key exists and has an expiration. If the key exists but does not have an expiration, the command will return `-1`. If the key does not exist, the command will return `-2`.

## Error Handling

The `PTTL` command can raise errors in the following scenarios:

1. `Wrong number of arguments`: If the command is called without the required number of arguments, DiceDB will return an error.

   - `Error Message`: `ERR wrong number of arguments for 'pttl' command`

2. `Invalid key type`: If the key is not a string, DiceDB will return an error.

   - `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`

## Example Usage

### Example 1: Key with TTL

```plaintext
SET mykey "Hello"
PEXPIRE mykey 5000
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
