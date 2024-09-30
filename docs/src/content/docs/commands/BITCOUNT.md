---
title: BITCOUNT
description: Documentation for the DiceDB command BITCOUNT
---

The `BITCOUNT` command in DiceDB is used to count the number of set bits (i.e., bits with value 1) in a string. This command is particularly useful for applications that need to perform bitwise operations and analyze binary data stored in DiceDB.

## Syntax

```plaintext
BITCOUNT key [start end]
```

## Parameters

- `key`: The key of the string for which the bit count is to be calculated. This parameter is mandatory.
- `start`: The starting byte position (inclusive) to count the bits. This parameter is optional. If not specified, the default is the beginning of the string.
- `end`: The ending byte position (inclusive) to count the bits. This parameter is optional. If not specified, the default is the end of the string.

## Return Value

The `BITCOUNT` command returns an integer representing the number of bits set to 1 in the specified range of the string.

## Behaviour

When the `BITCOUNT` command is executed, DiceDB will:

1. Retrieve the string stored at the specified key.
2. If the `start` and `end` parameters are provided, DiceDB will consider only the specified range of bytes within the string.
3. Count the number of bits set to 1 within the specified range or the entire string if no range is specified.
4. Return the count as an integer.

## Example Usage

### Example 1: Counting bits in the entire string

```plaintext
SET mykey "foobar"
BITCOUNT mykey
```

`Output:`

```plaintext
26
```

### Example 2: Counting bits in a specified range

```plaintext
SET mykey "foobar"
BITCOUNT mykey 1 3
```

`Output:`

```plaintext
10
```

In this example, the command counts the bits set to 1 in the bytes from position 1 to 3 (inclusive) of the string "foobar".

## Error Handling

### Common Errors

1. `Non-existent Key`: If the specified key does not exist, DiceDB treats it as an empty string and returns 0.

   ```plaintext
   BITCOUNT non_existent_key
   ```

   `Output:`

   ```plaintext
   0
   ```

2. `Wrong Type of Key`: If the key exists but does not hold a string value, DiceDB will return an error.

   ```plaintext
   LPUSH mylist "element"
   BITCOUNT mylist
   ```

   `Output:`

   ```plaintext
   (error) WRONGTYPE Operation against a key holding the wrong kind of value
   ```

3. `Invalid Range`: If the `start` or `end` parameters are not integers, DiceDB will return an error.

   ```plaintext
   SET mykey "foobar"
   BITCOUNT mykey start end
   ```

   `Output:`

   ```plaintext
   (error) ERR value is not an integer or out of range
   ```

4. `Out of Range Indices`: If the `start` or `end` parameters are out of the range of the string length, DiceDB will handle it gracefully by considering the valid range within the string.

   ```plaintext
   SET mykey "foobar"
   BITCOUNT mykey 10 20
   ```

   `Output:`

   ```plaintext
   0
   ```

## Notes

- The `start` and `end` parameters are byte positions, not bit positions.
- Negative values for `start` and `end` are interpreted as offsets from the end of the string. For example, -1 means the last byte, -2 means the second last byte, and so on.
