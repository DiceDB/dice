---
title: RPOP
description: The `RPOP` command in DiceDB removes and returns the last element of a list. It is commonly used for processing elements in Last-In-First-Out (LIFO) order.
---

The `RPOP` command in DiceDB is used to remove and return the last element of a list. This command is useful when processing elements in a Last-In-First-Out (LIFO) order.

## Syntax

```bash
RPOP key
```

## Parameters

| Parameter | Description                                                      | Type   | Required |
| --------- | ---------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the list from which the last element will be removed. | String | Yes      |

## Return values

| Condition                    | Return Value                                       |
| ---------------------------- | -------------------------------------------------- |
| The command is successful    | `String` The value of the last element in the list |
| The key does not exist       | `nil`                                              |
| The key is of the wrong type | error                                              |

## Behaviour

- When the `RPOP` command is executed, DiceDB checks if the key exists and is associated with a list.
- If the list has elements, the last element is removed and returned.
- If the key does not exist, the command treats it as an empty list and returns `nil`.
- If the key exists but is not associated with a list, a `WRONGTYPE` error is returned.
- If more than one key is passed, an error is returned.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not associated with a list.

2. `Wrong number of arguments`

   - Error Message: `(error) ERR wrong number of arguments for 'lpop' command`
   - Occurs if command is executed without any arguments or with 2 or more arguments

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> LPUSH mylist "one" "two" "three"
(integer) 3
127.0.0.1:7379> RPOP mylist
"one"
```

### Non-Existent Key

Returns `(nil)` if the provided key is non-existent

```bash
127.0.0.1:7379> RPOP emptylist
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
RPOP mylist secondlist
(error) ERR wrong number of arguments for 'lpop' command
```

## Best Practices

- `Check Key Type`: Before using `RPOP`, ensure that the key is associated with a list to avoid errors.
- `Handle Non-Existent Keys`: Be prepared to handle the case where the key does not exist, as `RPOP` will return `nil` in such scenarios.
- `Use in Conjunction with Other List Commands`: The `RPOP` command is often used alongside other list commands like [`RPUSH`](/commands/rpush), [`LPUSH`](/commands/lpush), [`LLEN`](/commands/llen), and [`LPOP`](/commands/lpop) to manage and process lists effectively.

By understanding the `RPOP` command, you can effectively manage lists in DiceDB, ensuring that you can retrieve and process elements in a LIFO order.