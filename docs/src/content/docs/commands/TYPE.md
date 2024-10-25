---
title: TYPE
description: The `TYPE` command in DiceDB returns the type of the value stored at a specified key. If the key does not exist, it returns "none". This command is essential for understanding the structure of data stored in DiceDB, facilitating appropriate operations based on the data type.
---

The `TYPE` command in DiceDB returns the type of the value stored at a specified key. If the key does not exist, it returns "none". This command is essential for understanding the structure of data stored in DiceDB, facilitating appropriate operations based on the data type.

## Syntax
```
TYPE key
```

## Parameters

| Parameter | Description                      | Type   | Required |
|-----------|----------------------------------|--------|----------|
| `key`     | The name of the key to be checked | String | Yes      |

## Return values

| Condition                                    | Return Value          |
|----------------------------------------------|-----------------------|
| Key exists and type is found                 | Type of the key value |
| Key does not exist                           | "none"                |
| Syntax or specified constraints are invalid  | error                 |

## Behaviour

When the `TYPE` command is executed, DiceDB will check the type of the value stored at the specified key. The command operates in the following manner:

1. **Key Existence Check**: Checks if the specified key exists in the database.
2. **Return Type of Value**: If the key exists, the command returns the type of the value. Possible types include `string`, `list`, `set`, and `hash`.
3. **Return "none" for Non-existent Keys**: If the key does not exist, the command returns `"none"`.

## Error Handling

The `TYPE` command is generally straightforward, with only one main error scenario:

1. **No Key Provided**: If no key is provided to the `TYPE` command, DiceDB will raise a syntax error.
   - **Error Message**: `(error) ERR wrong number of arguments for 'type' command`

## Example Usage

### Basic Usage
Checking the type of a key `foo` that stores a string:

```bash
127.0.0.1:7379> TYPE foo
"string"
```

### Checking Non-existent Key
Attempting to check the type of a non-existent key:

```bash
127.0.0.1:7379> TYPE nonexistentkey
"none"
```

### Complex Example
Setting keys of different types and then checking their types:

```bash
127.0.0.1:7379> SET mystring "hello"
OK
127.0.0.1:7379> LPUSH mylist "item1"
(integer) 1
127.0.0.1:7379> TYPE mystring
"string"
127.0.0.1:7379> TYPE mylist
"list"
```

### Error Example
Calling `TYPE` without any arguments:

```bash
127.0.0.1:7379> TYPE
(error) ERR wrong number of arguments for 'type' command
```

## Function Logic
The function `evalTYPE` retrieves the type of the value stored at the specified key.
