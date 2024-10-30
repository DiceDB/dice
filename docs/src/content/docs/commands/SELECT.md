---
title: SELECT
description: Documentation for the DiceDB command SELECT
---

> **Important Note:** As of the current version, DiceDB does not support multiple databases. Therefore, the `SELECT` command is currently a dummy method and does not affect the database. It remains as a placeholder.

The `SELECT` command is used to switch the currently selected database for the current connection in DiceDB. By default, DiceDB starts with database 0, but it supports multiple databases, which can be accessed by using the `SELECT` command. This command is essential for managing data across different logical databases within a single DiceDB instance.

## Syntax

```bash
SELECT index
```

## Parameters

| Parameter | Description                                     | Type    | Required |
|-----------|-------------------------------------------------|---------|----------|
| `index`   | The zero-based index of the database to select. DiceDB databases are indexed starting from 0 up to a configurable maximum (default is 15, configurable via the `databases` configuration directive) | Integer | Yes      |

## Return Value

| Condition                | Return Value |
|--------------------------|--------------|
| Command is successful    | `OK`         |

## Behaviour

When the `SELECT` command is issued, the current connection's context is switched to the specified database. All subsequent commands will operate on the selected database until another `SELECT` command is issued or the connection is closed.

- `Initial State`: By default, the connection starts with database 0.
- `Post-Command State`: The connection will be associated with the specified database index.
- The number of databases is configurable in the DiceDB configuration file (`DiceDB.conf`) using the `databases` directive.
- Switching databases does not affect the data stored in other databases; it only changes the context for the current connection.
- The `SELECT` command is connection-specific. Different connections can operate on different databases simultaneously.

## Errors

1. `Invalid Database Index`:

   - Error Message: `(error) ERR DB index is out of range`
   - Occurs when the specified database index exceeds the configured maximum (default 15)

2. `Non-Integer Index`:
   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs when the provided index is not a valid integer

## Example Usage

### Switching to Database 1

```bash
127.0.0.1:7379> SELECT 1
OK
```

### Switching Back to Default Database

```bash
127.0.0.1:7379> SELECT 0
OK
```

### Invalid Database Index

### Error Example: Invalid Database Index

```bash
127.0.0.1:7379> SELECT 16
(error) ERR DB index is out of range
```

### Invalid Input Type

```bash
127.0.0.1:7379> SELECT one
(error) ERR value is not an integer or out of range
```

## Notes

As mentioned at the beginning of this document, the current version of DiceDB does not support multiple databases. The `SELECT` command is implemented as a placeholder for future functionality. All operations, regardless of the SELECT command, will continue to operate on a single database space.