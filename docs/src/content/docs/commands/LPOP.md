---
title: LPOP
description: The `LPOP` command in DiceDB removes and returns the first element of a list at a specified key. This command is ideal for implementing queue-like data structures that process elements in a First-In-First-Out (FIFO) order.
---

The `LPOP` command in DiceDB removes and returns the first element of a list at a specified key. This command is ideal for implementing queue-like data structures that process elements in a First-In-First-Out (FIFO) order.

## Syntax

```bash
LPOP key
```

## Parameters

| Parameter   | Description                                                                     | Type    | Required |
| ----------- | --------------------------------------------------------------------------------| ------- | -------- |
| `key`       | The key of the list from which the first element will be removed and returned.  | String  | Yes      |



## Return Value

| Condition                                      | Return Value                                                                              |
|------------------------------------------------|-------------------------------------------------------------------------------------------|
| Command is successful                          | `String` The value of the first element in the list, if the list exists and is not empty. |
| If the key does not exist or the list is empty | `nil`                                                                                     |
| Syntax or specified constraints are invalid    |  error                                                                                    |



## Behavior

- When the `LPOP` command is executed, DiceDB checks if the key exists and is associated with a list. If so, the first element of the list is removed and returned.
- If the key does not exist or the list is empty, the command returns `nil`.
- If the key exists but is not associated with a list, a `WRONGTYPE` error is returned.
- If more than one argument is passed, an error is returned.

## Errors

1. `Wrong type of key`

When the `LPOP` command is executed for a key that exists but is not associated with a list, an error is returned. This error occurs if the key is associated with a type other than a list, such as a string or set.

```bash
127.0.0.1:7379> LPOP mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```
2. `Wrong number of arguments`

If the `LPOP` command is executed with more than one key or no key, the following error is returned:

```bash
127.0.0.1:7379> LPOP
(error) ERR wrong number of arguments for 'lpop' command
```

```bash
127.0.0.1:7379> LPOP mylist secondlist
(error) ERR wrong number of arguments for 'lpop' command
```

## Example Usage

### Basic Usage
Setting a list `mylist` with elements \["one", "two", "three"\].

```bash
RPUSH mylist "one" "two" "three"
```
Removing the first element from the list `mylist`

```bash
LPOP mylist
"one"
```

The list `mylist` now contains \["two", "three"\].

If the LPOP command is executed one more time:
```bash
LPOP mylist
"two"
```

The list `mylist` now contains \["three"\].



### Empty List or Non-Existent Key

Returns `(nil)` if the list is empty or the provided key is non-existent

```bash
LPOP emptylist
(nil)
```




### Key is Not a List

Setting a key `mystring` with the value `hello`:

```bash
SET mystring "hello"
OK
```

Executing `LPOP` command on any key that is not associated with a List type will result in an error:

```bash
LPOP mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Wrong Number of Arguments

Passing more than one key will result in an error:

```bash
LPOP mylist secondlist
(error) ERR wrong number of arguments for 'lpop' command
```

### Notes

- The key used with the `LPOP` command should be associated with a list to avoid any potential errors.
- Use the `EXISTS` command to check if a key exists before using `LPOP` to handle cases where the key might not exist.
- Consider using `LLEN` to check the length of the list if you need to handle empty lists differently.

## Related Commands

- `RPUSH`: Append one or multiple elements to the end of a list.
- `LPUSH`: Prepend one or multiple elements to the beginning of a list.
- `RPOP`: Remove and return the last element of a list.
- `LRANGE`: Get a range of elements from a list.

By understanding and using the `LPOP` command effectively, you can manage list data structures in DiceDB efficiently, implementing queue-like behaviors and more.

