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

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The name of the key whose expiration time is to be retrieved              | String  | Yes      |

## Return Value

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| The key exists and has an expiration time      | Unix timestamp (in seconds)                       |
| The key exists but has no expiration time      | -1                                                |
| The key does not exist                         | -2                                                |


## Behaviour

When the `EXPIRETIME` command is executed:

1. DiceDB checks if the specified key exists in the database.
2. If the key exists and has an associated expiration time, DiceDB returns the absolute Unix timestamp (in seconds) at which the key will expire.
3. If the key exists without an expiration time, the command returns `-1`.
4. If the key doesn't exist, the command returns `-2`. 

## Error Handling

The `EXPIRETIME` command can raise errors in the following scenarios:

1. `Wrong number of arguments`: If the command is called with an incorrect number of arguments, DiceDB will return an error message:
   ```
   (error) ERROR wrong number of arguments for 'expiretime' command
   ```
2. `Invalid key type`: If the key is not a valid string, DiceDB will return an error message:
   ```
   (error) ERROR invalid key type
   ```

## Example Usage

### Example 1: Key with Expiration Time

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```
```bash
127.0.0.1:7379> EXPIRE mykey 60
(integer) 1
```
```bash
127.0.0.1:7379> EXPIRETIME mykey
```

`Output:`

```
(integer) 1728548993
```

In this example, the key `mykey` is set with a value "Hello" and an expiration time of 60 seconds. The `EXPIRETIME` command returns the Unix timestamp at which `mykey` will expire.

### Example 2: Key without Expiration Time

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```

```bash
127.0.0.1:7379> EXPIRETIME mykey
```

`Output:`

```
(integer) -1
```

In this example, the key `mykey` is set with a value "Hello" but no expiration time is set. The `EXPIRETIME` command returns `-1` indicating that the key does not have an associated expiration time.

### Example 3: Non-Existent Key

```bash
127.0.0.1:7379> EXPIRETIME nonExistentKey
```

`Output:`

```
(integer) -2
```

In this example, the key `nonExistentKey` does not exist in the database. The `EXPIRETIME` command returns `-2` indicating that the key does not exist.
