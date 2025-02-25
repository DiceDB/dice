---
title: INCRBYFLOAT
description: The `INCRBYFLOAT` command in DiceDB is used to increment the numeric value of a key by a fractional amount. This command is useful for scenarios where you need to increase a number by a fractional amount.
---

The `INCRBYFLOAT` command in DiceDB is used to increment the numeric value of a key by a fractional amount. This command is useful for scenarios where you need to increase a number by a fractional amount.

## Syntax

```bash
INCRBYFLOAT key delta
```

## Parameters

| Parameter | Description                                                                                                  | Type   | Required |
| --------- | ------------------------------------------------------------------------------------------------------------ | ------ | -------- |
| `key`     | The key whose value you want to increment. This key must hold a string that can be represented as an number. | String | Yes      |
| `delta`   | The fractional value by which the key's value should be increased. This value can be positive or negative.   | String | Yes      |

## Return values

| Condition                              | Return Value                                                |
| -------------------------------------- | ----------------------------------------------------------- |
| Key exists and holds an numeric string | `(float)` The value of the key after incrementing by delta. |
| Key does not exist                     | `(float)` delta                                             |

## Behaviour

When the `INCRBYFLOAT` command is executed, the following steps occur:

- DiceDB checks if the key exists.
- If the key does not exist, DiceDB treats the key's value as 0 before performing the increment operation.
- If the key exists but does not hold a string that can be represented as an number, an error is returned.
- The value of the key is incremented by the specified increment value.
- The new value of the key is returned.

## Errors

The `INCRBYFLOAT` command can raise errors in the following scenarios:

1. `Wrong Type Error`:

   - Error Message: `ERR  value is not a valid float`
   - This error occurs if the increment value provided is not a valid number.
   - This error occurs if the key exists but its value is not a string that can be represented as an number

2. `Syntax Error`:

   - Error Message: `ERR wrong number of arguments for 'incrbyfloat' command`
   - Occurs if the command is called without the required parameter.

3. `Overflow Error`:

   - Error Message: `(error) ERR value is out of range`
   - If the increment operation causes the value to exceed the maximum float value that DiceDB can handle, an overflow error will occur.

## Examples

### Example with Incrementing the Value of an Existing Key

```bash
127.0.0.1:7379>SET mycounter 10
OK
127.0.0.1:7379>INCRBYFLOAT mycounter 3.4
"13.4"
```

- In this example, the value of `mycounter` is set to 10
- The `INCRBYFLOAT` command incremented `mycounter` by 3.4, resulting in a new value of 13.4

### Example with Incrementing a Non-Existent Key (Implicit Initialization to 0)

```bash
127.0.0.1:7379>INCRBYFLOAT newcounter 5.3
"5.3"
```

- In this example, since `newcounter` does not exist, DiceDB treats its value as 0 and increments it by 5.3, resulting in a new value of 5.3.

### Example with Error Due to Wrong Value in Key

```bash
127.0.0.1:7379>SET mystring "hello"
OK
127.0.0.1:7379>INCRBYFLOAT mystring 2.3
(error) ERR value is not a valid float
```

- In this example, the key `mystring` holds a string value, so the `INCRBYFLOAT` command returns an error.

### Example with Error Due to Invalid Increment Value (Non-Integer Decrement)

```bash
127.0.0.1:7379>INCRBYFLOAT mycounter "two"
(error) ERR value is not a valid float
```

- In this example, the increment value "two" is not a valid number, so the `INCRBYFLOAT` command returns an error.
