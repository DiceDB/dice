---
title: GETBIT
description: Documentation for the DiceDB command GETBIT
---

The `GETBIT` command is used to retrieve the bit value at a specified offset in the string value stored at a given key. This command is particularly useful for bitwise operations and managing binary data within DiceDB.

## Syntax

```
GETBIT key offset
```

## Parameters

- `key`: The key of the string from which the bit value is to be retrieved. This key must reference a string value.
- `offset`: The position of the bit to retrieve. The offset is a zero-based integer, meaning the first bit is at position 0.

## Return Value

- `Integer`: The command returns the bit value at the specified offset, which will be either `0` or `1`.

## Behaviour

When the `GETBIT` command is executed, DiceDB will:

1. Check if the specified key exists.
2. If the key does not exist, it is treated as if it contains a string of zero bytes, and the bit at any offset will be `0`.
3. If the key exists but does not hold a string value, an error is returned.
4. If the key exists and holds a string value, the bit at the specified offset is retrieved and returned.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is returned if the key exists but does not contain a string value.
- `ERR bit offset is not an integer or out of range`: This error is returned if the offset is not a valid integer or is negative.

## Example Usage

### Example 1: Retrieving a bit from a string

```shell
SET mykey "a"          # ASCII value of 'a' is 97, binary representation is 01100001
GETBIT mykey 1         # Returns 1, as the second bit of 'a' (01100001) is 1
```

### Example 2: Retrieving a bit from a non-existent key

```shell
GETBIT nonExistentKey 5  # Returns 0, as the key does not exist and is treated as a string of zero bytes
```

### Example 3: Error when key holds a non-string value

```shell
LPUSH mylist "item"    # Create a list
GETBIT mylist 0        # Returns an error: WRONGTYPE Operation against a key holding the wrong kind of value
```

### Example 4: Error with invalid offset

```shell
SET mykey "a"
GETBIT mykey -1        # Returns an error: ERR bit offset is not an integer or out of range
GETBIT mykey "abc"     # Returns an error: ERR bit offset is not an integer or out of range
```

## Notes

- The `GETBIT` command operates on the raw binary representation of the string. This means that the offset is counted in bits, not bytes.
- The maximum offset that can be specified is 2^32 - 1 (4294967295), as DiceDB strings are limited to 512 MB.

By understanding and utilizing the `GETBIT` command, you can efficiently manage and manipulate binary data within your DiceDB database.

