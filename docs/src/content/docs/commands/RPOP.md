---
title: RPOP
description: The `RPOP` command in DiceDB removes and returns elements from the end of a list. When provided with the optional `count` argument, the reply will consist of up to `count` elements, depending on the list's length. It is commonly used for processing elements in Last-In-First-Out (LIFO) order.
---

The `RPOP` command in DiceDB is used to remove and return elements from the end of a list. By default, it removes and returns a single element. When provided with the optional `count` argument, the reply will consist of up to `count` elements, depending on the list's length. This command is useful for processing elements in a Last-In-First-Out (LIFO) order.

## Syntax

```bash
RPOP key [count]
```

## Parameters

| Parameter | Description                                                      | Type   | Required |
| --------- | ---------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the list from which the last element will be removed. | String | Yes      |
| `count`   | The count of elemments to be removed.                            | Number | No       |

## Return values

| Condition                    | Return Value                                       |
| ---------------------------- | -------------------------------------------------- |
| The command is successful    | `String` The value of the last element in the list |
| The key does not exist       | `nil`                                              |
| The key is of the wrong type | error                                              |
| The count is invalid         | error                                              |


#### Behavior:

1. **Key Existence and Type**:
   - If the key does not exist, the command treats it as an empty list and returns `nil`.
   - If the key exists but is not associated with a list, a `WRONGTYPE` error is returned.
   - If more than one key is provided, an error is returned.

2. **Without `count` Parameter**:
   - If the list has elements and the optional `count` parameter is not provided, the last element of the list is removed and returned.

3. **With `count` Parameter**:
   - If the `count` parameter is provided:
     - If the list has elements, up to `count` elements are removed from the end of the list and returned in order (from the last element to the earliest removed).
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
127.0.0.1:7379> RPOP mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

2. `Wrong number of arguments`

   - Error Message: `(error) ERR wrong number of arguments for 'rpop' command`
   - Occurs if command is executed without any arguments or with 2 or more arguments

 3. `Invalid Count`

   - Error Message: `(error) value is out of range, must be positive`
   - Cause: Occurs if the count parameter is less than 0.    

## Example Usage

### Basic Usage

```bash
LPUSH mylist "one" "two" "three"
(integer) 3
RPOP mylist
"one"
```

### Non-Existent Key

Returns `(nil)` if the provided key is non-existent

```bash
RPOP emptylist
(nil)
```

### Invalid Usage: Key is Not a List

Trying to `LPOP` a key `mystring` which is holding wrong data type `string`.

```bash
SET mystring "hello"
OK
RPOP mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid Usage: Wrong Number of Arguments

Passing more than one key will result in an error:

```bash
RPOP mylist secondlist
(error) ERR wrong number of arguments for 'rpop' command
```

### Invalid Count Parameter

Passing a negative value for count will result in an error:

```bash
RPOP mylist -1
(error) value is out of range, must be positive
```

Passing a non integer value as the count paramater will result in an error:

```bash
RPOP mylist secondlist
(error) value is out of range, must be positive
```


## Best Practices

- `Check Key Type`: Before using `RPOP`, ensure that the key is associated with a list to avoid errors.
- `Handle Non-Existent Keys`: Be prepared to handle the case where the key does not exist, as `RPOP` will return `nil` in such scenarios.
- `Use in Conjunction with Other List Commands`: The `RPOP` command is often used alongside other list commands like [`RPUSH`](/commands/rpush), [`LPUSH`](/commands/lpush), [`LLEN`](/commands/llen), and [`LPOP`](/commands/lpop) to manage and process lists effectively.

By understanding the `RPOP` command, you can effectively manage lists in DiceDB, ensuring that you can retrieve and process elements in a LIFO order.
