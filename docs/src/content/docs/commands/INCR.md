---
title: INCR
description: The `INCR` command in DiceDB is used to increment the integer value of a key by one. If the key does not exist, it is set to 0 before performing the increment operation. This command is atomic, meaning that even if multiple clients issue `INCR` commands concurrently, DiceDB ensures that the value is incremented correctly.
---

The `INCR` command in DiceDB is used to increment the integer value of a key by one. If the key does not exist, it is set to 0 before performing the increment operation. This command is atomic, meaning that even if multiple clients issue `INCR` commands concurrently, DiceDB ensures that the value is incremented correctly.

## Syntax

```bash
INCR key
```

## Parameters

- `key`: The key whose value you want to increment. This key must hold a string that can be represented as an integer.

## Return values

| Condition                                   | Return Value |
| ------------------------------------------- | ------------ |
| Command is successful                       | `Integer`    |
| Syntax or specified constraints are invalid | error        |

## Behaviour

When the `INCR` command is executed, the following steps occur:

1. `Key Existence Check`: DiceDB checks if the specified key exists.
2. `Initialization`: If the key does not exist, DiceDB initializes it to 0.
3. `Type Check`: DiceDB checks if the value stored at the key is a string that can be interpreted as an integer.
4. `Increment Operation`: The value is incremented by 1.
5. `Return New Value`: The new value is returned to the client.

## Errors

1. `Wrong Type Error`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - If the key exists but does not hold a string value that can be interpreted as an integer, DiceDB will return an error.

2. `Overflow Error`:

   - Error Message: `(error) ERR increment or decrement would overflow`
   - If the increment operation causes the value to exceed the maximum integer value that DiceDB can handle, an overflow error will occur.

## Example Usage

### Basic Usage

Setting a key `mykey` to `10` and incrementing it by `1`:

```bash
127.0.0.1:7379> SET mykey 10
127.0.0.1:7379> INCR mykey
(integer) 11
```

### Key Does Not Exist

Incrementing a non-existent key `newkey` by `1`:

```bash
127.0.0.1:7379> INCR newkey
(integer) 1
```

### Error Scenario: Non-Integer Value

Incrementing a key `mykey` with a non-integer value:

```bash
127.0.0.1:7379> SET mykey "hello"
127.0.0.1:7379> INCR mykey
(error) ERROR value is not an integer or out of range
```

### Error Scenario: Overflow

Incrementing a key `mykey` with a value that exceeds the maximum integer value:

```bash
127.0.0.1:7379> SET mykey 9223372036854775807
127.0.0.1:7379> INCR mykey
(error) ERROR increment or decrement would overflow
```

## Additional Notes

- The `INCR` command is often used in scenarios where counters are needed, such as counting page views, tracking user actions, or generating unique IDs.
- The atomic nature of the `INCR` command ensures that it is safe to use in concurrent environments without additional synchronization mechanisms.
- For decrementing a value, you can use the [`DECR`](/commands/decr) command, which works similarly but decreases the value by one.

## Alternatives

- You can also use the [`INCRBY`](/commands/incrby) command to increment the value of a key by a specified amount.
- You can also use the [`INCRBYFLOAT`](/commands/incrbyfloat) command to increment the value of a key by a fractional amount.