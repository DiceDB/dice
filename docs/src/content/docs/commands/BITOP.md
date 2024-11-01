---
title: BITOP
description: Documentation for the DiceDB command BITOP
---

The `BITOP` command in DiceDB is used to perform bitwise operations between strings. This command supports several bitwise operations such as AND, OR, XOR, and NOT. The result of the operation is stored in a destination key.

## Syntax

```bash
BITOP operation destkey key [key ...]
```

## Parameters

- `operation`: The bitwise operation to perform. It can be one of the following:
  - `AND`: Perform a bitwise AND operation.
  - `OR`: Perform a bitwise OR operation.
  - `XOR`: Perform a bitwise XOR operation.
  - `NOT`: Perform a bitwise NOT operation (only one key is allowed for this operation).
- `destkey`: The key where the result of the bitwise operation will be stored.
- `key [key ...]`: One or more keys containing the strings to be used in the bitwise operation. For the `NOT` operation, only one key is allowed.

## Return Value

The command returns the size of the string stored in the destination key, which is equal to the size of the longest input string.

## Behaviour

When the `BITOP` command is executed, it performs the specified bitwise operation on the provided keys and stores the result in the destination key. The length of the resulting string is determined by the longest input string. If the input strings are of different lengths, the shorter strings are implicitly padded with zero-bytes to match the length of the longest string.

### Bitwise Operations

- `AND`: Each bit in the result is set to 1 if the corresponding bits of all input strings are 1. Otherwise, it is set to 0.
- `OR`: Each bit in the result is set to 1 if at least one of the corresponding bits of the input strings is 1. Otherwise, it is set to 0.
- `XOR`: Each bit in the result is set to 1 if an odd number of the corresponding bits of the input strings are 1. Otherwise, it is set to 0.
- `NOT`: Each bit in the result is set to the complement of the corresponding bit in the input string (only one input string is allowed).

## Error Handling

The `BITOP` command can raise errors in the following cases:

1. `Wrong number of arguments`: If the number of arguments is incorrect, DiceDB will return an error.

   - `Error Message`: `ERR wrong number of arguments for 'bitop' command`

2. `Invalid operation`: If the specified operation is not one of `AND`, `OR`, `XOR`, or `NOT`, DiceDB will return an error.

   - `Error Message`: `ERR syntax error`

3. `Invalid number of keys for NOT operation`: If more than one key is provided for the `NOT` operation, DiceDB will return an error.

   - `Error Message`: `ERR BITOP NOT must be called with a single source key`

4. `Non-string keys`: If any of the provided keys are not strings, DiceDB will return an error.

   - `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`

## Example Usage

### Example 1: Bitwise AND Operation

```bash
SET key1 "foo"
SET key2 "bar"
BITOP AND result key1 key2
GET result
```

`Explanation`:

- `SET key1 "foo"`: Sets the value of `key1` to "foo".
- `SET key2 "bar"`: Sets the value of `key2` to "bar".
- `BITOP AND result key1 key2`: Performs a bitwise AND operation between the values of `key1` and `key2`, and stores the result in `result`.
- `GET result`: Retrieves the value of `result`.

### Example 2: Bitwise OR Operation

```bash
SET key1 "foo"
SET key2 "bar"
BITOP OR result key1 key2
GET result
```

`Explanation`:

- `SET key1 "foo"`: Sets the value of `key1` to "foo".
- `SET key2 "bar"`: Sets the value of `key2` to "bar".
- `BITOP OR result key1 key2`: Performs a bitwise OR operation between the values of `key1` and `key2`, and stores the result in `result`.
- `GET result`: Retrieves the value of `result`.

### Example 3: Bitwise XOR Operation

```bash
SET key1 "foo"
SET key2 "bar"
BITOP XOR result key1 key2
GET result
```

`Explanation`:

- `SET key1 "foo"`: Sets the value of `key1` to "foo".
- `SET key2 "bar"`: Sets the value of `key2` to "bar".
- `BITOP XOR result key1 key2`: Performs a bitwise XOR operation between the values of `key1` and `key2`, and stores the result in `result`.
- `GET result`: Retrieves the value of `result`.

### Example 4: Bitwise NOT Operation

```bash
SET key1 "foo"
BITOP NOT result key1
GET result
```

`Explanation`:

- `SET key1 "foo"`: Sets the value of `key1` to "foo".
- `BITOP NOT result key1`: Performs a bitwise NOT operation on the value of `key1`, and stores the result in `result`.
- `GET result`: Retrieves the value of `result`.
