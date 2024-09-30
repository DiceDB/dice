---
title: RPOP
description: Documentation for the DiceDB command RPOP
---

The `RPOP` command in DiceDB is used to remove and return the last element of a list. This command is particularly useful when you need to process elements in a Last-In-First-Out (LIFO) order.

## Syntax

```
RPOP key
```

## Parameters

- `key`: The key of the list from which the last element will be removed and returned. The key must be of type list. If the key does not exist, it is treated as an empty list and the command returns `nil`.

## Return Value

- `String`: The value of the last element in the list, after removing it.
- `nil`: If the key does not exist or the list is empty.

## Behaviour

When the `RPOP` command is executed, the following steps occur:

1. DiceDB checks if the key exists and is of type list.
2. If the key does not exist, the command returns `nil`.
3. If the key exists but the list is empty, the command returns `nil`.
4. If the key exists and the list is not empty, the last element of the list is removed and returned.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the key exists but is not of type list.
- `(nil)`: This is returned if the key does not exist or the list is empty.

## Example Usage

### Example 1: Basic Usage

```shell
# Add elements to the list
LPUSH mylist "one"
LPUSH mylist "two"
LPUSH mylist "three"

# Current state of 'mylist': ["three", "two", "one"]

# Remove and return the last element
RPOP mylist
# Output: "one"

# Current state of 'mylist': ["three", "two"]
```

### Example 2: Empty List

```shell
# Create an empty list
LPUSH emptylist

# Remove and return the last element from an empty list
RPOP emptylist
# Output: (nil)
```

### Example 3: Non-List Key

```shell
# Set a string key
SET mystring "Hello"

# Attempt to RPOP from a string key
RPOP mystring
# Output: (error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `RPOP` command is atomic, meaning it is safe to use in concurrent environments.
- If you need to remove and return the first element of the list, use the `LPOP` command instead.
- For blocking behavior, consider using `BRPOP`.

## Related Commands

- `LPUSH`: Insert all the specified values at the head of the list stored at key.
- `LPOP`: Removes and returns the first element of the list stored at key.
- `BRPOP`: Removes and returns the last element of the list stored at key, or blocks until one is available.

By understanding the `RPOP` command, you can effectively manage lists in DiceDB, ensuring that you can retrieve and process elements in a LIFO order.

