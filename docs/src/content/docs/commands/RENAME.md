---
title: RENAME
description: The `RENAME` command in DiceDB allows you to change the name of an existing key. If the new key name already exists, it will be overwritten.
---

The `RENAME` command in DiceDB is used to change the name of an existing key to a new name. If the new key name already exists, it will be overwritten. This command is useful for renaming keys in a DiceDB database without having to delete and recreate them.

## Syntax

```bash
RENAME oldkey newkey
```

## Parameters

| Parameter | Description                                                                                 | Type   | Required |
| --------- | ------------------------------------------------------------------------------------------- | ------ | -------- |
| `oldkey`  | The current name of the key you want to rename. This key must exist in the DiceDB database. | String | Yes      |
| `newkey`  | The new name for the key. If a key with this name already exists, it will be overwritten.   | String | Yes      |

## Return values

| Condition                | Return Value                                                 |
| ------------------------ | ------------------------------------------------------------ |
| Command is successful    | `OK`                                                         |
| `oldkey` does not exist  | `(error) ERR no such key`                                    |
| arguments not equal to 2 | `(error) ERR wrong number of arguments for 'rename' command` |

## Behaviour

- The `RENAME` command renames an existing key (`oldkey`) to a new name (`newkey`).
- If `newkey` already exists, it will be overwritten by the value of `oldkey`.
- If `oldkey` does not exist, an error is returned.
- The operation is atomic, ensuring no other commands can interfere while it executes.
- If the source and destination keys are the same, the command will return `OK` without making changes.

## Errors

1. `Key does not exist`:

   - Error Message: `(error) ERR no such key`
   - Occurs when the specified `oldkey` does not exist in the database.

1. `Wrong number of arguments`:
   - Error Message: `(error) ERR wrong number of arguments for 'rename' command`
   - Occurs if the command is not used with exactly two arguments.

## Example Usage

### Basic Example

Renaming a key from `mykey` to `mynewkey`

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> RENAME mykey mynewkey
OK
127.0.0.1:7379> GET mynewkey
"Hello"
```

### Overwriting an Existing Key

Renaming `key1` to `key2` will overwrite `key2`:

```bash
127.0.0.1:7379> SET key1 "Value1"
OK
127.0.0.1:7379> SET key2 "Value2"
OK
127.0.0.1:7379> RENAME key1 key2
OK
127.0.0.1:7379> GET key2
"Value1"
```

### Error Example: Non-Existent Key

Attempting to rename a non-existing key

```bash
127.0.0.1:7379> RENAME nonexistingkey newkey
(error) ERR no such key
```

### Error Example: Incorrect Number of Arguments

Trying to rename with only one argument

```bash
127.0.0.1:7379> RENAME key1
(error) ERR wrong number of arguments for 'rename' command
```

## Best Practices

- `Check Key Existence`: Before renaming a key, ensure that the `oldkey` exists to avoid errors.
- `Atomic Operations`: The `RENAME` command is atomic, meaning it is executed as a single, indivisible operation. This ensures that no other commands can interfere with the renaming process.
- `Avoid Overwriting`: Be cautious when renaming keys to names that already exist, as this will overwrite the existing key and its value.
