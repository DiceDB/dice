---
title: LLEN
description: Documentation for the DiceDB command LLEN
---

The `LLEN` command in DiceDB is used to obtain the length of a list stored at a specified key. This command is particularly useful for determining the number of elements in a list, which can help in various list management and processing tasks.

## Syntax

```
LLEN key
```

## Parameters

- `key`: The key associated with the list whose length you want to retrieve. The key must be a valid string.

## Return Value

- `Integer`: The length of the list at the specified key. If the key does not exist, it is interpreted as an empty list and `0` is returned.

## Behaviour

When the `LLEN` command is executed, DiceDB checks the specified key:

1. If the key exists and is associated with a list, the command returns the number of elements in the list.
1. If the key does not exist, the command returns `0`, indicating that the list is empty.
1. If the key exists but is not associated with a list, an error is returned.

## Error Handling

The `LLEN` command can raise errors in the following scenarios:

1. `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the key exists but is not associated with a list. For example, if the key is associated with a string, set, hash, or any other data type, DiceDB will return an error.

## Example Usage

### Example 1: Basic Usage

```DiceDB
> RPUSH mylist "one"
(integer) 1
> RPUSH mylist "two"
(integer) 2
> RPUSH mylist "three"
(integer) 3
> LLEN mylist
(integer) 3
```

In this example, we first create a list `mylist` and add three elements to it. The `LLEN` command then returns `3`, indicating that the list contains three elements.

### Example 2: Non-Existent Key

```DiceDB
> LLEN nonExistentList
(integer) 0
```

Here, the key `nonExistentList` does not exist. The `LLEN` command returns `0`, indicating that the list is empty.

### Example 3: Key with Wrong Data Type

```DiceDB
> SET mystring "Hello, World!"
OK
> LLEN mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, the key `mystring` is associated with a string, not a list. When the `LLEN` command is executed, DiceDB returns an error indicating that the operation is against a key holding the wrong kind of value.

## Best Practices

- `Check Key Type`: Before using `LLEN`, ensure that the key is associated with a list to avoid errors.
- `Handle Non-Existent Keys`: Be prepared to handle the case where the key does not exist, as `LLEN` will return `0` in such scenarios.
- `Use in Conjunction with Other List Commands`: The `LLEN` command is often used alongside other list commands like `RPUSH`, `LPUSH`, `LPOP`, and `RPOP` to manage and process lists effectively.
