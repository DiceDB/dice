---
title: ZCOUNT
description: The ZCOUNT command in DiceDB counts the number of members in a sorted set at the specified key whose scores fall within a given range. The command takes three arguments: the key of the sorted set, the minimum score, and the maximum score. If the key does not exist or contains no members, the command returns 0. It allows for flexible range queries by using special values like -inf and +inf to specify unbounded limits.
---

## Introduction

The ZCOUNT command in DiceDB counts the number of members in a sorted set at the specified key whose scores fall within a given range. The command takes three arguments: the key of the sorted set, the minimum score, and the maximum score. If the key does not exist or contains no members, the command returns 0. It allows for flexible range queries by using special values like -inf and +inf to specify unbounded limits.

## Syntax

```bash
ZCOUNT key min max
```

## Parameters

- **`key`**: The name of the sorted set. If the key does not exist, it returns 0.
- **`min`**: Minimum score (inclusive) for counting members. This can be a float, or special values like `-inf`.
- **`max`**: Maximum score (inclusive) for counting members. This can also be a float, or special values like `+inf`.

## Return Values

- **Count of matching members**: If the key exists and is a sorted set.
- **0**: If the key does not exist or if the sorted set is empty.
- **Count of 0**: If non-numeric values are provided for `min` or `max`, they are treated as 0.

## Errors

1. **Wrong Type**: 
   - **Message**: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
2. **Wrong Argument Count**: 
   - **Message**: `(error) ERROR wrong number of arguments for 'zcount' command`


## Examples

### Non-Existing Key

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

127.0.0.1:7379> ZCOUNT myzset
(error) ERROR wrong number of arguments for 'zcount' command

127.0.0.1:7379> ZCOUNT myzset "invalid" 100
(integer) 0
