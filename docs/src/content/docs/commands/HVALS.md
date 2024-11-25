---
title: HVALS
description: The `HVALS` command in DiceDB retrieves all values in a hash stored at a given key. This command allows you to access only the values within a hash, which is helpful for data inspection or retrieval without needing the field names.
---

The `HVALS` command in DiceDB retrieves all values in a hash stored at a given key. This command allows you to access only the values within a hash, which is helpful for data inspection or retrieval without needing the field names.

## Syntax

```bash
HVALS key
```

## Parameters

| Parameter | Description                        | Type   | Required |
| --------- | ---------------------------------- | ------ | -------- |
| `key`     | The name of the key holding a hash | String | Yes      |

## Return values

| Condition                             | Return Value                    |
| ------------------------------------- | ------------------------------- |
| If the key exists and holds a hash    | Array of values within the hash |
| If the key does not exist or is empty | Empty array `[]`                |

## Behaviour

- The `HVALS` command retrieves all values stored in the hash at the specified `key`, without returning the associated field names.
- If the hash is empty or `key` does not exist, it returns an empty array `[]`.
- If `key` exists but does not contain a hash, an error is returned.

## Errors

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure, such as a string or list.

2. `Missing required parameter`:

   - Error Message: `(error) ERR wrong number of arguments for 'HVALS' command`
   - Occurs if the `key` parameter is missing from the command.

## Example Usage

### Basic Usage

Retrieving all values in the hash stored at key `user:1001`

```bash
127.0.0.1:7379> HVALS user:1001
1) "John Doe"
2) "30"
3) "john@example.com"
```

### Empty hash

If the hash stored at `user:1002` exists but has no fields:

```bash
127.0.0.1:7379> HVALS user:1002
(nil)
```

### Non-existent key

If the hash `user:1003` does not exist:

```bash
127.0.0.1:7379> HVALS user:1003
(nil)
```

## Best Practices

- `Use for Values Only`: Use `HVALS` when only the values within a hash are needed without requiring field names, simplifying value extraction.

## Alternatives

- [`HGETALL`](/commands/hgetall): The `HGETALL` command retrieves all field-value pairs in a hash, providing both names and values.

## Notes

- Ensure `key` is a hash type to avoid errors when using `HVALS`.

Using the `HVALS` command enables efficient access to all values within a hash structure in DiceDB, simplifying data retrieval when field names are unnecessary.
