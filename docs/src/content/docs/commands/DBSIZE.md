---
title: DBSIZE
description: The `DBSIZE` command in DiceDB returns the number of keys in the currently selected database, providing a quick way to understand the size of your database.
---

> **Important Note:** As of the current version, DiceDB does not support multiple databases. Therefore, while the documentation mentions database selection via the `SELECT` command, all operations occur on a single database space.

The `DBSIZE` command in DiceDB is used to return the number of keys in the currently selected database. This command is useful for monitoring and managing the size of your DiceDB database, providing a quick way to understand the number of keys stored.

## Syntax

```bash
DBSIZE
```

## Parameters

The `DBSIZE` command does not take any parameters.

## Return values

| Condition             | Return Value                            |
| --------------------- | --------------------------------------- |
| Command is successful | Integer representing the number of keys |
| Connection issues     | Error: `ERR Connection lost`            |
| Authentication issues | Error: `NOAUTH Authentication required` |

## Behaviour

- When executed, `DBSIZE` counts the number of keys in the currently selected database and returns this count as an integer.
- This operation is generally very fast, as DiceDB is optimized for such operations.
- The command does not modify the database in any way; it is purely informational.
- If multiple databases are in use, `DBSIZE` will only count keys in the currently selected database.

## Errors

The `DBSIZE` command is straightforward and does not typically result in errors under normal usage. However, there are a few scenarios where errors might be encountered:

1. `Connection Issues`:

   - Error Message: `ERR Connection lost`
   - Occurs when there is a problem with the connection to the DiceDB server.

2. `Authentication Issues`:
   - Error Message: `NOAUTH Authentication required`
   - Occurs if the DiceDB server requires authentication and the client has not authenticated.

## Example Usage

### Basic Usage

Getting the number of keys in the currently selected database:

```bash
127.0.0.1:7379> DBSIZE
(integer) 42
```

In this example, the currently selected database contains 42 keys.

### Using with Multiple Databases

While the following example shows the traditional syntax for working with multiple databases, please note that in the current version of DiceDB, all operations occur on a single database space:

```bash
127.0.0.1:7379> SELECT 0
OK
127.0.0.1:7379> DBSIZE
(integer) 42

127.0.0.1:7379> SELECT 1
OK
127.0.0.1:7379> DBSIZE
(integer) 42
```

In the current implementation, both commands will return the same count as they operate on the same database space.

### Errors

Attempting to use `DBSIZE` without proper authentication:

```bash
127.0.0.1:7379> DBSIZE
(error) NOAUTH Authentication required
```

## Notes

As mentioned at the beginning of this document, the current version of DiceDB operates on a single database space. While the `SELECT` command is available as a placeholder, switching databases will not affect the operation of the `DBSIZE` command, and it will always return the count of keys from the single available database space.
