---
title: BITPOS
description: Documentation for the DiceDB command BITPOS
---

The `BITPOS` command in DiceDB is used to find the position of the first bit set to 1 or 0 in a string. This command is particularly useful for efficiently locating the first occurrence of a bit in a binary string stored in a DiceDB key.

## Syntax

```bash
BITPOS key bit [start] [end]
```

## Parameters

- `key`: The key of the string in which to search for the bit.
- `bit`: The bit value to search for, which can be either 0 or 1.
- `start`: (Optional) The starting byte position to begin the search. If not specified, the search starts from the beginning of the string.
- `end`: (Optional) The ending byte position to end the search. If not specified, the search continues until the end of the string.

## Return Value

- `Integer`: The position of the first bit set to the specified value (0 or 1). The position is returned as a zero-based integer.
- `-1`: If the specified bit is not found within the specified range.

## Behaviour

When the `BITPOS` command is executed, DiceDB scans the string stored at the specified key to find the first occurrence of the specified bit (0 or 1). The search can be limited to a specific range by providing the optional `start` and `end` parameters. If the bit is found, the command returns the position of the first occurrence. If the bit is not found, the command returns -1.

## Error Handling

The `BITPOS` command can raise errors in the following cases:

1. `Non-String Key`: If the key does not hold a string value, an error is raised.

   - `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`

2. `Invalid Bit Value`: If the bit value is not 0 or 1, an error is raised.

   - `Error Message`: `ERR bit is not an integer or out of range`

3. `Invalid Range`: If the `start` or `end` parameters are not valid integers, an error is raised.

   - `Error Message`: `ERR value is not an integer or out of range`

## Example Usage

### Basic Usage

Find the position of the first bit set to 1 in the string stored at key `mykey`:

```bash
SET mykey "foobar"
BITPOS mykey 1
```

`Output`:

```bash
6
```

### Specifying a Range

Find the position of the first bit set to 0 in the string stored at key `mykey`, starting from byte position 2 and ending at byte position 4:

```bash
SET mykey "foobar"
BITPOS mykey 0 2 4
```

`Output`:

```bash
16
```

### Bit Not Found

If the specified bit is not found within the specified range, the command returns -1:

```bash
SET mykey "foobar"
BITPOS mykey 1 2 4
```

`Output`:

```bash
-1
```

### Error Cases

#### Non-String Key

Attempting to use `BITPOS` on a key that holds a non-string value:

```bash
LPUSH mylist "item"
BITPOS mylist 1
```

`Output`:

```bash
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

#### Invalid Bit Value

Using a bit value other than 0 or 1:

```bash
SET mykey "foobar"
BITPOS mykey 2
```

`Output`:

```bash
(error) ERR bit is not an integer or out of range
```

#### Invalid Range

Using non-integer values for the `start` or `end` parameters:

```bash
SET mykey "foobar"
BITPOS mykey 1 "a" "b"
```

`Output`:

```bash
(error) ERR value is not an integer or out of range
```
