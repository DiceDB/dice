---
title: SETBIT
description: Documentation for the DiceDB command SETBIT
---

The `SETBIT` command in DiceDB is used to set or clear the bit at a specified offset in the string value stored at a given key. This command is particularly useful for bitwise operations and can be used to implement various data structures and algorithms efficiently.

## Syntax

```
SETBIT key offset value
```

## Parameters

- `key`: The key of the string where the bit is to be set or cleared. If the key does not exist, a new string value is created.
- `offset`: The position of the bit to be set or cleared. The offset is a zero-based integer, meaning the first bit is at position 0.
- `value`: The value to set the bit to. This must be either 0 or 1.

## Return Value

The command returns the original bit value stored at the specified offset before the `SETBIT` operation was performed. The return value is an integer, either 0 or 1.

## Behaviour

- If the key does not exist, a new string of sufficient length to accommodate the specified offset is created. The string is padded with zero bits.
- If the offset is beyond the current length of the string, the string is automatically extended, and the new bits are set to 0.
- The command only affects the bit at the specified offset and leaves all other bits unchanged.

## Error Handling

- `ERR bit is not an integer or out of range`: This error is raised if the `offset` is not a valid integer or if it is negative.
- `ERR bit is not an integer or out of range`: This error is also raised if the `value` is not 0 or 1.
- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the key exists but does not hold a string value.

## Example Usage

### Setting a Bit

```DiceDB
SETBIT mykey 7 1
```

This command sets the bit at offset 7 in the string value stored at `mykey` to 1. If `mykey` does not exist, it is created and padded with zero bits up to the 7th bit.

### Clearing a Bit

```DiceDB
SETBIT mykey 7 0
```

This command clears the bit at offset 7 in the string value stored at `mykey` to 0.

### Checking the Original Bit Value

```DiceDB
SETBIT mykey 7 1
```

If the bit at offset 7 was previously 0, this command will return 0 and then set the bit to 1.

### Extending the String

```DiceDB
SETBIT mykey 100 1
```

If the string stored at `mykey` is shorter than 101 bits, it will be extended, and all new bits will be set to 0 except for the bit at offset 100, which will be set to 1.

## Error Handling Examples

### Invalid Offset

```DiceDB
SETBIT mykey -1 1
```

This command will raise an error: `ERR bit is not an integer or out of range` because the offset is negative.

### Invalid Value

```DiceDB
SETBIT mykey 7 2
```

This command will raise an error: `ERR bit is not an integer or out of range` because the value is not 0 or 1.

### Wrong Type

```DiceDB
SET mykey "Hello"
SETBIT mykey 7 1
```

If `mykey` holds a string value that is not a bit string, the `SETBIT` command will raise an error: `WRONGTYPE Operation against a key holding the wrong kind of value`.
