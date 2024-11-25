---
title: ZRANK
description: The `ZRANK` command in DiceDB is used to determine the rank of a member in a sorted set. It returns the position of a member in the sorted set, with the lowest score having rank 0.
---

## Syntax

```bash
ZRANK key member [WITHSCORE]
```

## Parameters

| Parameter   | Description                                                        | Type   | Required |
| ----------- | ------------------------------------------------------------------ | ------ | -------- |
| `key`       | The key of the sorted set.                                         | String | Yes      |
| `member`    | The member whose rank is to be determined.                         | String | Yes      |
| `WITHSCORE` | If provided, the command will also return the score of the member. | String | No       |

## Return values

| Condition                          | Return Value                         |
| ---------------------------------- | ------------------------------------ |
| If member exists in the sorted set | Integer (rank of the member)         |
| If `WITHSCORE` option is used      | Array (rank and score of the member) |
| If member or key does not exist    | `nil`                                |

## Behaviour

- The `ZRANK` command searches for the specified member within the sorted set associated with the given key.
- If the key exists and is a sorted set, the command returns the rank of the member based on its score, with the lowest score having rank 0.
- If the `WITHSCORE` option is provided, the command returns both the rank and the score of the member as an array.
- If the key does not exist or the member is not found in the sorted set, the command returns `nil`.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when the specified key exists but is not associated with a sorted set.

2. `Invalid syntax or number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for 'zrank' command`
   - Occurs if the command is issued with an incorrect number of arguments.

3. `Invalid option`:
   - Error Message: `(error) ERR syntax error`
   - Occurs if an invalid option is provided.

## Example Usage

### Basic Usage

Retrieve the rank of `member1` in the sorted set `myzset`:

```bash
127.0.0.1:7379> ZADD myzset 1 member1 2 member2 3 member3
(integer) 3
127.0.0.1:7379> ZRANK myzset member1
(integer) 0
127.0.0.1:7379> ZRANK myzset member3
(integer) 2
```

### Using WITHSCORE Option

Retrieve both the rank and the score of `member2` in the sorted set `myzset`:

```bash
127.0.0.1:7379> ZADD myzset 1 member1 2 member2
(integer) 2
127.0.0.1:7379> ZRANK myzset member2 WITHSCORE
(integer) [1, 2]
```

## Best Practices

- Use `ZRANK` in combination with [`ZADD`](/commands/zadd) and `ZSCORE` for efficient management of sorted sets and leaderboards.

## Notes

- This command is particularly useful for implementing leaderboards, pagination in ranked lists, and analytics on data distribution.