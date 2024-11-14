---
title: RANDOMKEY
description: The `RANDOMKEY` command in DiceDB return a random key from the currently selected database.
---

The `RANDOMKEY` command in DiceDB is used to return a random key from the currently selected database.

## Syntax

```
RANDOMKEY
```

## Parameters

The `RANDOMKEY` command does not take any parameters.

## Return values

| Condition                                     | Return Value                                        |
|-----------------------------------------------|-----------------------------------------------------|
| Command is successful                         | A random key from the keyspace of selected database |
| Failure to scan keyspace or pick a random key | Error                                               |

## Behaviour

- When executed, `RANDOMKEY` fetches the keyspace from currently selected database and picks a random key from it.
- The operation is slow and may return an expired key if it hasn't been evicted.
- The command does not modify the database in any way; it is purely informational.

## Errors
The `RANDOMKEY` command is straightforward and does not typically result in errors under normal usage. However, since it internally depends on KEYS command, it can fail for the same cases as KEYS.

## Example Usage

### Basic Usage

Getting a random key from the currently selected database:

```shell
127.0.0.1:7379> RANDOMKEY
"key_6"
```

### Using with Multiple Databases

If you are working with multiple databases, you can switch between them using the `SELECT` command and then use `RANDOMKEY` to get a random key from selected database:

```shell
127.0.0.1:7379> SELECT 0
OK
127.0.0.1:7379> RANDOMKEY
"db0_key_54"

127.0.0.1:7379> SELECT 1
OK
127.0.0.1:7379> RANDOMKEY
"db1_key_435"
```
