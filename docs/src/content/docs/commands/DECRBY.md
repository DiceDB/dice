---
title: DECRBY
description: Documentation for the DiceDB command DECRBY
---

The `DECRBY` command in DiceDB is used to decrement the integer value of a key by a specified amount. This command is useful for scenarios where you need to decrease a counter or a numeric value stored in a key.

## Syntax

```
DECRBY key decrement
```

## Parameters

- `key`: The key whose value you want to decrement. This key must hold a string that can be represented as an integer.
- `decrement`: The integer value by which the key's value should be decreased. This value can be positive or negative.

## Return Value

The command returns the value of the key after the decrement operation has been performed.

## Behaviour

When the `DECRBY` command is executed, the following steps occur:

1. DiceDB checks if the key exists.
2. If the key does not exist, DiceDB treats the key's value as 0 before performing the decrement operation.
3. If the key exists but does not hold a string that can be represented as an integer, an error is returned.
4. The value of the key is decremented by the specified decrement value.
5. The new value of the key is returned.

## Error Handling

The `DECRBY` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the key exists but its value is not a string that can be represented as an integer.
- `ERR value is not an integer or out of range`: This error occurs if the decrement value provided is not a valid integer.

## Example Usage

### Example 1: Basic Decrement

```DiceDB
SET mycounter 10
DECRBY mycounter 3
```

`Output:`

```
(integer) 7
```

In this example, the value of `mycounter` is decremented by 3, resulting in a new value of 7.

### Example 2: Decrementing a Non-Existent Key

```DiceDB
DECRBY newcounter 5
```

`Output:`

```
(integer) -5
```

In this example, since `newcounter` does not exist, DiceDB treats its value as 0 and decrements it by 5, resulting in a new value of -5.

### Example 3: Error Handling - Non-Integer Value

```DiceDB
SET mystring "hello"
DECRBY mystring 2
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, the key `mystring` holds a non-integer value, so the `DECRBY` command returns an error.

### Example 4: Error Handling - Invalid Decrement Value

```DiceDB
DECRBY mycounter "two"
```

`Output:`

```
(error) ERR value is not an integer or out of range
```

In this example, the decrement value "two" is not a valid integer, so the `DECRBY` command returns an error.

## Notes

- The `DECRBY` command is atomic, meaning that even if multiple clients issue `DECRBY` commands concurrently, DiceDB ensures that the value is decremented correctly.
- If the key's value is not a valid integer, the command will fail with an error.
- The decrement value can be negative, which effectively makes the `DECRBY` command an increment operation.

By understanding the `DECRBY` command, you can effectively manage and manipulate integer values stored in DiceDB keys, ensuring accurate and efficient data handling in your applications.

