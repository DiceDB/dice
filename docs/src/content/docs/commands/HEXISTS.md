---
title: HEXISTS
description: The `HEXISTS` command in DiceDB checks if a specified field exists within a hash stored at a given key. This command is used to verify the presence of a field within hash data structures, making it essential for conditional logic.
---

The `HEXISTS` command in DiceDB checks if a specified field exists within a hash stored at a given key. This command is used to verify the presence of a field within hash data structures, making it essential for conditional logic.

## Syntax

```bash
HEXISTS key field
```

## Parameters

| Parameter | Description                        | Type   | Required |
| --------- | ---------------------------------- | ------ | -------- |
| `key`     | The name of the key holding a hash | String | Yes      |
| `field`   | The field to check within the hash | String | Yes      |

## Return values

| Condition                                   | Return Value |
| ------------------------------------------- | ------------ |
| If the field exists within the hash         | `1`          |
| If the field does not exist within the hash | `0`          |

## Behaviour

- The `HEXISTS` command checks if the specified `field` exists within the hash stored at `key`.
- If the specified `field` is present in the hash, `HEXISTS` returns `1`.
- If the specified `field` is not present or if `key` does not contain a hash, it returns `0`.
- If `key` does not exist, `HEXISTS` returns `0`.

## Errors

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure, such as a string or list.

2. `Invalid syntax or missing parameter`:

   - Error Message: `(error) ERR syntax error`
   - Occurs if the syntax is incorrect or required parameters (`key` and `field`) are missing.

## Example Usage

### Basic Usage

Checking if field `name` exists in the hash stored at key `user:1001`

```bash
127.0.0.1:7379> HEXISTS user:1001 name
1
```

If the field `name` is not present

```bash
127.0.0.1:7379> HEXISTS user:1001 age
0
```

### Checking non-existent key

If the hash `user:1002` does not exist:

```bash
127.0.0.1:7379> HEXISTS user:1002 name
0
```

## Best Practices

- `Check for Field Existence`: Use `HEXISTS` to check for a field’s existence in conditional logic, especially if subsequent commands depend on the field's presence.

## Alternatives

- [`HGET`](/commands/hget): The `HGET` command retrieves the value of a specified field within a hash. However, unlike `HEXISTS`, it returns `nil` if the field does not exist, rather than a boolean response.

## Notes

- If `key` is not of type hash, consider using commands specifically designed for other data types.

By utilizing the `HEXISTS` command, you can conditionally manage hash data in DiceDB, verifying field presence before performing operations based on field existence.
