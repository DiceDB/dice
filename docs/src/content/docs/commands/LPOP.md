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

| Parameter | Description                                                                    | Type   | Required |
| --------- | ------------------------------------------------------------------------------ | ------ | -------- |
| `key`     | The key of the list from which the first element will be removed and returned. | String | Yes      |
| `count`   | The count of elemments to be removed.                                          | Number | No       |


## Return Value

| Condition                    | Return Value                                         |
| ---------------------------- | ---------------------------------------------------- |
| Command is successful        | `String` The value of the first element in the list. |
| If the key does not exist    | `nil`                                                |
| The key is of the wrong type | error                                                |
| The count is invalid         | error                                                |


## Behavior

1. **Key Existence and Type**:
   - If the key does not exist, the command treats it as an empty list and returns `nil`.
   - If the key exists but is not associated with a list, a `WRONGTYPE` error is returned.
   - If more than one key is provided, an error is returned.

2. **Without `count` Parameter**:
   - If the list has elements and the optional `count` parameter is not provided, the first element of the list is removed and returned.

3. **With `count` Parameter**:
   - If the `count` parameter is provided:
     - If the list has elements, up to `count` elements are removed from the start of the list and returned in order (from the first to the last removed element).
     - If `count` is greater than the number of elements in the list, all elements are removed and returned.
     - If `count` is less than 0, the following error is returned:
       ```bash
       (error) value is out of range, must be positive
       ```

4. **Empty List**:
   - If the list is empty, the command returns `nil`.



## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not associated with a list.

```bash
127.0.0.1:7379> LPOP mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

2. `Wrong number of arguments`

   - Error Message: `(error) ERR wrong number of arguments for 'lpop' command`
   - Occurs if command is executed without any arguments or with 2 or more arguments

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

### Non-Existent Key

Returns `(nil)` if the provided key is non-existent

```bash
LPOP emptylist
(nil)
```

### Invalid Usage: Key is Not a List

Trying to `LPOP` a key `mystring` which is holding wrong data type `string`.

```bash
SET mystring "hello"
OK
LPOP mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid Usage: Wrong Number of Arguments

Passing more than one key will result in an error:

```bash
LPOP mylist secondlist
(error) value is out of range, must be positive
```

## Best Practices

- `Check Key Type`: Before using `LPOP`, ensure that the key is associated with a list to avoid errors.
- `Handle Non-Existent Keys`: Be prepared to handle the case where the key does not exist, as `LPOP` will return `nil` in such scenarios.
- `Use in Conjunction with Other List Commands`: The `LPOP` command is often used alongside other list commands like [`RPUSH`](/commands/rpush), [`LPUSH`](/commands/lpush), [`LLEN`](/commands/llen), and [`RPOP`](/commands/rpop) to manage and process lists effectively.

By understanding and using the `LPOP` command effectively, you can manage list data structures in DiceDB efficiently, implementing queue-like behaviors and more.