---
title: PERSIST
description: Documentation for the DiceDB command PERSIST
---

The `PERSIST` command is used to remove the expiration from a key in DiceDB. If a key is set to expire after a certain amount of time, using the `PERSIST` command will make the key persistent, meaning it will no longer have an expiration time and will remain in the database until explicitly deleted.

## Syntax

```
PERSIST key
```

## Parameters

- `key`: The key for which the expiration should be removed. This is a required parameter and must be a valid key in the DiceDB database.

## Return Value

- `Integer reply`: The command returns an integer:
  - `1` if the timeout was successfully removed.
  - `0` if the key does not exist or does not have an associated timeout.

## Behaviour

When the `PERSIST` command is executed:

1. DiceDB checks if the specified key exists in the database.
2. If the key exists and has an expiration time, the expiration time is removed, making the key persistent.
3. If the key does not exist or does not have an expiration time, no changes are made.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the key exists but is not of a type that supports expiration (e.g., a key holding a stream).
- `(nil)`: This is not an error but an indication that the key does not exist or does not have an expiration time.

## Example Usage

### Example 1: Removing Expiration from a Key

```DiceDB
SET mykey "Hello"
EXPIRE mykey 10
PERSIST mykey
TTL mykey
```

`Explanation`:

1. `SET mykey "Hello"`: Sets the value of `mykey` to "Hello".
2. `EXPIRE mykey 10`: Sets an expiration of 10 seconds on `mykey`.
3. `PERSIST mykey`: Removes the expiration from `mykey`.
4. `TTL mykey`: Returns `-1`, indicating that `mykey` does not have an expiration time.

### Example 2: Attempting to Persist a Non-Existent Key

```DiceDB
PERSIST nonExistentKey
```

`Explanation`:

- The command returns `0` because `nonExistentKey` does not exist in the database.

### Example 3: Persisting a Key Without Expiration

```DiceDB
SET mykey "Hello"
PERSIST mykey
```

`Explanation`:

- The command returns `0` because `mykey` does not have an expiration time set.

## Error Handling

### Case 1: Key Does Not Exist

If the key does not exist, the `PERSIST` command will return `0` and no error will be raised.

### Case 2: Key Does Not Have an Expiration

If the key exists but does not have an expiration time, the `PERSIST` command will return `0` and no error will be raised.

### Case 3: Key Holds the Wrong Kind of Value

If the key exists but holds a value type that does not support expiration (e.g., a stream), the command will raise an error:

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Summary

The `PERSIST` command is a useful tool for managing the lifecycle of keys in a DiceDB database. By removing the expiration from a key, you can ensure that the key remains in the database until explicitly deleted. This command is straightforward but powerful, allowing for greater control over key persistence.

