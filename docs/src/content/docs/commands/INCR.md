---
title: INCR
description: The `INCR` command in DiceDB is used to increment the integer value of a key by one. If the key does not exist, it is set to 0 before performing the increment operation. This command is atomic, meaning that even if multiple clients issue `INCR` commands concurrently, DiceDB ensures that the value is incremented correctly.
---

The `INCR` command in DiceDB is used to increment the integer value of a key by one. If the key does not exist, it is set to 0 before performing the increment operation. This command is atomic, meaning that even if multiple clients issue `INCR` commands concurrently, DiceDB ensures that the value is incremented correctly.

## Syntax

```plaintext
INCR key
```

## Parameters

- `key`: The key whose value you want to increment. This key must hold a string that can be represented as an integer.

## Return Value

- `Integer`: The new value of the key after the increment operation.

## Behaviour

When the `INCR` command is executed, the following steps occur:

1. `Key Existence Check`: DiceDB checks if the specified key exists.
2. `Initialization`: If the key does not exist, DiceDB initializes it to 0.
3. `Type Check`: DiceDB checks if the value stored at the key is a string that can be interpreted as an integer.
4. `Increment Operation`: The value is incremented by 1.
5. `Return New Value`: The new value is returned to the client.

## Error Handling

The `INCR` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If the key exists but does not hold a string value that can be interpreted as an integer, DiceDB will return an error.

   - `Error Message`: `(error) ERR value is not an integer or out of range`

2. `Overflow Error`: If the increment operation causes the value to exceed the maximum integer value that DiceDB can handle, an overflow error will occur.

   - `Error Message`: `(error) ERR increment or decrement would overflow`

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> SET mykey 10
127.0.0.1:7379> INCR mykey
(integer) 11
```

### Key Does Not Exist

```bash
127.0.0.1:7379> INCR newkey
(integer) 1
```

### Error Scenario: Non-Integer Value

```bash
127.0.0.1:7379> SET mykey "hello"
127.0.0.1:7379> INCR mykey
(error) ERR value is not an integer or out of range
```

### Error Scenario: Overflow

```bash
127.0.0.1:7379> SET mykey 9223372036854775807
127.0.0.1:7379> INCR mykey
(error) ERR increment or decrement would overflow
```

## Additional Notes

- The `INCR` command is often used in scenarios where counters are needed, such as counting page views, tracking user actions, or generating unique IDs.
- The atomic nature of the `INCR` command ensures that it is safe to use in concurrent environments without additional synchronization mechanisms.
- For decrementing a value, you can use the `DECR` command, which works similarly but decreases the value by one.
