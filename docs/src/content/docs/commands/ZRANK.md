---
title: ZRANK
description: Documentation for the DiceDB command ZRANK
---

The `ZRANK` command is used to determine the rank of a member in a sorted set stored at a specific key in DiceDB. The rank is based on the score of the member, with the member with the lowest score having rank 0.

## Syntax

```
ZRANK key member [WITHSCORE]
```

## Parameters

| Parameter   | Description                                                                 | Type   | Required |
|-------------|-----------------------------------------------------------------------------|--------|----------|
| `key`       | The key of the sorted set.                                                  | String | Yes      |
| `member`    | The member whose rank is to be determined.                                 | String | Yes      |
| `WITHSCORE` | (Optional) If provided, the command will also return the score of the member.| String | No       |

## Return Value

- **Integer**: The rank of the member in the sorted set, with the scores ordered from low to high. The rank is 0-based.
- **Array**: If the `WITHSCORE` option is used, returns an array containing the rank and the score of the member.
- **`nil`**: If the member does not exist in the sorted set or the key does not exist.

## Behaviour

When the `ZRANK` command is executed:

1. **Key Verification**:
   - DiceDB checks if the specified key exists.
   - If the key does not exist, `nil` is returned.
   
2. **Type Verification**:
   - If the key exists, DiceDB verifies that it is associated with a sorted set.
   - If the key is not a sorted set, a `WRONGTYPE` error is returned.
   
3. **Member Search**:
   - DiceDB searches for the specified member within the sorted set.
   - If the member exists, its rank is determined based on its score.
   
4. **`WITHSCORE` Option**:
   - If the `WITHSCORE` option is provided and the member exists, DiceDB returns both the rank and the score of the member.
   
5. **Non-Existent Member**:
   - If the member does not exist in the sorted set, `nil` is returned.

## Error Handling

The `ZRANK` command can raise the following errors:

- **`WRONGTYPE Operation against a key holding the wrong kind of value`**:
  - Occurs if the specified key exists but is not associated with a sorted set.
  
- **`ERROR wrong number of arguments for 'zrank' command`**:
  - Occurs if the command is issued with an incorrect number of arguments.
  
- **`ERROR syntax error`**:
  - Occurs if an invalid option is provided.

## Examples

### Basic Example

Retrieve the rank of `member1` in the sorted set `myzset`.

```bash
127.0.0.1:7379> ZADD myzset 1 member1 2 member2 3 member3
(integer) 3
127.0.0.1:7379> ZRANK myzset member1
(integer) 0
127.0.0.1:7379> ZRANK myzset member3
(integer) 2
```


### Example with `WITHSCORE` Option

Retrieve both the rank and the score of `member2` in the sorted set `myzset`.

```bash
127.0.0.1:7379> ZADD myzset 1 member1 2 member2
(integer) 2
127.0.0.1:7379> ZRANK myzset member2 WITHSCORE
(integer) [1, 2]
```


## Additional Notes

- **Zero-Based Rank**: The `ZRANK` command is zero-based; the first element has a rank of 0.
  
- **Atomic Operation**: The command is atomic, ensuring that the rank calculation is consistent even in concurrent environments.
  
- **Performance Considerations**: The `ZRANK` command has a time complexity of O(log N), where N is the number of elements in the sorted set.
  
- **Use Cases**:
  - **Leaderboards**: Determining the position of a player in a game leaderboard.
  - **Pagination**: Navigating through ranked lists in applications.
  - **Analytics**: Understanding the distribution and ranking of data points.

By following this documentation, you should be able to effectively use the `ZRANK` command in DiceDB to determine the ranking of members within sorted sets and integrate this functionality into your applications seamlessly.