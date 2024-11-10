---
title: ZREM
description: Documentation for the DiceDB command ZREM
---

The `ZREM` command in DiceDB is used to remove the specified members from the sorted set stored at key and return the number of members removed from the sorted set. This command ignores the non-existing members.

## Syntax

```bash
ZREM key member [member ...]
```

## Parameters

| Parameter | Description                                                            | Type   | Required | Multiple |
| --------- | ---------------------------------------------------------------------- | ------ | -------- | -------- |
| `key`     | The key associated with the sorted set whose members are to be removed | String | Yes      | No       |
| `member`  | One or more members to remove from the sorted set                      | String | Yes      | Yes      |

## Return Value

| Condition                           | Return Value                                        |
| ----------------------------------- | --------------------------------------------------- |
| If specified key and members exists | Count of members removed from the sorted set at key |
| If key doesn't exist                | `0`                                                 |
| If member doesn't exist             | `0`                                                 |

## Behaviour

- DiceDB checks if the specified key exists.
- If the key exists and is associated with a sorted set, DiceDB removes the members from the sorted set and returns number of members removed.
- If the key does not exist, DiceDB returns `0`.
- If the key exists but is not associated with a sorted set, an error is returned.

## Errors

1. `Wrong type of key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non sorted set value.

2. `Wrong number of arguments`:

   - Error Message: `(error) -ERR wrong number of arguments for 'ZREM' command`
   - Occurs if key or member isn't specified in the command.

## Example Usage

### Basic Usage

Creating sorted set `myzset` with fields `one`, `two`, `three`, `four`, `five` with scores 1, 2, 3, 4, 5 respectively. Removing elements from `myzset`.

```bash
127.0.0.1:7379> ZADD myzset 1 "one" 2 "two" 3 "three" 4 "four" 5 "five"
(integer) 5
127.0.0.1:7379> ZREM myzset one
(integer) 1
127.0.0.1:7379> ZREM myzset two six
(integer) 1
127.0.0.1:7379> ZREM myzset three four
(integer) 2
```

### Invalid Usage on non-existent sorted set

Removing element from a non-existent sorted set `nonExistentZSet`.

```bash
127.0.0.1:7379> ZREM nonExistentZSet one
(integer) 0
```

### Invalid Usage on a non sorted set key

Getting cardinality of a key `mystring` associated with a non sorted set type.

```bash
127.0.0.1:7379> SET mystring "This is a string"
OK
127.0.0.1:7379> ZREM mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `ZREM` command is a O(M\*log(N)) time-complexity operation, with N being the number of elements in the sorted set and M the number of elements to be removed.
