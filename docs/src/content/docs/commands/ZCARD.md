---
title: ZCARD
description: Documentation for the DiceDB command ZCARD
---

The `ZCARD` command in DiceDB is used to obtain the cardinality (number of elements) of the sorted set stored at the specified key. This command is useful for understanding the cardinality of the sorted set and for performing operations that depend on the number of elements in the sorted set.

## Syntax

```bash
ZCARD key
```

## Parameters

| Parameter | Description                                                                     | Type   | Required |
| --------- | ------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key associated with the sorted set for whose cardinality is to be retrieved | String | Yes      |

## Return Value

| Condition               | Return Value                         |
| ----------------------- | ------------------------------------ |
| If specified key exists | cardinality of the sorted set at key |
| If key doesn't exist    | `0`                                  |

## Behaviour

- DiceDB checks if the specified key exists.
- If the key exists and is associated with a sorted set, DiceDB counts the number of elements in the sorted set and returns this count.
- If the key does not exist, DiceDB returns `0`.
- If the key exists but is not associated with a sorted set, an error is returned.

## Errors

1. `Wrong type of key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non sorted set value.

2. `Wrong number of arguments`:

   - Error Message: `(error) -ERR wrong number of arguments for 'ZCARD' command`
   - Occurs if key isn't specified in the command.

## Example Usage

### Basic Usage

Creating sorted set `myzset` with two fields `one`, `two` with scores 1, 2 respectively. Getting cardinality of `myzset`. Adding new element into `myzset` and getting updated cardinality.

```bash
127.0.0.1:7379> ZADD myzset 1 "one" 2 "two"
(integer) 2

127.0.0.1:7379> ZCARD myzset
(integer) 2

127.0.0.1:7379> ZADD myzset 3 "three"
(integer) 1

127.0.0.1:7379> ZCARD myzset
(integer) 3
```

### Invalid Usage on non-existent sorted set

Getting cardinality of a non-existent sorted set `nonExistentZSet`.

```bash
127.0.0.1:7379> ZCARD nonExistentZSet
(integer) 0
```

### Invalid Usage on a non sorted set key

Getting cardinality of a key `mystring` associated with a non sorted set type.

```bash
127.0.0.1:7379> SET mystring "This is a string"
OK

127.0.0.1:7379> ZCARD mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `ZCARD` command is a constant-time operation, meaning its execution time is O(1) regardless of the number of elements in the sorted set.
- This command is useful for quickly determining the size of the sorted set without needing to retrieve all the fields and values.
