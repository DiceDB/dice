---
title: GETBIT
description: The `GETBIT` command is used to retrieve the bit value at a specified offset in the string value stored at a given key. This command is particularly useful for bitwise operations and managing binary data within DiceDB.
---

The `GETBIT` command is used to retrieve the bit value at a specified offset in the string value stored at a given key. This command is particularly useful for bitwise operations and managing binary data within DiceDB.

## Syntax

```bash
GETBIT key offset
```

## Parameters

| Parameter | Description                                                                                                      | Type    | Required |
| --------- | ---------------------------------------------------------------------------------------------------------------- | ------- | -------- |
| `key`     | The key of the string from which the bit value is to be retrieved. This key must reference a string value.       | String  | Yes      |
| `offset`  | The position of the bit to retrieve. The offset is a zero-based integer, meaning the first bit is at position 0. | Integer | Yes      |

## Return Values

| Condition                                   | Return Value |
| ------------------------------------------- | ------------ |
| Command is successful                       | `0` or `1`   |
| Syntax or specified constraints are invalid | error        |

## Behaviour

- Check if the specified key exists.
- If the key does not exist, it is treated as if it contains a string of zero bytes, and the bit at any offset will be `0`.
- If the key exists but does not hold a string value, an error is returned.
- If the key exists and holds a string value, the bit at the specified offset is retrieved and returned.
- If the key exists and holds a string value, however the offset is more than string length, `0` will be returned.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) wrong number of arguments for 'GETBIT' command`
   - Occurs if both key and offset are not provided.

2. `Non-string value stored against the key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but does not contain a string value.

3. `Non-integer or negative value for offset`:

   - Error Message: `(error) ERR bit offset is not an integer or out of range`
   - Occurs if the offset is not a valid integer or is negative.

## Example Usage

### Basic Usage

Setting a key `foo` with the value `a` (ASCII value of `a` is `97`, represented as `01100001` in binary). Then retrieve the bit value at index `1`.

```bash
127.0.0.1:7379> SET foo "a"
OK
127.0.0.1:7379> GETBIT foo 1
1
```

### Key does not exist

Trying to retrieve bit value from a non-existent key.

```bash
127.0.0.1:7379> GETBIT bar 5
0
```

### Key holds a non-string value

Setting a key `baz` with a list of items and retrieving bit value from the key.

```bash
127.0.0.1:7379> LPUSH baz "item"
127.0.0.1:7379> GETBIT baz 0
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid offset

Setting a key `foo` with the value `a` and trying to retrieve non-integer and negative bit value.

```bash
127.0.0.1:7379> SET foo "a"
127.0.0.1:7379> GETBIT foo -1
(error) ERR bit offset is not an integer or out of range
127.0.0.1:7379> GETBIT foo "abc"
(error) ERR bit offset is not an integer or out of range
```

### Insufficient parameters

Trying to execute `GETBIT` command without `offset` argument.

```bash
127.0.0.1:7379> GETBIT foo
(error) wrong number of arguments for 'GETBIT' command
```

## Notes

- The `GETBIT` command operates on the raw binary representation of the string. This means that the offset is counted in bits, not bytes.
- The maximum offset that can be specified is `2^32 - 1 (4294967295)`, as DiceDB strings are limited to `512 MB`.
