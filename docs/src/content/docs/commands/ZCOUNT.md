---
title: ZCOUNT
description: The ZCOUNT command in DiceDB counts the number of members in a sorted set at the specified key whose scores fall within a given range. The command takes three arguments - the key of the sorted set, the minimum score, and the maximum score. If the key does not exist or contains no members, the command returns 0. It allows for flexible range queries by using special values like -inf and +inf to specify unbounded limits.
---

The ZCOUNT command in DiceDB counts the number of members in a sorted set at the specified key whose scores fall within a given range. The command takes three arguments: the key of the sorted set, the minimum score, and the maximum score. If the key does not exist or contains no members, the command returns 0. It allows for flexible range queries by using special values like -inf and +inf to specify unbounded limits.

## Syntax

```bash
ZCOUNT key min max
```

## Parameters

| Parameter | Description                                      | Type   | Required |
| --------- | ------------------------------------------------ | ------ | -------- |
| key       | The name of the sorted set to operate on.        | String | Yes      |
| min       | Minimum score (inclusive) of the range to count. | Int    | Yes      |
| max       | Maximum score (inclusive) of the range to count. | Int    | Yes      |

## Return Values

| Condition                             | Return Value                                          |
| ------------------------------------- | ----------------------------------------------------- |
| If the key exists and is a sorted set | Returns the count of elements in the specified range. |
| If the key does not exist             | Returns `0`.                                          |
| If the key is not a sorted set        | Returns an error.                                     |

## Behaviour

Retrieves the sorted set associated with the given key.
Counts all elements whose scores fall between the given min and max (inclusive).
If the key does not exist, it behaves as if it is an empty sorted set and returns 0.
If the key is not of the sorted set type, an error is returned.

## Errors

1. **Wrong Type**:
   - **Message**: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
2. **Wrong Argument Count**:
   - **Message**: `(error) ERROR wrong number of arguments for 'zcount' command`

## Example Usage

```bash
127.0.0.1:7379> ZCOUNT NON_EXISTENT_KEY 0 100
0

127.0.0.1:7379> ZADD myzset 10 member1 20 member2 30 member3
(integer) 3
127.0.0.1:7379> ZCOUNT myzset 15 25
1

127.0.0.1:7379> ZCOUNT myzset 50 100
0

127.0.0.1:7379> ZCOUNT myzset 30 10
0
```

### Invalid usage

```bash
127.0.0.1:7379> ZCOUNT myzset
(error) ERROR wrong number of arguments for 'zcount' command

127.0.0.1:7379> ZCOUNT myzset "invalid" 100
(integer) 0
```
