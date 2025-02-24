---
title: ZPOPMIN
description: The `ZPOPMIN` command in DiceDB is used to remove and return the members with the lowest scores from the sorted set data structure at the specified key. If a count is provided, it returns up to that number of members with the lowest scores, removing them from the set.
---

The `ZPOPMIN` command in DiceDB is used to remove and return the members with the lowest scores from the sorted set data structure at the specified key. If a count is provided, it returns up to that number of members with the lowest scores, removing them from the set.

## Syntax

```bash
ZPOPMIN key [count]
```

## Parameters

| Parameter | Description                                                                          | Type    | Required |
| --------- | ------------------------------------------------------------------------------------ | ------- | -------- |
| `key`     | The name of the sorted set. If it does not exist, an empty array is returned.        | String  | Yes      |
| `count`   | The count argument specifies the number of members to return with the lowest scores. | Integer | No       |

## Return values

| Condition                                               | Return Value                           |
| ------------------------------------------------------- | -------------------------------------- |
| If the key is of valid type and records are present     | List of members including their scores |
| If the key does not exist or if the sorted set is empty | `(empty list or set)`                  |

## Behaviour

- The command first checks if the specified key exists.
- If the key does not exist, an empty array is returned.
- If the key exists but is not a sorted set, an error is returned.
- If the `count` argument is specified, up to that number of members with the lowest scores are returned and removed.
- The returned array contains the members and their corresponding scores in the order of lowest to highest.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when trying to use the command on a key that is not a sorted set.

2. `Invalid syntax or conflicting options`:

   - Error Message: `(error) ERROR wrong number of arguments for 'zpopmin' command`
   - Occurs when the command syntax is incorrect or missing required parameters.

3. `Non-integer value for `count\`\`:
   - Error Message : `(error) ERR value is not an integer or out of range`
   - Occurs when the count argument passed to the command is not an integer.

## Examples

### Basic Usage

Popping the member with the lowest score from an existing sorted set:

```bash
127.0.0.1:7379> ZADD myzset 1 member1 2 member2 3 member3
(integer) 3
127.0.0.1:7379> ZPOPMIN myzset
1) 1 "member1"
```

### Non-Existing Key

Attempting to pop the member with the lowest score from a non-existent sorted set:

```bash
127.0.0.1:7379> ZPOPMIN NON_EXISTENT_KEY
(empty array)
```

### With count argumment

Popping multiple members with the lowest scores using the count argument:

```bash
127.0.0.1:7379> ZADD myzset 1 member1 2 member2 3 member3
(integer) 3
127.0.0.1:7379> ZPOPMIN myzset 2
1) 1 "member1"
2) 2 "member2"
```

### Count Argument but Multiple Members Have the Same Score

Popping members when multiple members share the same score:

```bash
127.0.0.1:7379> ZADD myzset 1 member1 1 member2 1 member3
(integer) 3
127.0.0.1:7379> ZPOPMIN myzset 2
1) 1 "member1"
2) 1 "member2"
```

### Negative Count Argument

Attempting to pop members using a negative count argument:

```bash
127.0.0.1:7379> ZADD myzset 1 member1 2 member2 3 member3
(integer) 3
127.0.0.1:7379> ZPOPMIN myzset -1
(empty array)
```

### Floating-Point Scores

Popping members with floating-point scores:

```bash
127.0.0.1:7379> ZADD myzset 1.5 member1 2.7 member2 3.8 member3
(integer) 3
127.0.0.1:7379> ZPOPMIN myzset
1) 1.5 "member1"
```

### Wrong number of arguments

Attempting to pop from a key that is not a sorted set:

```bash
127.0.0.1:7379> SET stringkey "string_value"
OK
127.0.0.1:7379> ZPOPMIN stringkey
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid Usage

Using an invalid (non-integer) count argument:

```bash
127.0.0.1:7379> ZADD myzset 1 member1
(integer) 1
127.0.0.1:7379> ZPOPMIN myzset INCORRECT_COUNT_ARGUMENT
(error) ERR value is not an integer or out of range
```

### Wrong Type of Key

Attempting to pop from a key that is not a sorted set:

```bash
127.0.0.1:7379> SET stringkey "string_value"
OK
127.0.0.1:7379> ZPOPMIN stringkey
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```
