---
title: INCRBY
description: The `INCRBY` command in DiceDB is used to increment the integer value of a key by a specified amount. This command is useful for scenarios where you need to increase a counter or a numeric value stored in a key.
---

The `INCRBY` command in DiceDB is used to increment the integer value of a key by a specified amount. This command is useful for scenarios where you need to increase a counter or a numeric value stored in a key.

## Syntax

```bash
INCRBY key delta
```

## Parameters

| Parameter | Description                                                                                                   | Type    | Required |
|-----------|---------------------------------------------------------------------------------------------------------------|---------|----------|
| `key`     | The key whose value you want to increment. This key must hold a string that can be represented as an integer. | String  | Yes      |
|`delta`    | The integer value by which the key's value should be increased. This value can be positive or negative.       | String  | Yes      |


## Return values

| Condition                                        | Return Value                                                     |
|--------------------------------------------------|------------------------------------------------------------------|
| Key exists and holds an integer string           | `(integer)` The value of the key after incrementing by delta.    |
| Key does not exist                               | `(integer)` delta                                               |


## Behaviour
When the `INCRBY` command is executed, the following steps occur:

-  DiceDB checks if the key exists.
-  If the key does not exist, DiceDB treats the key's value as 0 before performing the increment operation.
-  If the key exists but does not hold a string that can be represented as an integer, an error is returned.
-  The value of the key is incremented by the specified increment value.
-  The new value of the key is returned.
## Errors

The `INCRBY` command can raise errors in the following scenarios:

1. `Wrong Type Error`:

   - Error Message: `ERR  value is not an integer or out of range`
   - This error occurs if the increment value provided is not a valid integer.
   - This error occurs if the key exists but its value is not a string that can be represented as an integer

2. `Syntax Error`:

   - Error Message: `ERR wrong number of arguments for 'incrby' command`
   - Occurs if the command is called without the required parameter.

3. `Overflow Error`:

   - Error Message: `ERR increment or decrement would overflow`
   - If the increment operation causes the value to exceed the maximum integer value that DiceDB can handle, an overflow error will occur.


## Examples

### Example with Incrementing the Value of an Existing Key


```bash
127.0.0.1:7379>SET mycounter 10
OK
127.0.0.1:7379>INCRBY mycounter 3
(integer)13
```
`Explanation:` 

- In this example, the value of `mycounter` is set to 10
- The `INCRBY` command incremented `mycounter` by 3, resulting in a new value of 13.

### Example with Incrementing a Non-Existent Key (Implicit Initialization to 0)

```bash
127.0.0.1:7379>INCRBY newcounter 5
(integer)5
```
`Explanation:` 
- In this example, since `newcounter` does not exist, DiceDB treats its value as 0 and increments it by 5, resulting in a new value of 5.
### Example with Error Due to Non-Integer Value in Key

```bash
127.0.0.1:7379>SET mystring "hello"
OK
127.0.0.1:7379>INCRBY mystring 2
(error) ERR value is not an integer or out of range
```
`Explanation:` 
- In this example, the key `mystring` holds a non-integer value, so the `INCRBY` command returns an error.

### Example with Error Due to Invalid Increment Value (Non-Integer Decrement)

```bash
127.0.0.1:7379>INCRBY mycounter "two"
(error) ERR value is not an integer or out of range
```

`Explanation:` 
- In this example, the increment value "two" is not a valid integer, so the `INCRBY` command returns an error.


