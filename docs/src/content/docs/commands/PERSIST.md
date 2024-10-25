---
title: PERSIST  
description: The `PERSIST` command in DiceDB is used to remove the expiration time from a key, making it persistent. This allows the key to remain in the database indefinitely until it is explicitly deleted.
---

The `PERSIST` command is used to remove the expiration from a key in DiceDB. If a key is set to expire after a certain period, using the `PERSIST` command will make the key persistent, meaning it will no longer have an expiration time and will remain in the database until explicitly deleted.

## Syntax

```bash
PERSIST key
```

## Parameters

| Parameter | Description                      | Type   | Required |
|-----------|----------------------------------|--------|----------|
| `key`     | The name of the key to persist.   | String | Yes      |

## Return Value

| Condition                                    | Return Value                                      |
|----------------------------------------------|---------------------------------------------------|
| Timeout is successfully removed              | `1`                                               |
| Key does not exist or lacks an expiration     | `0`                                               |

## Behaviour

- If the specified key has an expiration time, the `PERSIST` command will remove that expiration, ensuring the key remains in the database.
- If the key does not have an expiration or if the key does not exist, the command returns `0` and no action is taken.
- The value of the key remains unchanged. Only the expiration is affected.

## Error Handling

1. `Wrong type of value or key`:
    - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
    - Occurs when the command is used on a key that holds a non-string value or a type that does not support expiration.

2. `No timeout to persist`:
    - Although this is not classified as an error, the command returns `0` when trying to persist a key that either does not exist or does not have an expiration.

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

**Explanation**:

- The `PERSIST` command removes the expiration from `mykey`, and the `TTL` command confirms that the key no longer has a time-to-live (TTL) value (`-1` indicates no expiration).

### Example 2: Attempting to Persist a Non-Existent Key

```bash
127.0.0.1:7379> PERSIST mykey
(integer) 0
```

**Explanation**:

- The command returns `0` because the key `mykey` does not exist in the database.

### Example 3: Persisting a Key Without Expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```
```bash
127.0.0.1:7379> PERSIST mykey
(integer) 0
```

**Explanation**:

- The command returns `0` because the key `mykey` does not have an expiration time set.

## Summary

The `PERSIST` command is a simple but powerful tool for managing the lifetime of keys in DiceDB. By removing the expiration from a key, you ensure that it remains in the database until explicitly deleted, giving you more control over key persistence in your system.

