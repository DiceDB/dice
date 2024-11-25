---
title: HINCRBYFLOAT
description: The `HINCRBYFLOAT` command in DiceDB is used to increment the value of a hash stored at a specified key representing a floating point value. This command is useful when you need to increment or decrement the value associated with a field stored in a hash key using a floating point value.
---

The `HINCRBYFLOAT` command in DiceDB is used to increment the value of a hash stored at a specified key representing a floating point value. This command is useful when you need to increment or decrement the value associated with a field stored in a hash key using a floating point value.

## Syntax

```bash
HINCRBYFLOAT key field increment
```

## Parameters

| Parameter   | Description                                                                                       | Type   | Required |
| ----------- | ------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`       | The key of the hash, which consists of the field whose value you want to increment/decrement      | String | Yes      |
| `field`     | The field present in the hash, whose value you want to increment/decrement                        | String | Yes      |
| `increment` | The floating-point value with which you want to increment/decrement the value stored in the field | Float  | Yes      |

## Return values

| Condition                                                                                                 | Return Value |
| --------------------------------------------------------------------------------------------------------- | ------------ |
| The `key` and the `field` exists and is a hash                                                            | `String`     |
| The `key` exists and is a hash , but the `field` does not exist                                           | `String`     |
| The `key` and `field` do not exist                                                                        | `String`     |
| The `key` exists and is a hash , but the `field` does not have a value that is a valid integer or a float | `error`      |
| The `increment` is not a valid integer/float                                                              | `error`      |

## Behaviour

- DiceDB checks if the increment is a valid integer or a float.
- If not, a error is returned.
- It then checks if the key exists
- If the key is not a hash, a error is returned.
- If the key exists and is of type hash, dicedb then retrieves the key along-with the field and increments the value stored at the specified field.
- If the value stored in the field is not an integer or a float, a error is returned.
- If the key does not exist, a new hash key is created with the specified field and its value is set to the `increment` passed.

## Errors

1. `Wrong type of value or key`:
   - Error Message: `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that is not a hash.
2. `Non-integer / Non-float hash value`:
   - Error Message: `(error) ERROR value is not an integer or a float`
   - Occurs when attempting to increment a value that is not a valid integer or a float
3. `Invalid increment type`
   - Error Message: `(error) ERROR value is not an integer or a float`
   - Occurs when attempting to increment a value with an invalid increment type.
4. `Invalid number of arguments`
   - Error Message: `(error) ERROR wrong number of arguments for 'hincrbyfloat' command`
   - Occurs when an invalid number of arguments are passed to the command.

## Example Usage

### Basic Usage

Executing `hincrbyfloat` on a non-exsiting hash key

```bash
127.0.0.1:7379> HINCRBYFLOAT keys field1 10.2
"10.2"
```

### Usage on a non-existing field

Executing `hincrbyfloat` on a existing hash with a non-existing field

```bash
127.0.0.1:7379> HINCRBYFLOAT keys field2 0.2
"0.2"
```

### Usage on a existing hash and a field

Executing `hincrbyfloat` on a existing hash with a existing field

```bash
127.0.0.1:7379> HINCRBYFLOAT keys field2 1.2
"1.4"
```

### Invalid key usage

Executing `hincrbyfloat` on a non-hash key

```bash
127.0.0.1:7379> SET user:3000 "This is a string"
OK
127.0.0.1:7379> HINCRBYFLOAT user:3000 field1 10
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid number of arguments

Passing invalid number of arguments to the `hincrbyfloat` command

```bash
127.0.0.1:7379> HINCRBYFLOAT user:3000 field
(error) ERROR wrong number of arguments for 'hincrbyfloat' command
```

### Invalid value type

Executing `hincrbyfloat` on a non-integer hash value

```bash
127.0.0.1:7379> HSET user:3000 field "hello"
(integer) 1
127.0.0.1:7379> HINCRBYFLOAT user:3000 field 2
(error) ERROR value is not an integer or a float
```

### Invalid increment type passed

Executing `hincrbyfloat` with a string increment value

```bash
127.0.0.1:7379> HINCRBYFLOAT user:3000 field new
(error) ERROR value is not an integer or a float
```

## Best Practices

-`Ensure Field is Numeric`: Always verify that the field being incremented holds a valid numeric value (integer or float) to avoid errors related to type mismatches. -`Use Appropriate Data Types`: Choose the right data types for your fields. For floating-point operations, ensure that the values are stored and manipulated as floats to maintain precision.

## Alternatives

- [`HINCRBY`](/commands/hincrby): Use this command if you only need to increment a hash field by an integer. It is specifically designed for integer increments and may be more efficient for non-floating-point operations.
- [`HSET`](/commands/hset) and [`HGET`](/commands/hget): If you need to set or retrieve values without incrementing, consider using `HSET` to assign a value directly and `HGET` to retrieve the current value of a field.

## Notes

- The `HINCRBYFLOAT` command is a powerful tool for managing floating-point counters and numerical values stored in hash fields, making it essential for applications that require precision in incremental updates.
- The command operates atomically, meaning it will complete without interruption, making it safe to use in concurrent environments where multiple clients may modify the same hash fields.
- `HINCRBYFLOAT` can be beneficial in scenarios such as tracking scores in a game, maintaining balances in accounts, or managing quantities in inventory systems where floating-point values are common.