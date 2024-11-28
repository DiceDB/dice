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

## Return Value

| Condition                    | Return Value                                         |
| ---------------------------- | ---------------------------------------------------- |
| Command is successful        | `String` The value of the first element in the list. |
| If the key does not exist    | `nil`                                                |
| The key is of the wrong type | error                                                |

## Behavior

- When the `LPOP` command is executed, DiceDB checks if the key exists and is associated with a list.
- If the list has elements, the first element is removed and returned.
- If the key does not exist, the command treats it as an empty list and returns `nil`.
- If the key exists but is not associated with a list, a `WRONGTYPE` error is returned.
- If more than one key is passed, an error is returned.

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
(error) ERR wrong number of arguments for 'lpop' command
```

## Best Practices

- `Check Key Type`: Before using `LPOP`, ensure that the key is associated with a list to avoid errors.
- `Handle Non-Existent Keys`: Be prepared to handle the case where the key does not exist, as `LPOP` will return `nil` in such scenarios.
- `Use in Conjunction with Other List Commands`: The `LPOP` command is often used alongside other list commands like [`RPUSH`](/commands/rpush), [`LPUSH`](/commands/lpush), [`LLEN`](/commands/llen), and [`RPOP`](/commands/rpop) to manage and process lists effectively.

By understanding and using the `LPOP` command effectively, you can manage list data structures in DiceDB efficiently, implementing queue-like behaviors and more.