---
title: LPUSH
description: The `LPUSH` command is used to insert one or multiple values at the head (left) of a list stored at a specified key. If the key does not exist, a new list is created before performing the push operations. If the key exists but is not a list, an error is returned.
---

The `LPUSH` command is used to insert one or multiple values at the head (left) of a list stored at a specified key. If the key does not exist, a new list is created before performing the push operations. If the key exists but is not a list, an error is returned.

## Syntax

```
LPUSH key value [value ...]
```

## Parameters

| Parameter          | Description                                                                               | Type   | Required |
| ------------------ | ----------------------------------------------------------------------------------------- | ------ | -------- |
| `key`              | The name of the list where values are inserted. If it does not exist, it will be created. | String | Yes      |
| `value [value...]` | One or more values to be inserted at the head of the list.                                | String | Yes      |

## Return Value

| Condition                                   | Return Value                                   |
| ------------------------------------------- | ---------------------------------------------- |
| Command is successful                       | `Integer` - length of the list after execution |
| Syntax or specified constraints are invalid | error                                          |

## Behaviour

- The specified values are inserted at the head of the list provided by the key. 
- If multiple values are provided, they are inserted in the order they are given, with the leftmost value being the first to be inserted. 
- If the key does not exist, a new list is created. 
- If the key exists but is not a list, an error is returned.

## Errors

1. `Wrong Number of Arguments`

    - Error Message: `(error) ERR wrong number of arguments for 'lpush' command`
    - Occurs if the key parameters is not provided or at least one value is not provided. 

2. `Wrong Type of Key or Value`: 
   
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists and is not a list. DiceDB expects the key to either be non-existent or to hold a list data type.

## Example Usage

### Single Value Insertion

Insert the value `world` at the head of the list stored at key `mylist`. If `mylist` does not exist, a new list is created.

```shell
127.0.0.1:7379> LPUSH mylist "world"
(integer) 1
```

### Multiple Values Insertion

Insert the value `hello` and `world` at the head of the list stored at key `mylist`.  After execution, `world` will be the first element, followed by `hello`.
<!-- Once LRANGE command is added, update docs to use LRANGE in examples. -->

```shell
127.0.0.1:7379> LPUSH mylist "hello" "world"
(integer) 2
```

### Creating a New List

Create a new list with the key `newlist` and inserts the value `first` at the head.

```shell
127.0.0.1:7379> LPUSH newlist "first"
(integer) 1
```

### Error Case - Wrong Type

Insert the value `value` at the head of the key `mystring`, which stores a string, not a list.

```shell
127.0.0.1:7379> SET mystring "not a list"
OK
127.0.0.1:7379> LPUSH mystring "value"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `LPUSH` command is often used in conjunction with the `RPUSH` command, which inserts values at the tail (right) of the list.
- The `LPUSH` command can be used to implement a stack (LIFO) by always pushing new elements to the head of the list.

By understanding the `LPUSH` command, you can efficiently manage lists in DiceDB, ensuring that elements are added to the head of the list as needed.

