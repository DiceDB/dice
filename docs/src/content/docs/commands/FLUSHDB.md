---
title: FLUSHDB
description: The FLUSHDB command removes all keys from the currently selected database, providing a way to clear all data in a specific database space.
---

The `FLUSHDB` command is used to remove all keys from the currently selected database in a DiceDB instance. This command is useful when you need to clear all the data in a specific database without affecting other databases in the same DiceDB instance.

## Syntax

```bash
FLUSHDB
```

## Parameters

The `FLUSHDB` command does not take any parameters.

## Return Values

| Condition                | Return Value |
|--------------------------|--------------|
| Command is successful    | `OK`         |
| Authentication required  | Error: `NOAUTH Authentication required` |
| Permission denied       | Error: `NOPERM this user has no permissions to run the 'flushdb' command` |
| Read-only mode          | Error: `READONLY You can't write against a read-only replica` |

## Behaviour

When the `FLUSHDB` command is executed, the following actions occur:

1. `Immediate Deletion`: All keys in the currently selected database are immediately removed.
2. `Database Isolation`: Only the keys in the currently selected database are affected. Other databases in the same DiceDB instance remain unchanged.
3. `Persistence`: If DiceDB persistence is enabled (e.g., RDB snapshots or AOF), the deletion of keys will be reflected in the next persistence operation.

## Error Handling

The `FLUSHDB` command is straightforward and does not typically raise errors under normal circumstances. However, there are a few scenarios where issues might arise:

1. `Authentication Issues`:
   - Error Message: `(error) NOAUTH Authentication required`
   - Occurs when authentication is required but not provided

2. `Permission Issues`:
   - Error Message: `(error) NOPERM this user has no permissions to run the 'flushdb' command`
   - Occurs when the user lacks necessary permissions to execute the command

3. `Read-Only Mode`:
   - Error Message: `(error) READONLY You can't write against a read-only replica`
   - Occurs when attempting to execute on a read-only instance

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> FLUSHDB
OK
```

### Verifying Empty Database

```bash
127.0.0.1:7379> FLUSHDB
OK
127.0.0.1:7379> DBSIZE
(integer) 0
```

### Example with SELECT Command

While the following example shows the traditional syntax for working with multiple databases, please note that in the current version, all operations occur on a single database space:

```bash
127.0.0.1:7379> SELECT 1
OK
127.0.0.1:7379> FLUSHDB
OK
127.0.0.1:7379> DBSIZE
(integer) 0
```

## Best Practices

- Always verify that you're operating on the intended database before executing FLUSHDB
- Consider using backup mechanisms before executing FLUSHDB in production environments
- Use appropriate access controls to restrict FLUSHDB usage to authorized users only

## Notes

- The current version of DiceDB operates on a single database space. While the `SELECT` command is available as a placeholder, switching databases will not affect the operation of the `FLUSHDB` command, and it will always clear the keys from the single available database space.

- The command is particularly powerful and should be used with caution as it results in immediate, irreversible data loss for all keys in the database.