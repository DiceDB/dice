---
title: COPY
description: Documentation for the DiceDB command COPY
---

The `COPY` command in DiceDB is used to create a copy of a key. This command allows you to duplicate the value stored at a specified key to a new key. The new key can be in the same database or a different database. This command is useful for duplicating data without the need to retrieve and reinsert it manually.

## Syntax

```plaintext
COPY source destination [DB destination-db] [REPLACE]
```

## Parameters

- `source`: The key of the value you want to copy. This key must exist.
- `destination`: The key where the value will be copied to. This key must not exist unless the `REPLACE` option is specified.
- `DB destination-db`: (Optional) The database number where the destination key will be created. If not specified, the destination key will be created in the same database as the source key.
- `REPLACE`: (Optional) If specified, the command will overwrite the destination key if it already exists.

## Return Value

- `Integer`: Returns `1` if the key was copied successfully, and `0` if the key was not copied.

## Behaviour

When the `COPY` command is executed, DiceDB will:

1. Check if the source key exists. If it does not, the command will return `0`.
2. Check if the destination key exists. If it does and the `REPLACE` option is not specified, the command will return `0`.
3. If the `DB destination-db` option is specified, DiceDB will switch to the specified database for the destination key.
4. Copy the value from the source key to the destination key.
5. Return `1` if the copy operation was successful.

## Error Handling

The `COPY` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the source key holds a value that is not compatible with the `COPY` operation.
- `ERR no such key`: This error occurs if the source key does not exist.
- `ERR target key name is busy`: This error occurs if the destination key already exists and the `REPLACE` option is not specified.

## Example Usage

### Basic Copy

Copy the value from `key1` to `key2` in the same database.

```plaintext
COPY key1 key2
```

### Copy with REPLACE

Copy the value from `key1` to `key2`, replacing `key2` if it already exists.

```plaintext
COPY key1 key2 REPLACE
```

### Copy to a Different Database

Copy the value from `key1` in the current database to `key2` in database 2.

```plaintext
COPY key1 key2 DB 2
```

### Copy to a Different Database with REPLACE

Copy the value from `key1` in the current database to `key2` in database 2, replacing `key2` if it already exists.

```plaintext
COPY key1 key2 DB 2 REPLACE
```

## Notes

- The `COPY` command is available starting from DiceDB version 6.2.
- The `COPY` command is atomic, meaning that the copy operation is performed as a single, indivisible operation.
- The `COPY` command does not modify the source key; it only duplicates its value to the destination key.

By understanding the `COPY` command and its parameters, you can effectively duplicate keys within your DiceDB databases, ensuring data consistency and reducing the need for manual data manipulation.

