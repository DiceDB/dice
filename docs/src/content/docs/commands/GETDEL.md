---
title: GETDEL
description: The `GETDEL` command in DiceDB is used to retrieve the value of a specified key and then delete the key from the database. This command is useful when you need to fetch a value and ensure that it is removed from the database in a single atomic operation.
---

The `GETDEL` command in DiceDB is used to retrieve the value of a specified key and then delete the key from the database. This command is useful when you need to fetch a value and ensure that it is removed from the database in a single atomic operation.

## Syntax

```bash
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

- If the specified key exists, `GETDEL` retrieves its value and then deletes the key from the database.
- The retrieved value is returned to the client.
- If the key does not exist, `GETDEL` returns `nil` and no deletion is performed.

## Errors

1. `Wrong Type Error`:

   - Error Message: `ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not a string (e.g., it is a list, set, hash, etc.).

2. `Syntax Error`:

   - Error Message: `ERROR wrong number of arguments for 'getdel' command`
   - Occurs if the command is called without the required parameter.

## Example Usage

### Retreive and Delete an Existing Key

Setting a key `mykey` with the value `"Hello, World!"` and then using `GETDEL` to retrieve and delete it.

```bash
127.0.0.1:7379> SET mykey "Hello, World!"
OK
127.0.0.1:7379> GETDEL mykey
"Hello, World!"
127.0.0.1:7379> GET mykey
(nil)
```

### Using `GETDEL` on a Non-Existent Key

Trying to retrieve and delete a key `nonexistingkey` that does not exist.

```bash
127.0.0.1:7379> GETDEL nonexistingkey
(nil)
```

### Using `GETDEL` on a Key with a Different Data Type

Setting a key `mylist` as a list and then trying to use `GETDEL`, which is incompatible with non-string data types.

```bash
127.0.0.1:7379> LPUSH mylist "item1"
(integer) 1
127.0.0.1:7379> GETDEL mylist
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```