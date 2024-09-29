---
title: RENAME
description: Documentation for the DiceDB command RENAME
---

The `RENAME` command in DiceDB is used to change the name of an existing key to a new name. If the new key name already exists, it will be overwritten. This command is useful for renaming keys in a DiceDB database without having to delete and recreate them.

## Syntax

```plaintext
RENAME oldkey newkey
```

## Parameters

- `oldkey`: The current name of the key you want to rename. This key must exist in the DiceDB database.
- `newkey`: The new name for the key. If a key with this name already exists, it will be overwritten.

## Return Value

- `Simple String Reply`: Returns `OK` if the key was successfully renamed.

## Behaviour

When the `RENAME` command is executed, the following sequence of events occurs:

1. DiceDB checks if the `oldkey` exists.
2. If `oldkey` does not exist, an error is returned.
3. If `newkey` already exists, it is deleted.
4. The `oldkey` is renamed to `newkey`.
5. The command returns `OK` to indicate success.

## Error Handling

The `RENAME` command can raise the following errors:

- `(error) ERR no such key`: This error is returned if the `oldkey` does not exist in the database.
- `(error) ERR syntax error`: This error is returned if the command is not used with exactly two arguments.

## Example Usage

### Basic Example

```plaintext
SET mykey "Hello"
RENAME mykey mynewkey
GET mynewkey
```

`Explanation:`

1. `SET mykey "Hello"`: Sets the value of `mykey` to "Hello".
2. `RENAME mykey mynewkey`: Renames `mykey` to `mynewkey`.
3. `GET mynewkey`: Retrieves the value of `mynewkey`, which should be "Hello".

### Overwriting an Existing Key

```plaintext
SET key1 "Value1"
SET key2 "Value2"
RENAME key1 key2
GET key2
```

`Explanation:`

1. `SET key1 "Value1"`: Sets the value of `key1` to "Value1".
2. `SET key2 "Value2"`: Sets the value of `key2` to "Value2".
3. `RENAME key1 key2`: Renames `key1` to `key2`, overwriting the existing `key2`.
4. `GET key2`: Retrieves the value of `key2`, which should now be "Value1".

### Error Example

```plaintext
RENAME nonexistingkey newkey
```

`Explanation:`

1. `RENAME nonexistingkey newkey`: Attempts to rename `nonexistingkey` to `newkey`.
2. Since `nonexistingkey` does not exist, the command returns an error: `(error) ERR no such key`.

## Best Practices

- `Check Key Existence`: Before renaming a key, ensure that the `oldkey` exists to avoid errors.
- `Atomic Operations`: The `RENAME` command is atomic, meaning it is executed as a single, indivisible operation. This ensures that no other commands can interfere with the renaming process.
- `Avoid Overwriting`: Be cautious when renaming keys to names that already exist, as this will overwrite the existing key and its value.
