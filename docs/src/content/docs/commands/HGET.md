---
title: HGET
description: The `HGET` command in DiceDB is used to retrieve the value associated with a specified field within a hash stored at a given key. If the key or the field does not exist, the command returns a `nil` value.
---

The `HGET` command in DiceDB is used to retrieve the value associated with a specified field within a hash stored at a given key. If the key or the field does not exist, the command returns a `nil` value.

## Syntax

```bash
HGET key field
```

## Parameters

| Parameter | Description                                                          | Type   | Required |
| --------- | -------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the hash from which the field's value is to be retrieved. | String | Yes      |
| `field`   | The field within the hash whose value is to be retrieved.            | String | Yes      |

## Return Values

| Condition                               | Return Value                                                                |
| --------------------------------------- | --------------------------------------------------------------------------- |
| Field exists in the hash                | `String` (value of the field)                                               |
| Key does not exist or field not present | `nil`                                                                       |
| Wrong data type                         | `(error) WRONGTYPE Operation against a key holding the wrong kind of value` |
| Incorrect Argument Count                | `(error) ERR wrong number of arguments for 'hget' command`                  |

## Behaviour

When the `HGET` command is executed, DiceDB performs the following steps:

- It checks if the key exists in the database.
- If the key exists and is of type hash, it then checks if the specified field exists within the hash.
- If the field exists, it retrieves and returns the value associated with the field.
- If the key does not exist or the field is not present in the hash, it returns `nil`.

## Errors

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if `key` holds a non-hash data structure.

2. `Incorrect Argument Count`:

   - Error Message: `(error) ERR wrong number of arguments for 'hget' command`
   - Occurs if the command is not provided with the correct number of arguments (i.e., fewer than two).

## Example Usage

### Retrieving a value from a hash

```bash
127.0.0.1:7379> HSET user:1000 name "John Doe"
(integer) 1
127.0.0.1:7379> HSET user:1000 age "30"
(integer) 1
127.0.0.1:7379> HGET user:1000 name
"John Doe"
```

### Field does not exist

```bash
127.0.0.1:7379> HGET user:1000 email
(nil)
```

### Key does not exist

```bash
127.0.0.1:7379> HGET user:2000 name
(nil)
```

### Key is not a hash

```bash
127.0.0.1:7379> SET user:3000 "Not a hash"
OK
127.0.0.1:7379> HGET user:3000 name
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid Usage

```bash
127.0.0.1:7379> SET product:2000 "This is a string"
OK
127.0.0.1:7379> HGET product:2000 name
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

```bash
127.0.0.1:7379> HGET product:2000
(error) ERR wrong number of arguments for 'hget' command

127.0.0.1:7379> HGET product:2000 name name2
(error) ERR wrong number of arguments for 'hget' command
```

## Notes

- The `HGET` command is a read-only command and does not modify the hash or any other data in the DiceDB database.

By understanding the `HGET` command, you can efficiently retrieve values from hashes stored in your DiceDB database, ensuring that your application can access the necessary data quickly and reliably.
