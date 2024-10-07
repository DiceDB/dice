---
title: DECRBY
description: The `DECRBY` command in DiceDB is used to decrement the integer value of a key by a specified amount. This command is useful for scenarios where you need to decrease a counter or a numeric value stored in a key.
---

The `DECRBY` command in DiceDB is used to decrement the integer value of a key by a specified amount. This command is useful for scenarios where you need to decrease a counter or a numeric value stored in a key.

## Syntax

```
DECRBY key delta
```

## Parameters

| Parameter | Description                                                                                                   | Type    | Required |
|-----------|---------------------------------------------------------------------------------------------------------------|---------|----------|
| `key`     | The key whose value you want to decrement. This key must hold a string that can be represented as an integer. | String  | Yes      |
|`delta`    | The integer value by which the key's value should be decreased. This value can be positive or negative.       | String  | Yes      |


## Return values

| Condition                                        | Return Value                                                     |
|--------------------------------------------------|------------------------------------------------------------------|
| Key exists and holds an integer string           | `(integer)` The value of the key after decrementing by delta.    |
| Key does not exist                               | `(integer)` -delta                                               |


## Behaviour
When the `DECRBY` command is executed, the following steps occur:

-  DiceDB checks if the key exists.
-  If the key does not exist, DiceDB treats the key's value as 0 before performing the decrement operation.
-  If the key exists but does not hold a string that can be represented as an integer, an error is returned.
-  The value of the key is decremented by the specified decrement value.
-  The new value of the key is returned.
## Errors

The `DECRBY` command can raise errors in the following scenarios:

1. `Wrong Type Error`:

   - Error Message: `ERROR  value is not an integer or out of range`
   - This error occurs if the decrement value provided is not a valid integer.
   - This error occurs if the key exists but its value is not a string that can be represented as an integer

2. `Syntax Error`:

   - Error Message: `ERROR wrong number of arguments for 'decrby' command`
   - Occurs if the command is called without the required parameter.


## Examples

### Example with Decrementing the Value of an Existing Key


```bash
127.0.0.1:7379>SET mycounter 10
OK
127.0.0.1:7379>DECRBY mycounter 3
(integer)7
```
`Explanation:` 

- In this example, the value of `mycounter` is set to 10
- The `DECRBY` command decremented `mycounter`by 3, resulting in a new value of 7.

### Example with Decrementing a Non-Existent Key (Implicit Initialization to 0)

```bash
127.0.0.1:7379>DECRBY newcounter 5
(integer)-5
```
`Explanation:` 
- In this example, since `newcounter` does not exist, DiceDB treats its value as 0 and decrements it by 5, resulting in a new value of -5.
### Example with Error Due to Non-Integer Value in Key

```bash
127.0.0.1:7379>SET mystring "hello"
OK
127.0.0.1:7379>DECRBY mystring 2
(error) ERROR value is not an integer or out of range
```
`Explanation:` 
- In this example, the key `mystring` holds a non-integer value, so the `DECRBY` command returns an error.

### Example with Error Due to Invalid Decrement Value (Non-Integer Decrement)

```bash
127.0.0.1:7379>DECRBY mycounter "two"
(error) ERROR value is not an integer or out of range
```

`Explanation:` 
- In this example, the decrement value "two" is not a valid integer, so the `DECRBY` command returns an error.


