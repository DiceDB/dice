---
title: SELECT
description: Documentation for the DiceDB command SELECT
---

The `SELECT` command is used to switch the currently selected database for the current connection in DiceDB. By default, DiceDB starts with database 0, but it supports multiple databases, which can be accessed by using the `SELECT` command. This command is essential for managing data across different logical databases within a single DiceDB instance.

## Parameters

### index

- `Type`: Integer
- `Description`: The zero-based index of the database to select. DiceDB databases are indexed starting from 0 up to a configurable maximum (default is 15, configurable via the `databases` configuration directive).
- `Constraints`: Must be a non-negative integer within the range of available databases.

## Return Value

- `Type`: Simple String
- `Description`: Returns `OK` if the database switch was successful.

## Behaviour

When the `SELECT` command is issued, the current connection's context is switched to the specified database. All subsequent commands will operate on the selected database until another `SELECT` command is issued or the connection is closed.

- `Initial State`: By default, the connection starts with database 0.
- `Post-Command State`: The connection will be associated with the specified database index.

## Error Handling

The `SELECT` command can raise errors under the following conditions:

1. `Invalid Database Index`:

   - `Error`: `(error) ERR DB index is out of range`
   - `Condition`: If the specified database index is outside the range of available databases.

2. `Non-Integer Index`:

   - `Error`: `(error) ERR value is not an integer or out of range`
   - `Condition`: If the provided index is not a valid integer.

## Example Usage

### Switching to Database 1

```shell
127.0.0.1:6379> SELECT 1
OK
```

In this example, the connection switches to database 1. All subsequent commands will operate on database 1.

### Switching to Database 0

```shell
127.0.0.1:6379> SELECT 0
OK
```

Here, the connection switches back to the default database 0.

### Error Example: Invalid Database Index

```shell
127.0.0.1:6379> SELECT 16
(error) ERR DB index is out of range
```

In this example, an error is raised because the specified database index 16 is outside the default range of 0-15.

### Error Example: Non-Integer Index

```shell
127.0.0.1:6379> SELECT one
(error) ERR value is not an integer or out of range
```

In this example, an error is raised because the provided index is not a valid integer.

## Notes

- The number of databases is configurable in the DiceDB configuration file (`DiceDB.conf`) using the `databases` directive.
- Switching databases does not affect the data stored in other databases; it only changes the context for the current connection.
- The `SELECT` command is connection-specific. Different connections can operate on different databases simultaneously.

By understanding and using the `SELECT` command, you can efficiently manage and segregate data across multiple logical databases within a single DiceDB instance.

