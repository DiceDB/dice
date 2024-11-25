---
title: DECR
description: Documentation for the DiceDB command DECR
---

The `DECR` command in DiceDB is used to decrement the integer value of a key by one. If the key does not exist, it is set to 0 before performing the decrement operation. This command is useful for counters and other numerical operations where you need to decrease the value stored at a specific key.

## Syntax

```bash
DECR key
```

## Parameters

| Parameter | Description                                                                                                   | Type   | Required |
| --------- | ------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key whose value you want to decrement. This key must hold a string that can be represented as an integer. | String | Yes      |

## Return Value

| Condition                                   | Return Value                                                       |
| ------------------------------------------- | ------------------------------------------------------------------ |
| Command is successful                       | `Integer`: The new value of the key after the decrement operation. |
| Syntax or specified constraints are invalid | error                                                              |

## Behaviour

When the `DECR` command is executed, the following steps occur:

- If the specified key doesn't exist, it is created, with -1 as it's value, and same is returned.
- If the specified key exists with an integer value, value is decremented and new value is returned.
- If the specified key exists with a non-integer OR out-of-range value, error message is returned as:
  - `(error) ERR value is not an integer or out of range`

## Errors

The `DECR` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If the key exists but the value is not a string that can be represented as an integer, DiceDB will return an error.
   - `Error Message`: `(error) ERR value is not an integer or out of range`
1. `Out of Range Error`: If the value of the key is out of the range of a 64-bit signed integer after the decrement operation, DiceDB will return an error.
   - `Error Message`: `(error) ERR increment or decrement would overflow`

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> SET mycounter 10
OK
127.0.0.1:7379> DECR mycounter
9
```

`Explanation`:

1. The `SET` command initializes the key `mycounter` with the value `10`.
2. The `DECR` command decrements the value of `mycounter` by 1, resulting in `9`.

### Key Does Not Exist

```bash
127.0.0.1:7379> DECR newcounter
(integer) -1
```

`Explanation`:

1. The key `newcounter` does not exist.
2. DiceDB sets `newcounter` to `0` and then decrements it by 1, resulting in `-1`.

### Error Scenario: Non-Integer Value

```bash
127.0.0.1:7379> SET mystring "hello"
OK
127.0.0.1:7379> DECR mystring
(error) ERR value is not an integer or out of range
```

`Explanation`:

1. The `SET` command initializes the key `mystring` with the value `"hello"`.
2. The `DECR` command attempts to decrement the value of `mystring`, but since it is not an integer, an error is raised.

### Error Scenario: Out of Range

```bash
127.0.0.1:7379> SET mycounter 234293482390480948029348230948
OK
127.0.0.1:7379> DECR mycounter
(error) ERR value is not an integer or out of range
```

`Explanation`:

1. The `SET` command initializes the key `mycounter` with the out-of-range value for a 64-bit signed integer.
1. The `DECR` command attempts to decrement the value of `mycounter`, but this would result in an overflow, so an error is raised.

## Alternatives

You can also use the [`DECRBY`](/commands/decrby) command to decrement the value of a key by a specified amount.