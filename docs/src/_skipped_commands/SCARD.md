---
title: SCARD
description: Documentation for the DiceDB command SCARD
---

The `SCARD` command in DiceDB is used to get the number of members in a set. This command is useful for determining the size of a set stored at a given key.

## Syntax

```bash
SCARD key
```

## Parameters

| Parameter | Description                                                                    | Type   | Required |
| --------- | ------------------------------------------------------------------------------ | ------ | -------- |
| `key`     | The key of the set whose cardinality (number of members) you want to retrieve. | String | Yes      |

## Return Values

| Condition                               | Return Value                  |
| --------------------------------------- | ----------------------------- |
| Key of Set type exists                  | Number of elements in the set |
| Key doesn't exist                       | `0`                           |
| Invalid syntax/key is of the wrong type | error                         |

## Behaviour

When the `SCARD` command is executed, DiceDB will:

1. Check if the key exists.
2. If the key exists and is a set, it will return the number of elements in the set.
3. If the key does not exist, it will return 0.
4. If the key exists but is not a set, an error will be returned.

## Errors

1. `Wrong type of key`:

   - Error Message: `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not a set. DiceDB expects the key to be associated with a set data type. If the key is associated with a different data type (e.g., a string, list, hash, or sorted set), this error will be raised.

2. `Wrong number of arguments`:

   - Error Message: `(error) ERROR wrong number of arguments for 'scard' command`
   - Occurs if wrong number of keys is passed to the command, such as passing more than 1 key, or passing no key.

## Examples

### Basic Example

Add three members to the set `myset` and then get the cardinality of the set using `SCARD` command.

```bash
127.0.0.1:7379> SADD myset "apple"
(integer) 1
127.0.0.1:7379> SADD myset "banana"
(integer) 1
127.0.0.1:7379> SADD myset "cherry"
(integer) 1
127.0.0.1:7379> SCARD myset
(integer) 3
```

### Non-Existent Key

Get the cardinality of a set that does not exist.

```bash
127.0.0.1:7379> SCARD nonexistingset
(integer) 0
```

### Error Example: Wrong Type

Get the cardinality of a key holding string value

```bash
127.0.0.1:7379> SET mystring "hello"
OK
127.0.0.1:7379> SCARD mystring
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

### Error Example: Wrong Argument

Get cardinality of a set holding 0 or more than 1 argument

```bash
127.0.0.1:7379> SCARD
(error) ERROR wrong number of arguments for 'scard' command
127.0.0.1:7379> SCARD myset1 myset2
(error) ERROR wrong number of arguments for 'scard' command
```
