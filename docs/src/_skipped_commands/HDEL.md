---
title: HDEL
description: The HDEL command in DiceDB deletes a specified field within a hash stored at a given key. If either the key or field does not exist, no action is taken, and 0 is returned.
---

# HDEL

The HDEL command in DiceDB deletes a specified field within a hash stored at a given key. If either the key or field does not exist, no action is taken, and 0 is returned.

## Syntax

```bash
HDEL key field [field ...]
```

## Parameters

| Parameter | Description                                                    | Type   | Required |
| --------- | -------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the hash from which the field(s) are to be deleted. | String | Yes      |
| `field`   | One or more fields within the hash to be deleted.              | String | Yes      |

## Return Values

| Condition                     | Return Value                                                                |
| ----------------------------- | --------------------------------------------------------------------------- |
| Field(s) deleted successfully | `Integer` (Number of fields deleted)                                        |
| Field does not exist          | `0`                                                                         |
| Key does not exist            | `0`                                                                         |
| Wrong data type               | `(error) WRONGTYPE Operation against a key holding the wrong kind of value` |
| Incorrect Argument Count      | `(error) ERR wrong number of arguments for 'hdel' command`                  |

## Behaviour

When the `HDEL` command is executed, DiceDB performs the following steps:

- It checks if the key exists in the database.
- If the key exists and is of type hash, it then checks for the specified field(s) within the hash.
- If the specified field(s) exist, they are deleted, and DiceDB returns the count of deleted fields.
- If the key does not exist or the field(s) are not present, it returns `0`.

## Errors

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure.

2. `Incorrect Argument Count`:

   - Error Message: `(error) ERR wrong number of arguments for 'hdel' command`
   - Occurs if the command is not provided with the correct number of arguments (i.e., fewer than two).

## Example Usage

### Deleting a field from a hash

```bash
127.0.0.1:7379> HSET user:1000 name "John Doe"
(integer) 1
127.0.0.1:7379> HDEL user:1000 name
(integer) 1
127.0.0.1:7379> HGET user:1000 name
(nil)
```

### Deleting multiple fields from a hash

```bash
127.0.0.1:7379> HSET user:1000 name "John Doe"
(integer) 1
127.0.0.1:7379> HSET user:1000 age "30"
(integer) 1
127.0.0.1:7379> HDEL user:1000 name age
(integer) 2
```

### Field does not exist

```bash
127.0.0.1:7379> HDEL user:1000 email
(integer) 0
```

### Key does not exist

```bash
127.0.0.1:7379> HDEL user:2000 name
(integer) 0
```

### Key is not a hash

```bash
127.0.0.1:7379> SET user:3000 "Not a hash"
OK
127.0.0.1:7379> HDEL user:3000 name
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Attempting to delete a field in a key that is not a hash

```bash
127.0.0.1:7379> SET user:5000 "This is a string"
OK
127.0.0.1:7379> HDEL user:5000 name
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

- **Behavior**: The `SET` command sets the key `user:5000` to a string value.
- **Error**: The `HDEL` command raises a `WRONGTYPE` error because user:5000 is not a hash, and `HDEL` only operates on hash data structures.

### Wrong Number of Arguments

```bash
127.0.0.1:7379> HDEL
(error) ERR wrong number of arguments for 'hdel' command

127.0.0.1:7379> HDEL user:5000
(error) ERR wrong number of arguments for 'hdel' command
```

- **Behavior**: The `HDEL` command requires at least two arguments: the key and the field name.
- **Error**: The command fails because `HDEL` requires at least a `key` and one `field` as arguments. If these are not provided, DiceDB raises an error indicating an incorrect number of arguments.

## Notes

- The `HDEL` command is essential for managing hash data in DiceDB, allowing fields to be efficiently removed when no longer needed.
