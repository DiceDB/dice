---
title: GETRANGE
description: The `GETRANGE` command in DiceDB is used to get a substring of a string, provided the start and end indices
---

The `GETRANGE` command in DiceDB is used to get a substring of a string, provided the start and end indices.

## Syntax

```bash
GETRANGE key start end
```

## Parameters

| Parameter | Description                                   | Type    | Required |
| --------- | --------------------------------------------- | ------- | -------- |
| `key`     | The name of the key containing the string.    | String  | Yes      |
| `start`   | The starting index of the required substring. | Integer | Yes      |
| `end`     | The ending index of the required substring.   | Integer | Yes      |

## Return values

| Condition                                                      | Return Value                                   |
| -------------------------------------------------------------- | ---------------------------------------------- |
| if key contains a valid string                                 | a substring based on the start and end indices |
| if `start` is greater than `end`                               | an empty string is returned                    |
| if the `end` exceeds the length of the string present at `key` | the entire string is returned                  |
| if the `start` is greater than the length of the string        | an empty string is returned.                   |

## Behaviour

- If the specified key does not exist, the `GETRANGE` command returns an empty string.
- If `start` is greater than `end`, the `GETRANGE` command returns an empty string.
- If `start` is not within the length of the string, the `GETRANGE` command returns an empty string.
- `start` and `end` can be negative which removed `end + 1` characters from the other side of the string.
- Both `start` and `end` can be negative, which removes characters from the string, starting from the `end + 1` position on the right.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

2. `Invalid syntax or conflicting options`:

   - Error Message: `(error) ERROR wrong number of arguments for 'GETRANGE' command`
   - If the number of arguments are not exactly equal to 4.

3. `Invalid input type for start and end`

   - Error Message: `(error) ERROR value is not an integer or out of range`
   - If `start` and `end` are not integers.

## Example Usage

### Basic Usage

Assume we have a string stored in `foo`

```bash
127.0.0.1:7379> SET foo apple
OK
```

```bash
127.0.0.1:7379> GETRANGE foo 1 3
"ppl"
```

```bash
127.0.0.1:7379> GETRANGE foo 0 -1
"apple"
```

```bash
127.0.0.1:7379> GETRANGE foo 0 -10
""
```

```bash
127.0.0.1:7379> GETRANGE foo 0 -2
"appl"
```

```bash
127.0.0.1:7379> GETRANGE foo 0 1001
"apple"
```

`GETRANGE` returns string representation of byte array stored in bitmap

```bash
127.0.0.1:7379> SETBIT bitmapkey 2 1
(integer) 0
127.0.0.1:7379> SETBIT bitmapkey 3 1
(integer) 0
127.0.0.1:7379> SETBIT bitmapkey 5 1
(integer) 0
127.0.0.1:7379> SETBIT bitmapkey 10 1
(integer) 0
127.0.0.1:7379> SETBIT bitmapkey 11 1
(integer) 0
127.0.0.1:7379> SETBIT bitmapkey 14 1
(integer) 0
127.0.0.1:7379> GETRANGE bitmapkey 0 -1
"42"
```

### Invalid usage

Trying to use `GETRANGE` without giving the value

```bash
127.0.0.1:7379> GETRANGE foo
(error) ERROR wrong number of arguments for 'getrange' command
```

Trying to use `GETRANGE` on an invalid data type.

```bash
127.0.0.1:7379> LPUSH foo apple
127.0.0.1:7379> GETRANGE foo 0 5
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

```bash
127.0.0.1:7379> GETRANGE foo s e
(error) ERROR value is not an integer or out of range
```
