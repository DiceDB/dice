---
title: GETDEL
description: The `GETDEL` command in DiceDB is used to retrieve the value of a specified key and then delete the key from the database. This command is useful when you need to fetch a value and ensure that it is removed from the database in a single atomic operation.
---

The `GETDEL` command in DiceDB is used to retrieve the value of a specified key and then delete the key from the database. This command is useful when you need to fetch a value and ensure that it is removed from the database in a single atomic operation.

## Syntax

```
GETDEL key
```

## Parameters

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The key whose value you want to retrieve and delete.                      | String  | Yes      |

## Return values

| Condition            | Return Value                                                     |
|----------------------|------------------------------------------------------------------|
| Key exists           | `String`: The command returns the value associated with the key. |
| Key does not exist   | `nil`: The command returns `nil`.                                |

## Behaviour

When the `GETDEL` command is executed, the following steps occur:
  1. The command checks if the specified key exists in the DiceDB database.
  2. If the key exists, the value associated with the key is retrieved.
  3. The key is then deleted from the database.
  4. The retrieved value is returned to the client.
  5. If the key does not exist, `nil` is returned, and no deletion occurs.

## Errors

The `GETDEL` command can raise errors in the following scenarios:

1. `Wrong Type Error`:

   - Error Message: `ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not a string (e.g., it is a list, set, hash, etc.).

2. `Syntax Error`:

   - Error Message: `ERROR wrong number of arguments for 'getdel' command`
   - Occurs if the command is called without the required parameter.

## Examples

### Example with Existent key

```bash
127.0.0.1:7379> SET mykey "Hello, World!"
OK
127.0.0.1:7379> GETDEL mykey
"Hello, World!"
127.0.0.1:7379> GET mykey
(nil)
```

`Explanation:` 

- The key `mykey` is set with the value `"Hello, World!"`.
- The `GETDEL` command retrieves the value `"Hello, World!"` and deletes the key `mykey` from the database.
- The `GET` command attempts to retrieve the value associated with the key `mykey` and returns `nil` as the key no longer exists.

### Example with a Non-Existent Key

```bash
127.0.0.1:7379> GETDEL nonexistingkey
(nil)
```

`Explanation:` 

- The key `nonexistingkey` does not exist in the database.
- The `GETDEL` command returns `nil` since the key is not found.

### Example with a Wrong Type of Key

```bash
127.0.0.1:7379> LPUSH mylist "item1"
(integer) 1
127.0.0.1:7379> GETDEL mylist
ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

`Explanation:` 

- The key `mylist` is a list, not a string.
- The `GETDEL` command raises a `WRONGTYPE` error because it expects the key to be a string.