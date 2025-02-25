---
title: ZRANGE
description: The `ZRANGE` command in DiceDB is used to return a range of members in a sorted set, by index. The members are ordered from the lowest to the highest score.
---

The `ZRANGE` command in DiceDB is used to return a range of members in a sorted set, by index. The members are ordered from the lowest to the highest score.

## Syntax

```bash
ZRANGE key start stop [WITHSCORES] [REV]
```

## Parameters

| Parameter    | Description                                                              | Type    | Required |
| ------------ | ------------------------------------------------------------------------ | ------- | -------- |
| `key`        | The name of the sorted set.                                              | String  | Yes      |
| `start`      | The starting index of the range.                                         | Integer | Yes      |
| `stop`       | The ending index of the range.                                           | Integer | Yes      |
| `WITHSCORES` | Optional. Returns the scores of the elements in the result.              | None    | No       |
| `REV`        | Optional. Returns the elements in reverse order, from highest to lowest. | None    | No       |

## Return values

| Condition                                | Return Value                             |
| ---------------------------------------- | ---------------------------------------- |
| If the key exists and the range is valid | Array of elements in the specified range |
| If the key does not exist                | Empty array                              |
| If the key is not a sorted set           | Error                                    |

## Behaviour

- The `ZRANGE` command returns the specified range of elements in the sorted set stored at `key`.
- The elements are considered to be ordered from the lowest to the highest score.
- Both `start` and `stop` are 0-based indexes, where 0 is the first element, 1 is the next element, and so on.
- These indexes can also be negative numbers indicating offsets from the end of the sorted set, with -1 being the last element of the sorted set, -2 the penultimate element, and so on.
- If `WITHSCORES` is specified, the command returns the elements and their scores.
- If `REV` is specified, the command returns the elements in reverse order, from highest to lowest score.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-sorted set value.

2. `Invalid syntax or conflicting options`:
   - Error Message: `(error) ERR syntax error`
   - Occurs if the command's syntax is incorrect, such as incompatible options like `WITHSCORES` and `REV` used together, or a missing required parameter.

## Example Usage

### Basic Usage

Retrieving a range of elements from a sorted set `leaderboard` from index 0 to 2

```bash
127.0.0.1:7379> ZADD leaderboard 50 "Alice" 70 "Bob" 60 "Charlie"
(integer) 3
127.0.0.1:7379> ZRANGE leaderboard 0 2
1) "Alice"
2) "Charlie"
3) "Bob"
```

### Using `WITHSCORES`

Retrieving a range of elements from a sorted set `leaderboard` from index 0 to 2 with scores

```bash
127.0.0.1:7379> ZRANGE leaderboard 0 2 WITHSCORES
1) "Alice"
2) "50"
3) "Charlie"
4) "60"
5) "Bob"
6) "70"
```

### Using `REV`

Retrieving a range of elements from a sorted set `leaderboard` from index 0 to 2 in reverse order

```bash
127.0.0.1:7379> ZRANGE leaderboard 0 2 REV
1) "Bob"
2) "Charlie"
3) "Alice"
```

### Invalid usage

Trying to use `ZRANGE` on a key that is not a sorted set

```bash
127.0.0.1:7379> SET foo bar
OK
127.0.0.1:7379> ZRANGE foo 0 2
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

Trying to use `ZRANGE` with invalid syntax

```bash
127.0.0.1:7379> ZRANGE leaderboard 0
(error) ERR syntax error
```
