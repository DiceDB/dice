---
title: HKEYS
description: The `HKEYS` command in DiceDB retrieves all fields in a hash stored at a given key. This command is essential for working with hash data structures, enabling retrieval of all field names for dynamic inspection or iteration.
---

The `HKEYS` command in DiceDB retrieves all fields in a hash stored at a given key. This command is essential for working with hash data structures, enabling retrieval of all field names for dynamic inspection or iteration.

## Syntax

```bash
HKEYS key
```

## Parameters

| Parameter | Description                        | Type   | Required |
| --------- | ---------------------------------- | ------ | -------- |
| `key`     | The name of the key holding a hash | String | Yes      |

## Return values

| Condition                             | Return Value         |
| ------------------------------------- | -------------------- |
| If the key exists and holds a hash    | Array of field names |
| If the key does not exist or is empty | Empty array `[]`     |

## Behaviour

- The `HKEYS` command retrieves all field names within the hash stored at the specified `key`.
- If the hash is empty or the `key` does not exist, it returns an empty array `[]`.
- If `key` exists but does not hold a hash, an error is returned.

## Errors

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure, such as a string or list.

2. `Missing required parameter`:

   - Error Message: `(error) ERR wrong number of arguments for 'HKEYS' command`
   - Occurs if the `key` parameter is missing from the command.

## Example Usage

### Basic Usage

Retrieving all field names in the hash stored at key `user:1001`

```bash
127.0.0.1:7379> HKEYS user:1001
1) "name"
2) "age"
3) "email"
```

### Empty hash

If the hash stored at `user:1002` exists but has no fields:

```bash
127.0.0.1:7379> HKEYS user:1002
(nil)
```

### Non-existent key

If the hash `user:1003` does not exist:

```bash
127.0.0.1:7379> HKEYS user:1003
(nil)
```

## Best Practices

- `Use Before Iterating`: Use `HKEYS` to retrieve field names in dynamic applications where the field names may not be predetermined.

## Alternatives

- [`HGETALL`](/commands/hgetall): The `HGETALL` command retrieves all field-value pairs in a hash as an array, rather than only the field names.

## Notes

- Ensure that `key` is of type hash before using `HKEYS`, as other data types will produce errors.

Using the `HKEYS` command, you can efficiently access all field names in hash structures, making it a valuable tool for dynamic data inspection in DiceDB.
