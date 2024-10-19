---
title: HINCRBY
description: Documentation for the DiceDB command HINCRBY
---

The `HINCRBY` command in DiceDB is used to increment the value of a hash stored at a specified key. This command is useful when you need to increment/decrement the value associated to a field stored in a hash key.

## Syntax

```
HINCRBY key field increment
```

## Parameters

| Parameter       | Description                                      | Type    | Required |
|-----------------|--------------------------------------------------|---------|----------|
| `key`           | The key of the hash, which consists of the field whose value you want to increment/decrement                 | String  | Yes      |
| `field`           | The field present in the hash, whose value you want to incremment/decrement          | String  | Yes      |
| `increment`           | The integer with which you want to increment/decrement the value stored in the field         | Integer  | Yes      |


## Return values


| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| The `key` and the `field` exists and is a hash | `Integer`
| The `key` exists and is a hash , but the `field` does not exist         |`Integer`                                            |
| The `key`  and `field` do not exist  | `Integer`               |
| The `key` exists and is a hash , but the `field` does not have a value that is an integer   | `error`               |
| The `increment` results in an integer overflow | `error`               |
| The `increment` is not a valid integer | `error`               |


## Behaviour

When the `HINCRBY` command is executed:

1. DiceDB checks if the increment is a valid integer.
2. If not, a error is returned.
3. It then checks if the key exists
4. If the key is not a hash, a error is returned.
5. If the key exists and is of type hash, DiceDB then retrieves the key along-with the field and increments the value stored at the specified field.
6. If the value stored in the field is not an integer, a error is returned.
7. If the increment results in an overflow, a error is returned.
8. If the key does not exist, a new hash key is created with the specified field and its value is set to the 'increment' passed.


## Errors

The `HINCRBY` command can raise the following errors:

- `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the specified key exists but is not a hash. DiceDB expects the key to be associated with a hash data structure, and if it is associated with a different data type (e.g., string, list, set, etc.), this error will be triggered.
- `(error) ERROR hash value is not an integer` : This error is raised when the value to be incremented is not a valid integer
- `(error) ERROR increment or decrement would overflow` : This error is raised if the increment on a value results in an integer overflow.
- `(error) ERROR value is not an integer or out of range` : This error is raised when the increment passed is not a valid integer.
- `(error) ERROR wrong number of arguments for 'hincrby' command` : This error is raised when invalid number of arguments are passed to the command.
## Example Usage
### Example 1: Incrementing a non-existing hash key

```bash
127.0.0.1:7379> HINCRBY keys field1 10
```
`Output:`
```bash
(integer) 10
```

### Example 2: Incrementing an existing hash key with a non-existing field

```bash
127.0.0.1:7379> HINCRBY keys field2 10
```
`Output:`
```bash
(integer) 10
```

### Example 3: Incrementing an existing hash key with an existing field

```bash
127.0.0.1:7379> HINCRBY keys field2 10
```
`Output:`
```bash
(integer) 20
```

### Example 4: Decrementing an existing hash key with an existing field

```bash
127.0.0.1:7379> HINCRBY keys field -20
```
`Output:`
```bash
(integer) 0
```
### Error Example:
Key is not a hash

```bash
127.0.0.1:7379> SET user:3000 "This is a string"
OK
127.0.0.1:7379> HINCRBY user:3000 field1 10
```

`Output:`

```bash
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

### Error Example: 
Invalid number of arguments are passed

```bash
127.0.0.1:7379> HINCRBY user:3000 field
```
`Output:`

```bash
(error) ERROR wrong number of arguments for 'hincrby' command
```

### Error Example: 
Hash value to be incremented is not an integer

```bash
127.0.0.1:7379> HSET user:3000 field "hello"
(integer) 1
127.0.0.1:7379> HINCRBY user:3000 field 2
```
`Output:`

```bash
(error) ERROR hash value is not an integer
```


### Error Example: 
Increment value passed is not a valid integer

```bash
127.0.0.1:7379> HINCRBY user:3000 field new
```
`Output:`

```bash
(error) ERROR value is not an integer or out of range
```


### Error Example: 
Increment the hash value results in an integer overflow

```bash
127.0.0.1:7379> HSET new-key field 9000000000000000000
(integer) 1
127.0.0.1:7379> HINCRBY new-key field 1000000000000000000
```
`Output:`

```bash
(error) ERROR increment or decrement would overflow
```