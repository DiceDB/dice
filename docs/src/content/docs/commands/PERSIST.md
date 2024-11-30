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

| Parameter | Description                     | Type   | Required |
| --------- | ------------------------------- | ------ | -------- |
| `key`     | The name of the key to persist. | String | Yes      |

## Return Value

| Condition                                         | Return Value |
| ------------------------------------------------- | ------------ |
| The timeout was successfully removed              | `1`          |
| The key does not exist or does not have a timeout | `0`          |

## Behaviour

- If the specified key has an expiration time, the `PERSIST` command will remove that expiration, ensuring the key remains in the database.
- If the key does not have an expiration or if the key does not exist, the command returns `0` and no action is taken.
- The value of the key remains unchanged. Only the expiration is affected.

## Errors

1. `Wrong type of value or key`:
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value or one that does not support expiration.
2. `No timeout to persist`:
   - This is not an error but occurs when the key either does not exist or does not have an expiration time. The command will return `0` in such cases.

## Example Usage

### Removing Expiration from a Key

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

1. `SET mykey "Hello"`: Sets the value of `mykey` to "Hello".
2. `EXPIRE mykey 10`: Sets an expiration of 10 seconds on `mykey`.
3. `PERSIST mykey`: Removes the expiration from `mykey`.
4. `TTL mykey`: Returns `-1`, indicating that `mykey` does not have an expiration time.

### Attempting to Persist a Non-Existent Key

```bash
127.0.0.1:7379> PERSIST mykey
(integer) 0
```

**Explanation**:

- The command returns `0` because the key `mykey` does not exist in the database.

### Persisting a Key Without Expiration

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```

```bash
127.0.0.1:7379> PERSIST mykey
(integer) 0
```

**Explanation**:

- The command returns `0` because `mykey` does not have an expiration time set.

## Conclusion

The `PERSIST` command is a useful tool for managing the lifecycle of keys in a DiceDB database. By removing the expiration from a key, you can ensure that the key remains in the database until explicitly deleted. This command is straightforward but powerful, allowing for greater control over key persistence.
