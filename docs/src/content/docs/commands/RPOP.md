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

| Parameter | Description                                                          | Type   | Required |
|-----------|----------------------------------------------------------------------|--------|----------|
| `key`     | The key of the list from which the last element will be removed.      | String | Yes      |


## Return values

| Condition                                      | Return Value                                   |
|------------------------------------------------|------------------------------------------------|
| The command is successful                      | The value of the last element in the list      |
| The list is empty or the key does not exist    | `nil`                                          |
| The key is of the wrong type                   | Error: `WRONGTYPE Operation against a key holding the wrong kind of value` |


## Behaviour

- The `RPOP` command checks if the key exists and whether it contains a list. 
- If the key does not exist, the command treats it as an empty list and returns `nil`.
- If the key exists but the list is empty, `nil` is returned.
- If the list has elements, the last element is removed and returned.
- If the key exists but is not of type list, an error is raised.

## Errors

1. **Wrong type of value or key**:
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to run `RPOP` on a key that is not a list.

2. **Non-existent or empty list**:
   - Returns `nil` when the key does not exist or the list is empty.

## Example Usage

### Example 1: Basic Usage

```bash
127.0.0.1:7379> LPUSH mylist "one" "two" "three"
(integer) 3
127.0.0.1:7379> RPOP mylist
"one"
```

### Example 2: Empty List

```bash
127.0.0.1:7379> RPOP emptylist
(nil)
```

### Example 3: Non-List Key

```bash
127.0.0.1:7379> SET mystring "Hello"
OK
127.0.0.1:7379> RPOP mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `RPOP` command is atomic, meaning it is safe to use in concurrent environments.
- If you need to remove and return the first element of the list, use the `LPOP` command instead.

## Related Commands

- `LPUSH`: Insert all the specified values at the head of the list stored at key.
- `LPOP`: Removes and returns the first element of the list stored at key.

By understanding the `RPOP` command, you can effectively manage lists in DiceDB, ensuring that you can retrieve and process elements in a LIFO order.

