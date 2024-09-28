---
title: DECR
description: Documentation for the DiceDB command DECR
---

The `DECR` command in DiceDB is used to decrement the integer value of a key by one. If the key does not exist, it is set to 0 before performing the decrement operation. This command is useful for counters and other numerical operations where you need to decrease the value stored at a specific key.

## Syntax

```plaintext
DECR key
```

## Parameters

- `key`: The key whose value you want to decrement. This key must hold a string that can be represented as an integer.

## Return Value

- `Integer`: The new value of the key after the decrement operation.

## Behaviour

When the `DECR` command is executed, the following steps occur:

1. `Key Existence Check`: DiceDB checks if the specified key exists.
   - If the key does not exist, DiceDB sets the key to 0 before performing the decrement operation.
   - If the key exists, DiceDB retrieves the current value of the key.
1. `Type Check`: DiceDB ensures that the value associated with the key is a string that can be interpreted as an integer.
1. `Decrement Operation`: DiceDB decrements the integer value of the key by one.
1. `Return New Value`: DiceDB returns the new value of the key after the decrement operation.

## Error Handling

The `DECR` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If the key exists but the value is not a string that can be represented as an integer, DiceDB will return an error.
   - `Error Message`: `(error) ERR value is not an integer or out of range`
1. `Out of Range Error`: If the value of the key is out of the range of a 64-bit signed integer after the decrement operation, DiceDB will return an error.
   - `Error Message`: `(error) ERR increment or decrement would overflow`

## Example Usage

### Basic Usage

```plaintext
SET mycounter 10
DECR mycounter
```

`Explanation`:

1. The `SET` command initializes the key `mycounter` with the value `10`.
1. The `DECR` command decrements the value of `mycounter` by 1, resulting in `9`.

`Return Value`:

```plaintext
(integer) 9
```

### Key Does Not Exist

```plaintext
DECR newcounter
```

`Explanation`:

1. The key `newcounter` does not exist.
1. DiceDB sets `newcounter` to `0` and then decrements it by 1, resulting in `-1`.

`Return Value`:

```plaintext
(integer) -1
```

### Error Scenario: Non-Integer Value

```plaintext
SET mystring "hello"
DECR mystring
```

`Explanation`:

1. The `SET` command initializes the key `mystring` with the value `"hello"`.
1. The `DECR` command attempts to decrement the value of `mystring`, but since it is not an integer, an error is raised.

`Error Message`:

```plaintext
(error) ERR value is not an integer or out of range
```

### Error Scenario: Out of Range

```plaintext
SET mycounter 9223372036854775807
DECR mycounter
```

`Explanation`:

1. The `SET` command initializes the key `mycounter` with the maximum value for a 64-bit signed integer.
1. The `DECR` command attempts to decrement the value of `mycounter`, but this would result in an overflow, so an error is raised.

`Error Message`:

```plaintext
(error) ERR increment or decrement would overflow
```
