---
title: LPOP
description: Documentation for the DiceDB command LPOP
---

The `LPOP` command is used to remove and return the first element of a list stored at a specified key in DiceDB. This command is useful for implementing queue-like structures where elements are processed in a First-In-First-Out (FIFO) order.

## Syntax

```
LPOP key
```

## Parameters

- `key`: The key of the list from which the first element will be removed and returned. The key must be a valid string.

## Return Value

- `String`: The value of the first element in the list, if the list exists and is not empty.
- `nil`: If the key does not exist or the list is empty.

## Behaviour

When the `LPOP` command is executed:

1. If the key exists and is associated with a list, the first element of the list is removed and returned.
2. If the key does not exist, the command returns `nil`.
3. If the key exists but is not associated with a list, an error is returned.
4. If the list is empty, the command returns `nil`.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the key exists but is not associated with a list. For example, if the key is associated with a string, set, or hash, this error will be returned.

## Example Usage

### Example 1: Basic Usage

```DiceDB
RPUSH mylist "one" "two" "three"
LPOP mylist
```

`Output:`

```
"one"
```

`Explanation:`

- The list `mylist` initially contains the elements \["one", "two", "three"\].
- The `LPOP` command removes and returns the first element, which is "one".
- The list `mylist` now contains \["two", "three"\].

### Example 2: List is Empty

```DiceDB
RPUSH mylist "one"
LPOP mylist
LPOP mylist
```

`Output:`

```
"one"
(nil)
```

`Explanation:`

- The list `mylist` initially contains the element \["one"\].
- The first `LPOP` command removes and returns "one".
- The second `LPOP` command returns `nil` because the list is now empty.

### Example 3: Key Does Not Exist

```DiceDB
LPOP nonexistinglist
```

`Output:`

```
(nil)
```

`Explanation:`

- The key `nonexistinglist` does not exist in the database.
- The `LPOP` command returns `nil`.

### Example 4: Key is Not a List

```DiceDB
SET mystring "hello"
LPOP mystring
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

`Explanation:`

- The key `mystring` is associated with a string value "hello".
- The `LPOP` command returns an error because `mystring` is not a list.

## Best Practices

- Ensure that the key you are using with `LPOP` is associated with a list to avoid errors.
- Use `EXISTS` command to check if a key exists before using `LPOP` to handle cases where the key might not exist.
- Consider using `LLEN` to check the length of the list if you need to handle empty lists differently.

## Related Commands

- `RPUSH`: Append one or multiple elements to the end of a list.
- `LPUSH`: Prepend one or multiple elements to the beginning of a list.
- `RPOP`: Remove and return the last element of a list.
- `LRANGE`: Get a range of elements from a list.

By understanding and using the `LPOP` command effectively, you can manage list data structures in DiceDB efficiently, implementing queue-like behaviors and more.

