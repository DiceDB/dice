---
title: COPY
description: The COPY command in DiceDB is used to create a copy of a key.
---

The `COPY` command in DiceDB is used to create a copy of a key. This command allows you to duplicate the value stored at a specified key to a new key. The new key can be in the same database or a different database. This command is useful for duplicating data without the need to retrieve and reinsert it manually.

## Syntax

```plaintext
COPY source destination [REPLACE]
```

## Parameters
<!-- please add all parameters, small description, type and required, see example for SET command-->
| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `source`     | The name of the key to be set.                                            | String  | Yes      |
| `destination`   | The value to be set for the key.                                          | String  | Yes      |
| `REPLACE`      | (Optional) If specified, the command will overwrite the destination key if it already exists.                                | None | No       |

## Return Value
| Condition | Return Value |
|-----------| :-------------:|
| key was copied successfully | `1` |
| key was not copied | `0` | 

<!-- add all scenarios, see below example for SET -->

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| if copy is successful from source to destination                    | `1`                                              |
| if copy is unsuccessful from source to destination         | `0`                                             |

## Behaviour

When the `COPY` command is executed, DiceDB will:

1. Check if the source key exists. If it does not, the command will return `0`.
2. Check if the destination key exists. If it does and the `REPLACE` option is not specified, the command will return `0`.
3. Copy the value from the source key to the destination key.
4. Return `1` if the copy operation was successful.

## Error Handling

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

2. `Invalid syntax or conflicting options`:

   - Error Message: `(error) ERR target key already exists`
   - This error occurs if the destination key already exists and the `REPLACE` option is not specified.

## Example Usage

### Basic Copy

Copy the value from `key1` to `key2` in the same database.

```bash
127.0.0.1:7379> COPY key1 key2
(integer) 1
```

### Copy with REPLACE

Copy the value from `key1` to `key2`, replacing `key2` if it already exists.

```bash
127.0.0.1:7379> COPY key1 key2 REPLACE
(integer) 1
```

### Invalid usage

Trying to copy value from `key1` to `key2` without `REPLACE` option and `key2` value already exists

```bash
127.0.0.1:7379> COPY key1 key2
(error) ERR target key already exists
```
## Notes

- The `COPY` command is atomic, meaning that the copy operation is performed as a single, indivisible operation.
- The `COPY` command does not modify the source key; it only duplicates its value to the destination key.

By understanding the `COPY` command and its parameters, you can effectively duplicate keys within your DiceDB databases, ensuring data consistency and reducing the need for manual data manipulation.
