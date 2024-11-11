---
title: BITCOUNT
description: Documentation for the DiceDB command BITCOUNT
---

The `BITCOUNT` command in DiceDB is used to count the number of set bits (i.e., bits with value 1) in a string. This command is particularly useful for applications that need to perform bitwise operations and analyze binary data stored in DiceDB.

## Syntax

```bash
BITCOUNT key [start end [BYTE | BIT]]
```

- By default, the `start` and `end` parameters are byte positions, but we can use an additional argument BIT to specify a bit position.
- Negative values for `start` and `end` are interpreted as offsets from the end of the string. For example, -1 means the last byte/bit, -2 means the second last byte/bit, and so on.

## Parameters

| Parameter | Description                                                        | Type    | Required | Default |
| --------- | ------------------------------------------------------------------ | ------- | -------- | ------- |
| `key`     | The key of the string for which the bit count is to be calculated. | String  | Yes      |         |
| `start`   | The starting byte/bit position (inclusive) to count the bits.      | Integer | No       | 0       |
| `end`     | The ending byte/bit position (inclusive) to count the bits.        | Integer | No       | -1(end) |
| `BYTE`    | Use the specified range as byte indices                            | String  | No       |         |
| `BIT`     | Use the specified ragne as bit indices                             | String  | No       |         |

## Return Value

The `BITCOUNT` command returns an integer representing the number of bits set to 1 in the specified range of the string.

| Condition                       | Return Value                                                                 |
| ------------------------------- | ---------------------------------------------------------------------------- |
| Command is successful           | number of bits set to 1 in the specified range of the string                 |
| Non-existent key                | `0`                                                                          |
| Out of range indices            | no. of set bits based on the valid range (if any) inside the specified range |
| Invalid syntax / wrong key type | error                                                                        |

## Behaviour

When the `BITCOUNT` command is executed, DiceDB will:

1. Retrieve the string stored at the specified key.
2. If the `start` and `end` parameters are provided, DiceDB will consider only the specified range of bytes/bits within the string.
3. Whether it's the byte index or a bit index would be decided based on the optional parameter BYTE or BIT. (The default is BYTE)
4. Count the number of bits set to 1 within the specified range or the entire string if no range is specified.
5. Return the count as an integer.

## Errors

1. `Non-existent Key`:

   - Error Message: None (treats it as an empty string and returns 0)
   - Occurs if the specified key does not exist

   ```bash
   127.0.0.1:7379> BITCOUNT non_existent_key
   (integer) 0
   ```

2. `Wrong Type of Key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but does not hold a string value

   ```bash
   127.0.0.1:7379> LPUSH mylist "element"
   (integer) 1
   127.0.0.1:7379> BITCOUNT mylist
   (error) WRONGTYPE Operation against a key holding the wrong kind of value
   ```

3. `Invalid Syntax / Range`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs if the `start` or `end` parameters are not integers

   ```bash
   127.0.0.1:7379> SET mykey "foobar"
   OK
   127.0.0.1:7379> BITCOUNT mykey start end
   (error) ERR value is not an integer or out of range
   ```

   ```bash
   127.0.0.1:7379> SET mykey "foobar"
   OK
   127.0.0.1:7379> BITCOUNT mykey BIT
   (error) ERR value is not an integer or out of range
   ```

4. `Out of Range Indices`:

   - Error Message: None (will be handled gracefully by considering the valid range within the string)
   - Occurs if the `start` or `end` parameters are out of range of the string length

   ```bash
   127.0.0.1:7379> SET mykey "foobar"
   OK
   127.0.0.1:7379> BITCOUNT mykey 5 6
   (integer) 4
   127.0.0.1:7379> BITCOUNT mykey 5 10
   (integer) 4
   ```

## Example Usage

### Example 1: Counting bits in the entire string

```bash
127.0.0.1:7379> SET mykey "foobar"
OK
127.0.0.1:7379> BITCOUNT mykey
(integer) 26
```

### Example 2: Counting bits in a specified range

```bash
127.0.0.1:7379> SET mykey "foobar"
OK
127.0.0.1:7379> BITCOUNT mykey 1 3
(integer) 15
```

In this example, the command counts the bits set to 1 in the bytes from position 1 to 3 (inclusive) of the string "foobar".

### Example 3: Counting bits in a specified range using bit indices

```bash
127.0.0.1:7379> SET mykey "foobar"
OK
127.0.0.1:7379> BITCOUNT mykey 8 31 BIT
(integer) 15
```

In this example, the command counts the bits set to 1 in the bit ranges from position 8 to 31 (inclusive) of the string "foobar".
