---
title: GETSET
description: Documentation for the DiceDB command GETSET
---

The `GETSET` command in DiceDB is a powerful atomic operation that combines the functionality of `GET` and `SET` commands. It retrieves the current value of a key and simultaneously sets a new value for that key. This command is particularly useful when you need to update a value and also need to know the previous value in a single atomic operation.

## Syntax

```
GETSET key value
```

## Parameters

- `key`: The key whose value you want to retrieve and set. This must be a string.
- `value`: The new value to set for the specified key. This must be a string.

## Return Value

The `GETSET` command returns the old value stored at the specified key before the new value was set. If the key did not exist, it returns `nil`.

## Behaviour

When the `GETSET` command is executed, the following sequence of actions occurs:

1. The current value of the specified key is retrieved.
1. The specified key is updated with the new value.
1. The old value is returned to the client.

This operation is atomic, meaning that no other commands can be executed on the key between the get and set operations.

## Error Handling

The `GETSET` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If the key exists but is not a string, DiceDB will return an error.
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
1. `Syntax Error`: If the command is not provided with exactly two arguments (key and value), DiceDB will return a syntax error.
   - Error Message: `(error) ERR wrong number of arguments for 'getset' command`

## Example Usage

### Basic Example

```DiceDB
SET mykey "Hello"
GETSET mykey "World"
```

`Output:`

```
"Hello"
```

`Explanation:`

- The initial value of `mykey` is set to "Hello".
- The `GETSET` command retrieves the current value "Hello" and sets the new value "World".
- The old value "Hello" is returned.

### Example with Non-Existent Key

```DiceDB
GETSET newkey "NewValue"
```

`Output:`

```
(nil)
```

`Explanation:`

- The key `newkey` does not exist.
- The `GETSET` command sets the value of `newkey` to "NewValue".
- Since the key did not exist before, `nil` is returned.

### Error Example: Wrong Type

```DiceDB
LPUSH mylist "item"
GETSET mylist "NewValue"
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

`Explanation:`

- The key `mylist` is a list, not a string.
- The `GETSET` command cannot operate on a list and returns a `WRONGTYPE` error.

### Error Example: Syntax Error

```DiceDB
GETSET mykey
```

`Output:`

```
(error) ERR wrong number of arguments for 'getset' command
```

`Explanation:`

- The `GETSET` command requires exactly two arguments: a key and a value.
- Since only one argument is provided, DiceDB returns a syntax error.

## Best Practices

- Ensure that the key you are operating on is of type string to avoid `WRONGTYPE` errors.
- Use `GETSET` when you need to update a value and also need to know the previous value in a single atomic operation.
- Handle the `nil` return value appropriately in your application logic, especially when dealing with non-existent keys.

By following this documentation, you should be able to effectively use the `GETSET` command in DiceDB to manage key-value pairs with atomic get-and-set operations.

