---
title: HRANDFIELD
description: Documentation for the DiceDB command HRANDFIELD
---


The `HRANDFIELD` command in DiceDB is used to return one or more random fields from a hash stored at a specified key. It can also return the values associated with those fields if specified.

## Syntax

```
HRANDFIELD key [count [WITHVALUES]]
```

## Parameters

| Parameter   | Description                                                                | Type    | Required |
|-------------|----------------------------------------------------------------------------|---------|----------|
| `key`       | The key of the hash from which random fields are to be returned            | String  | Yes      |
| `count`     | The number of random fields to retrieve. If negative, allows repetition    | Integer | No       |
| `WITHVALUES`| Option to include the values associated with the returned fields           | Flag    | No       |

## Return values

| Condition                                            | Return Value                                    |
|------------------------------------------------------|-------------------------------------------------|
| Key exists and count is not specified                | One random field `(String)`                       |
| Key exists and count is specified                    | Array of random fields (or field-value pairs if `WITHVALUES` is used) |
| Key does not exist                                   | `nil`                                           |
| Key exists but is not a hash                         | `error`                                         |

## Behaviour

When the `HRANDFIELD` command is executed:

1. DiceDB checks if the specified `key` exists.
2. If the key does not exist, the command returns `nil`.
3. If the key exists but is not a hash, an error is returned.
4. If the key does not have any fields, an empty array is returned.
5. If no `count` parameter is passed, it is defaulted to `1` 
6. If the `count` or `WITHVALUES` parameters are passed,they are checked for typechecks and syntax errors.
7. If the `count` parameter is negative, the command allows repeated fields in the result.
8. The command will return the random field(s) based on the specified `count`.
9. If the `WITHVALUES` option is provided, the command returns the fields along with their associated values.

## Errors

The `HRANDFIELD` command can raise the following errors:

- `(error) WRONGTYPE Operation against a key holding the wrong kind of value`: Raised when the key exists but is not associated with a hash data structure.
- `(error) ERROR wrong number of arguments for 'hrandfield' command` : This error is raised when invalid number of arguments are passed to the command.
- `(error) ERROR value is not an integer or out of range` : This error is raised when a non-integer value is passed as the `count` parameter.
## Example Usage

### Example 1: Getting a random field from a hash

```bash
127.0.0.1:6379> HSET keys field1 value1 field2 value2 field3 value3
(integer) 3
127.0.0.1:6379> HRANDFIELD keys
```
`Output:`
```bash
"field1"
```

### Example 2: Getting two random fields from a hash

```bash
127.0.0.1:6379> HRANDFIELD keys 2
```
`Output:`
```bash
1) "field2"
2) "field1"
```

### Example 3: Getting random fields and their values

```bash
127.0.0.1:6379> HRANDFIELD keys 2 WITHVALUES
```
`Output:`
```bash
1) "field2"
2) "value2"
3) "field1"
4) "value1"
```

### Error Example: 
Key is not a hash

```bash
127.0.0.1:6379> SET key "not a hash"
OK
127.0.0.1:6379> HRANDFIELD key
```
`Output:`

```bash
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Error Example: 
Non-integer value passed as `count` 

```bash
127.0.0.1:6379> HRANDFIELD keys hello
```
`Output:`
```bash
(error) ERROR value is not an integer or out of range
```

### Error Example: 
Invalid number of arguments

```bash
127.0.0.1:6379> HRANDFIELD
```
`Output:`
```bash
(error) ERR wrong number of arguments for 'hrandfield' command
```