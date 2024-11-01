---
title: PERSIST
description: Documentation for the DiceDB command PERSIST
---

The `PERSIST` command is used to remove the expiration from a key in DiceDB. If a key is set to expire after a certain amount of time, using the `PERSIST` command will make the key persistent, meaning it will no longer have an expiration time and will remain in the database until explicitly deleted.

## Syntax

```bash
PERSIST key
```

## Parameters

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The name of the key to persist.                                            | String  | Yes      |

## Return Value

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| The timeout was successfully removed           | `1`                                              |
| The key does not exist or does not have a timeout| `0`                                             |

## Behaviour

When the `PERSIST` command is executed:

1. If the specified key has an expiration time, the `PERSIST` command removes that expiration, making the key persistent.
2. If the key does not exist or does not have an expiration, the command does nothing and returns `0`.
3. This command does not alter the key’s value, only its expiration state.

## Error Handling

1. `Wrong type of value or key`:
    - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
    - Occurs when attempting to use the command on a key that contains a non-string value or one that does not support expiration.
2. `No timeout to persist`:
    - This is not an error but occurs when the key either does not exist or does not have an expiration time. The command will return `0` in such cases.    

## Example Usage

### Example 1: Removing Expiration from a Key

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```
```bash
127.0.0.1:7379> EXPIRE mykey 10
(integer) 1
```
```bash
127.0.0.1:7379> PERSIST mykey
(integer) 1
```
```bash
127.0.0.1:7379> TTL mykey
(integer) -1
```

`Explanation`:

1. `SET mykey "Hello"`: Sets the value of `mykey` to "Hello".
2. `EXPIRE mykey 10`: Sets an expiration of 10 seconds on `mykey`.
3. `PERSIST mykey`: Removes the expiration from `mykey`.
4. `TTL mykey`: Returns `-1`, indicating that `mykey` does not have an expiration time.

### Example 2: Attempting to Persist a Non-Existent Key

```bash
127.0.0.1:7379> PERSIST mykey
(integer) 0
```

`Explanation`:

- The command returns `0` because `mykey` does not exist in the database.

### Example 3: Persisting a Key Without Expiration

```bash
127.0.0.1:7379> SET mykey
OK
```

```bash
127.0.0.1:7379> PERSIST mykey
(integer) 0
```

`Explanation`:

- The command returns `0` because `mykey` does not have an expiration time set.


## Summary

The `PERSIST` command is a useful tool for managing the lifecycle of keys in a DiceDB database. By removing the expiration from a key, you can ensure that the key remains in the database until explicitly deleted. This command is straightforward but powerful, allowing for greater control over key persistence.

