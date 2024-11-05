---
title: HSETNX
description: The `HSETNX` command in DiceDB is used to set the value of a field in a hash only if the field does not already exist. This command is useful for ensuring that a value is only set if it is not already present. 
---

The `HSETNX` command in DiceDB is used to set the value of a field in a hash only if the field does not already exist. This command is useful for ensuring that a value is only set if it is not already present.

## Syntax

```bash
HSETNX key field value
```

## Parameters

| Parameter | Description                                              | Type    | Required |
|-----------|----------------------------------------------------------|---------|----------|
| `key`     | The name of the hash.                                    | String  | Yes      |
| `field`   | The field within the hash to set the value for.          | String  | Yes      |
| `value`   | The value to set for the specified field.                | String  | Yes      |

## Return Values

| Condition                                   | Return Value                                                                |
|---------------------------------------------|-----------------------------------------------------------------------------|
| Field added                                 | `1`                                                                         |
| Field already exists                        | `0`                                                                         |
| Wrong data type                             | `(error) WRONGTYPE Operation against a key holding the wrong kind of value` |
| Incorrect Argument Count                    | `(error) ERR wrong number of arguments for 'hsetnx' command`                |

## Behaviour

When the `HSETNX` command is executed, the following actions occur:

- If the specified hash does not exist, a new hash is created.
- The specified field and value are set in the hash only if the field does not already exist.
- If the field already exists, the command does not modify the existing value and returns `0`.
- The command returns `1` if the field was successfully added to the hash.

## Errors

The `HSETNX` command can raise errors in the following scenarios:

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure.

2. `Incorrect Argument Count`:

   - Error Message: `(error) ERR wrong number of arguments for 'hsetnx' command`
   - Occurs if the command is not provided with the correct number of arguments (i.e., fewer than three).

## Example Usage

### Basic Usage

#### Creating a New Hash with `HSETNX`

```bash
127.0.0.1:7379> HSETNX product:3000 name "Smartphone"
1
```
- **Behaviour**: A new hash is created with the key `product:3000`. The field `name` is set with the value "Smartphone".
- **Return Value**: `1` (since the field was added).

#### Attempting to Set an Existing Field

```bash
127.0.0.1:7379> HSETNX product:3000 name "Tablet"
0
```

- **Behaviour**: The command attempts to set the `name` field to "Tablet", but since `name` already exists in the hash, it does not change the value.
- **Return Value**: `0` (since the field was not added).

### Invalid Usage

Trying to set a field in a key that is not a hash.

```bash
127.0.0.1:7379> SET product:3000 "This is a string"
OK
127.0.0.1:7379> HSETNX product:3000 name "Smartphone"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

- **Behaviour**: The `SET` command sets the key `product:3000` to a string value.
- **Error**: The `HSETNX` command raises a `WRONGTYPE` error because `product:3000` is not a hash.

Wrong Number of Arguments for `HSETNX` Command

```bash
127.0.0.1:7379> HSETNX product:3000
(error) ERR wrong number of arguments for 'hsetnx' command

127.0.0.1:7379> HSETNX product:3000 name
(error) ERR wrong number of arguments for 'hsetnx' command
```
- **Behavior**: The `HSETNX` command requires atleast three arguments: the key, the field name, and the field value.
- **Error**: The command fails because it requires the `key`, `field`, and `value` parameters. If insufficient arguments are provided, DiceDB raises an error indicating that the number of arguments is incorrect.

### Best Practices

- Use `HSETNX` when you need to ensure that a field is only set if it does not already exist, preventing accidental overwrites.
