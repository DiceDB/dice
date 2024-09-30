---
title: EXPIRETIME
description: Documentation for the DiceDB command EXPIRETIME
---

The `EXPIRETIME` command in DiceDB is used to retrieve the absolute Unix timestamp (in seconds) at which a given key will expire. This command is particularly useful for understanding the exact expiration time of a key, which can help in debugging and managing key lifetimes.

## Syntax

```
EXPIRETIME key
```

## Parameters

- `key`: The key for which you want to retrieve the expiration time. This parameter is mandatory and should be a valid key in the DiceDB database.

## Return Value

- `Integer`: The command returns the absolute Unix timestamp (in seconds) at which the key will expire.
- `Integer (Special Case)`: If the key does not exist or does not have an associated expiration time, the command returns `-1`.

## Behaviour

When the `EXPIRETIME` command is executed:

1. DiceDB checks if the specified key exists in the database.
2. If the key exists and has an associated expiration time, DiceDB returns the absolute Unix timestamp (in seconds) at which the key will expire.
3. If the key does not exist or does not have an associated expiration time, DiceDB returns `-1`.

## Error Handling

The `EXPIRETIME` command can raise errors in the following scenarios:

1. `Wrong number of arguments`: If the command is called with an incorrect number of arguments, DiceDB will return an error message:
   ```
   ERR wrong number of arguments for 'expiretime' command
   ```
2. `Invalid key type`: If the key is not a valid string, DiceDB will return an error message:
   ```
   ERR invalid key type
   ```

## Example Usage

### Example 1: Key with Expiration Time

```DiceDB
SET mykey "Hello"
EXPIRE mykey 60
EXPIRETIME mykey
```

`Output:`

```
(integer) 1672531199
```

In this example, the key `mykey` is set with a value "Hello" and an expiration time of 60 seconds. The `EXPIRETIME` command returns the Unix timestamp at which `mykey` will expire.

### Example 2: Key without Expiration Time

```DiceDB
SET mykey "Hello"
EXPIRETIME mykey
```

`Output:`

```
(integer) -1
```

In this example, the key `mykey` is set with a value "Hello" but no expiration time is set. The `EXPIRETIME` command returns `-1` indicating that the key does not have an associated expiration time.

### Example 3: Non-Existent Key

```DiceDB
EXPIRETIME nonExistentKey
```

`Output:`

```
(integer) -1
```

In this example, the key `nonExistentKey` does not exist in the database. The `EXPIRETIME` command returns `-1` indicating that the key does not exist.
