---
title: LPUSH
description: Documentation for the DiceDB command LPUSH
---

The `LPUSH` command is used to insert one or multiple values at the head (left) of a list stored at a specified key. If the key does not exist, a new list is created before performing the push operations. If the key exists but is not a list, an error is returned.

## Syntax

```
LPUSH key value [value ...]
```

## Parameters

- `key`: The name of the list where the values will be inserted. If the list does not exist, it will be created.
- `value`: One or more values to be inserted at the head of the list. Multiple values can be specified, and they will be inserted in the order they are provided, from left to right.

## Return Value

The command returns an integer representing the length of the list after the push operations.

## Behaviour

When the `LPUSH` command is executed, the specified values are inserted at the head of the list. If multiple values are provided, they are inserted in the order they are given, with the leftmost value being the first to be inserted. If the key does not exist, a new list is created. If the key exists but is not a list, an error is returned.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is returned if the key exists and is not a list. DiceDB expects the key to either be non-existent or to hold a list data type.

## Example Usage

### Single Value Insertion

```shell
LPUSH mylist "world"
```

`Description`: Inserts the value "world" at the head of the list stored at key `mylist`. If `mylist` does not exist, a new list is created.

`Return Value`: `1` (since the list now contains one element)

### Multiple Values Insertion

```shell
LPUSH mylist "hello" "world"
```

`Description`: Inserts the values "hello" and "world" at the head of the list stored at key `mylist`. "hello" will be the first element, followed by "world".

`Return Value`: `3` (assuming `mylist` already contained one element before the operation)

### Creating a New List

```shell
LPUSH newlist "first"
```

`Description`: Creates a new list with the key `newlist` and inserts the value "first" at the head.

`Return Value`: `1` (since the list now contains one element)

### Error Case

```shell
SET mystring "not a list"
LPUSH mystring "value"
```

`Description`: Attempts to insert the value "value" at the head of the key `mystring`, which is not a list but a string.

`Error`: `WRONGTYPE Operation against a key holding the wrong kind of value`

## Notes

- The `LPUSH` command is often used in conjunction with the `RPUSH` command, which inserts values at the tail (right) of the list.
- The `LPUSH` command can be used to implement a stack (LIFO) by always pushing new elements to the head of the list.

By understanding the `LPUSH` command, you can efficiently manage lists in DiceDB, ensuring that elements are added to the head of the list as needed.

