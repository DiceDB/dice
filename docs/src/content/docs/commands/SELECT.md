---
title: SELECT
description: Documentation for the DiceDB command SELECT
---

**Note:** As of today, DiceDB does not support multiple databases. Therefore, the `SELECT` command is currently a dummy method and does not affect the database. It remains as a placeholder.

The `SELECT` command is used to switch the currently selected database for the current connection in DiceDB. By default, DiceDB starts with database 0, but it supports multiple databases, which can be accessed by using the `SELECT` command. This command is essential for managing data across different logical databases within a single DiceDB instance.

## Syntax

```
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

   - `Error`: `(error) ERR DB index is out of range`
   - `Condition`: If the specified database index is outside the range of available databases.

2. `Non-Integer Index`:

   - `Error`: `(error) ERR value is not an integer or out of range`
   - `Condition`: If the provided index is not a valid integer.

## Example Usage

### Switching to Database 1

```shell
127.0.0.1:7379> SELECT 1
OK
```

In this example, the connection switches to database 1. All subsequent commands will operate on database 1.

### Switching to Database 0

```shell
127.0.0.1:7379> SELECT 0
OK
```

Here, the connection switches back to the default database 0.

### Error Example: Invalid Database Index

```shell
127.0.0.1:7379> SELECT 16
(error) ERR DB index is out of range
```

In this example, an error is raised because the specified database index 16 is outside the default range of 0-15.

### Error Example: Non-Integer Index

```shell
127.0.0.1:7379> SELECT one
(error) ERR value is not an integer or out of range
```

In this example, an error is raised because the provided index is not a valid integer.
