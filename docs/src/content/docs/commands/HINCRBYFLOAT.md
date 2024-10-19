---
title: HINCRBYFLOAT
description: Documentation for the DiceDB command HINCRBYFLOAT
---

The `HINCRBYFLOAT` command in DiceDB is used to increment the value of a hash stored at a specified key representing a floating point value. This command is useful when you need to increment/decrement the value associated to a field stored in a hash key using a floating point value.

## Syntax

```
HINCRBYFLOAT key field increment
```

## Parameters

| Parameter       | Description                                      | Type    | Required |
|-----------------|--------------------------------------------------|---------|----------|
| `key`           | The key of the hash, which consists of the field whose value you want to increment/decrement                  | String  | Yes      |
| `field`           | The field present in the hash, whose value you want to increment/decrement          | String  | Yes      |
| `increment`           | The floating-point value with which you want to increment/decrement the value stored in the field         | Float  | Yes      |


## Return values


| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| The `key` and the `field` exists and is a hash | `String`
| The `key` exists and is a hash , but the `field` does not exist         |`String`                                            |
| The `key`  and `field` do not exist  | `String`               |
| The `key` exists and is a hash , but the `field` does not have a value that is a valid integer or a float  | `error`               |
| The `increment` is not a valid integer/float | `error`               |


## Behaviour

When the `HINCRBYFLOAT` command is executed:

1. DiceDB checks if the increment is a valid integer or a float.
2. If not, a error is returned.
3. It then checks if the key exists
4. If the key is not a hash, a error is returned.
5. If the key exists and is of type hash, DiceDB then retrieves the key along-with the field and increments the value stored at the specified field.
6. If the value stored in the field is not an integer or a float, a error is returned.
7. If the key does not exist, a new hash key is created with the specified field and its value is set to the 'increment' passed.


## Errors

The `HINCRBYFLOAT` command can raise the following errors:

- `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the specified key exists but is not a hash. DiceDB expects the key to be associated with a hash data structure, and if it is associated with a different data type (e.g., string, list, set, etc.), this error will be triggered.
- `(error) ERROR value is not an integer or a float` : This error is raised when the hash value to be incremented is not a valid integer or a float
- `(error) ERROR value is not an integer or a float` : This error is raised when the increment passed is not a valid integer or a float.
- `(error) ERROR wrong number of arguments for 'hincrbyfloat' command` : This error is raised when invalid number of arguments are passed to the command.

## Example Usage

### Example 1: Incrementing a non-existing hash key

```bash
127.0.0.1:7379> HINCRBYFLOAT keys field1 10.2
```
`Output:`
```bash
"10.2"
```

### Example 2: Incrementing an existing hash key with a non-existing field

```bash
127.0.0.1:7379> HINCRBYFLOAT keys field2 0.2
```
`Output:`
```bash
"0.2"
```

### Example 3: Incrementing an existing hash key with an existing field

```bash
127.0.0.1:7379> HINCRBYFLOAT keys field2 1.2
```
`Output:`
```bash
"1.4"
```

### Error Example:
Key is not a hash

```DiceDB
127.0.0.1:7379> SET user:3000 "This is a string"
OK
127.0.0.1:7379> HINCRBYFLOAT user:3000 field1 10
```

`Output:`

```
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

### Error Example: 
Invalid number of arguments are passed

```bash
127.0.0.1:7379> HINCRBYFLOAT user:3000 field
```
`Output:`

```bash
(error) ERROR wrong number of arguments for 'hincrbyfloat' command
```

### Error Example: 
Hash value to be incremented is not an integer or a float

```bash
127.0.0.1:7379> HSET user:3000 field "hello"
(integer) 1
127.0.0.1:7379> HINCRBYFLOAT user:3000 field 2
```
`Output:`

```bash
(error) ERROR value is not an integer or a float
```


### Error Example: 
Increment value passed is not a valid integer or a float

```bash
127.0.0.1:7379> HINCRBYFLOAT user:3000 field new
```
`Output:`

```bash
(error) ERROR value is not an integer or a float
```