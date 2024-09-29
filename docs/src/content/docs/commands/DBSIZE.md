---
title: DBSIZE
description: Documentation for the DiceDB command DBSIZE
---

The `DBSIZE` command in DiceDB is used to return the number of keys in the currently selected database. This command is useful for monitoring and managing the size of your DiceDB database, providing a quick way to understand the number of keys stored.

## Parameters

The `DBSIZE` command does not take any parameters.

## Return Value

The `DBSIZE` command returns an integer that represents the number of keys in the currently selected database.

- `Type:` Integer
- `Description:` The number of keys in the currently selected database.

## Behaviour

When the `DBSIZE` command is executed, DiceDB will count the number of keys in the currently selected database and return this count as an integer. This operation is generally very fast, as DiceDB is optimized for such operations. The command does not modify the database in any way; it is purely informational.

## Error Handling

The `DBSIZE` command is straightforward and does not typically result in errors under normal usage. However, there are a few scenarios where errors might be encountered:

1. `Connection Issues:` If there is a problem with the connection to the DiceDB server, an error will be raised.

   - `Error Message:` `ERR Connection lost`

2. `Authentication Issues:` If the DiceDB server requires authentication and the client has not authenticated, an error will be raised.

   - `Error Message:` `NOAUTH Authentication required`

3. `Permission Issues:` If the client does not have the necessary permissions to execute the command, an error will be raised.

   - `Error Message:` `NOPERM this user has no permissions to run the 'dbsize' command`

## Example Usage

### Basic Usage

To get the number of keys in the currently selected database, simply execute the `DBSIZE` command:

```shell
127.0.0.1:6379> DBSIZE
(integer) 42
```

In this example, the currently selected database contains 42 keys.

### Using with Multiple Databases

If you are working with multiple databases, you can switch between them using the `SELECT` command and then use `DBSIZE` to get the number of keys in each database:

```shell
127.0.0.1:6379> SELECT 0
OK
127.0.0.1:6379> DBSIZE
(integer) 42

127.0.0.1:6379> SELECT 1
OK
127.0.0.1:6379> DBSIZE
(integer) 15
```

In this example, database 0 contains 42 keys, and database 1 contains 15 keys.
