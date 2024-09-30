---
title: FLUSHDB
description: Documentation for the DiceDB command FLUSHDB
---

The `FLUSHDB` command is used to remove all keys from the currently selected database in a DiceDB instance. This command is useful when you need to clear all the data in a specific database without affecting other databases in the same DiceDB instance.

## Parameters

The `FLUSHDB` command does not take any parameters.

## Return Value

The `FLUSHDB` command returns a simple string reply:

- `OK`: Indicates that the operation was successful and all keys in the current database have been removed.

## Behaviour

When the `FLUSHDB` command is executed, the following actions occur:

1. `Immediate Deletion`: All keys in the currently selected database are immediately removed.
2. `Database Isolation`: Only the keys in the currently selected database are affected. Other databases in the same DiceDB instance remain unchanged.
3. `Persistence`: If DiceDB persistence is enabled (e.g., RDB snapshots or AOF), the deletion of keys will be reflected in the next persistence operation.

## Error Handling

The `FLUSHDB` command is straightforward and does not typically raise errors under normal circumstances. However, there are a few scenarios where issues might arise:

1. `Permission Denied`: If the DiceDB instance is configured with ACLs (Access Control Lists) and the user does not have the necessary permissions to execute the `FLUSHDB` command, an error will be raised.

   - Error: `(error) NOAUTH Authentication required.` or `(error) NOPERM this user has no permissions to run the 'flushdb' command`

2. `Read-Only Mode`: If the DiceDB instance is in read-only mode (e.g., a read-only replica), the `FLUSHDB` command will not be allowed.

   - Error: `(error) READONLY You can't write against a read-only replica.`

## Example Usage

### Basic Usage

To clear all keys from the currently selected database:

```sh
127.0.0.1:6379> FLUSHDB
OK
```

### Using with Multiple Databases

If you have multiple databases and want to clear a specific one, you need to select the database first using the `SELECT` command, then execute `FLUSHDB`:

```sh
127.0.0.1:6379> SELECT 1
OK
127.0.0.1:6379[1]> FLUSHDB
OK
```

### Checking the Result

After executing `FLUSHDB`, you can verify that the database is empty by using the `DBSIZE` command, which returns the number of keys in the currently selected database:

```sh
127.0.0.1:6379> DBSIZE
(integer) 0
```

## Notes

- `Data Loss`: The `FLUSHDB` command will result in the loss of all data in the selected database. Use this command with caution, especially in production environments.
- `Atomic Operation`: The `FLUSHDB` command is atomic, meaning that all keys are removed in a single operation without any intermediate states.
