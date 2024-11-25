---
title: HINCRBY
description: The `HINCRBY` command in DiceDB is used to increment the value of a hash stored at a specified key. This command is useful when you need to increment or decrement the value associated with a field stored in a hash key.
---

The `HINCRBY` command in DiceDB is used to increment the value of a hash stored at a specified key. This command is useful when you need to increment or decrement the value associated with a field stored in a hash key.

## Syntax

```bash
HINCRBY key field increment
```

## Parameters

| Parameter   | Description                                                                                  | Type    | Required |
| ----------- | -------------------------------------------------------------------------------------------- | ------- | -------- |
| `key`       | The key of the hash, which consists of the field whose value you want to increment/decrement | String  | Yes      |
| `field`     | The field present in the hash, whose value you want to incremment/decrement                  | String  | Yes      |
| `increment` | The integer with which you want to increment/decrement the value stored in the field         | Integer | Yes      |

## Return values

| Condition                                                                                | Return Value |
| ---------------------------------------------------------------------------------------- | ------------ |
| If `key` and the `field` exists and is a hash                                            | `Integer`    |
| If `key` exists and is a hash , but the `field` does not exist                           | `Integer`    |
| If `key` and `field` do not exist                                                        | `Integer`    |
| If `key` exists and is a hash , but the `field` does not have a value that is an integer | `error`      |
| If `increment` results in an integer overflow                                            | `error`      |
| If `increment` is not a valid integer                                                    | `error`      |

## Behaviour

- DiceDB checks if the increment is a valid integer. If not, an error is returned.
- It then checks if the key exists. If the key is not a hash, an error is returned.
- If the key exists and is of type hash, DiceDB retrieves the key along with the field and increments the value stored at the specified field.
- If the value stored in the field is not an integer, an error is returned.
- If the increment results in an overflow, an error is returned.
- If the key does not exist, a new hash key is created with the specified field and its value is set to the `increment` passed.

## Errors

1. `Wrong type of value or key`:
   - Error Message: `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that is not a hash.
2. `Non-integer hash value`:
   - Error Message: `(error) ERROR hash value is not an integer`
   - Occurs when attempting to increment a value that is not a valid integer
3. `Increment or decrement overflow`:
   - Error Message: `(error) ERROR increment or decrement would overflow`
   - Occurs when attempting to increment a value that results in an integer overflow.
4. `Invalid increment type`
   - Error Message: `(error) ERROR value is not an integer or out of range`
   - Occurs when attempting to increment a value with an invalid increment type.
5. `Invalid number of arguments`
   - Error Message: `(error) ERROR wrong number of arguments for 'hincrby' command`
   - Occurs hwen an invalid number of arguments are passed to the command.

## Example Usage

### Basic Usage

Executing `hincrby` on a non-exsiting hash key

```bash
127.0.0.1:7379> HINCRBY keys field1 10
(integer) 10
```

### Usage on a non-existing field

Executing `hincrby` on a existing hash with a non-existing field

```bash
127.0.0.1:7379> HINCRBY keys field2 10
(integer) 10
```

### Usage on a existing hash and a field

Executing `hincrby` on a existing hash with a existing field

```bash
127.0.0.1:7379> HINCRBY keys field2 10
(integer) 20
```

### Usage by passing a negative increment

Executing `hincrby` to decrement a value

```bash
127.0.0.1:7379> HINCRBY keys field -20
(integer) 0
```

### Invalid key usage

Executing `hincrby` on a non-hash key

```bash
127.0.0.1:7379> SET user:3000 "This is a string"
OK
127.0.0.1:7379> HINCRBY user:3000 field1 10
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid number of arguments

Passing invalid number of arguments to the `hincrby` command

```bash
127.0.0.1:7379> HINCRBY user:3000 field
(error) ERROR wrong number of arguments for 'hincrby' command
```

### Invalid value type

Executing `hincrby` on a non-integer hash value

```bash
127.0.0.1:7379> HSET user:3000 field "hello"
(integer) 1
127.0.0.1:7379> HINCRBY user:3000 field 2
(error) ERROR hash value is not an integer
```

### Invalid increment type passed

Executing `hincrby` with a string increment value

```bash
127.0.0.1:7379> HINCRBY user:3000 field new
(error) ERROR value is not an integer or out of range
```

### Increment resulting in an integer overflow

Incrementing the hash value with a very large integer results in an integer overflow

```bash
127.0.0.1:7379> HSET new-key field 9000000000000000000
(integer) 1
127.0.0.1:7379> HINCRBY new-key field 1000000000000000000
(error) ERROR increment or decrement would overflow
```

## Best Practices

- `Ensure Field is Numeric`: Always ensure that the field being incremented holds a numeric value to avoid errors related to type mismatches.
- `Monitor Keyspace`: Keep track of hash keys and their fields to prevent unexpected behavior due to the creation of new hashes.

## Alternatives

- [`HINCRBYFLOAT`](/commands/hincrbyfloat): If you need to increment a hash field by a floating-point number, consider using the `HINCRBYFLOAT` command, which is specifically designed for that purpose.

## Notes

- The `HINCRBY` command is a powerful tool for managing counters and numerical values stored in hash fields, making it essential for applications that rely on incremental updates.