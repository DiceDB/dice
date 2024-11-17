---
title: ZADD
description: The ZADD command adds one or more members with scores to a sorted set in DiceDB. If the key doesn't exist, it creates a new sorted set. If a member already exists, its score is updated. This command is essential for managing sorted data efficiently.
---

The ZADD command in DiceDB is used to add one or more members with their associated scores to a sorted set. If the specified key doesn't exist, it creates a new sorted set. For existing members, their scores are updated. This command is crucial for maintaining ordered data structures efficiently.

## Syntax

```bash
ZADD key [NX|XX] [GT|LT] [CH] [INCR] score member [score member ...]
```

## Parameters

| Parameter | Description                                                                                                       | Type   | Required |
| --------- | ----------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the sorted set                                                                                         | String | Yes      |
| `score`   | The score associated with the member                                                                              | Float  | Yes      |
| `member`  | The member to be added to the sorted set                                                                          | String | Yes      |
| `NX`      | Only add new elements. Don't update existing elements.                                                            | Flag   | No       |
| `XX`      | Only update existing elements. Don't add new elements.                                                            | Flag   | No       |
| `GT`      | Only update existing elements if the new score is greater than the current score                                  | Flag   | No       |
| `LT`      | Only update existing elements if the new score is less than the current score                                     | Flag   | No       |
| `CH`      | Modify the return value from the number of new elements added, to the total number of elements changed            | Flag   | No       |
| `INCR`    | When this option is specified, ZADD acts like ZINCRBY. Only one score-element pair can be specified in this mode. | Flag   | No       |

## Return values

| Condition                                  | Return Value                                                                                                             |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| Command executed successfully              | The number of elements added to the sorted set (not including elements already existing for which the score was updated) |
| Key holds a value that is not a sorted set | Error message                                                                                                            |

## Behaviour

- If the key does not exist, a new sorted set is created and the specified members are added with their respective scores.
- If a specified member already exists in the sorted set, its score is updated to the new score provided.
- Members are always added in sorted order according to their score, from the lowest to the highest.
- If multiple score-member pairs are specified, they are processed left to right.
- The `NX` and `XX` options are mutually exclusive and cannot be used together.
- When `CH` is specified, the command returns the total number of elements changed (added and updated).
- The `INCR` option allows the command to behave like ZINCRBY, incrementing the existing score of a member (or setting it if it doesn't exist).

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use ZADD on a key that contains a non-sorted set value.

2. `Invalid score`:
   - Error Message: `(error) ERR value is not an integer or a float`
   - Occurs when the provided score is not a valid floating-point number.

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> ZADD myzset 1 "one" 2 "two" 3 "three"
(integer) 3
127.0.0.1:7379> ZADD myzset 4 "four"
(integer) 1
```

### Updating Existing Members

```bash
127.0.0.1:7379> ZADD myzset 5 "two"
(integer) 0
```

### Using NX Option

```bash
127.0.0.1:7379> ZADD myzset NX 6 "six" 7 "two"
(integer) 1
```

### Using XX Option

```bash
127.0.0.1:7379> ZADD myzset XX 8 "eight" 9 "two"
(integer) 0
```

### Using CH Option

```bash
127.0.0.1:7379> ZADD myzset CH 10 "ten" 11 "two"
(integer) 2
```

### using INCR Option

```bash
127.0.0.1:7379> ZADD myzset INCR 1 "two"
(integer) 12
```

## Invalid Usage

```bash
127.0.0.1:7379> ZADD myzset NX XX 12 "twelve"
(error) ERR XX and NX options at the same time are not compatible
```

```bash
127.0.0.1:7379> ZADD myzset LT GT  15 "twelve"
(error) ERR GT, LT, and/or NX options at the same time are not compatible
```

## Best Practices

- Use appropriate score values to maintain the desired order of elements in the sorted set.
- Consider using the `NX` or `XX` options when you want to specifically add new elements or update existing ones, respectively.
- Use the `CH` option when you need to know the total number of elements changed, including both additions and updates.

## Notes

- The time complexity of ZADD is O(log(N)) for each item added, where N is the number of elements in the sorted set.
- Scores can be any double-precision floating-point number.
