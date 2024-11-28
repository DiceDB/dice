---
title: LLEN
description: Returns the length of the list stored at a key. If the key does not exist, it is interpreted as an empty list and 0 is returned. An error is returned when the value stored at the key is not a list.
---

The `LLEN` command in DiceDB is used to obtain the length of a list stored at a specified key. This command is particularly useful for determining the number of elements in a list, which can help in various list management and processing tasks.

## Syntax

```bash
LLEN key
```

## Parameters

| Parameter | Description                                                         | Type   | Required |
| --------- | ------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key associated with the list whose length you want to retrieve. | String | Yes      |

## Return values

| Condition                                   | Return Value                                                    |
| ------------------------------------------- | --------------------------------------------------------------- |
| Command is successful                       | `Integer` denoting the length of the list at the specified key. |
| If the key does not exist                   | `0` (the key is interpreted as an empty list)                   |
| Syntax or specified constraints are invalid | error                                                           |

## Behaviour

- If the key exists and is associated with a list, the `LLEN` command returns the number of elements in the list.
- If the key does not exist, the `LLEN` command returns `0`, indicating that the list is empty.
- If the key exists but is not associated with a list, an error is returned.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not associated with a list.

2. `Wrong number of arguments`

   - Error Message: `(error) ERR wrong number of arguments for 'llen' command`
   - Occurs if command is executed without any arguments or with 2 or more arguments

## Example Usage

### Basic Usage

Getting the `LLEN` of a list `mylist` with values `["one", "two", "three"]`.

```bash
127.0.0.1:7379> RPUSH mylist "one"
(integer) 1
127.0.0.1:7379> RPUSH mylist "two"
(integer) 2
127.0.0.1:7379> RPUSH mylist "three"
(integer) 3
127.0.0.1:7379> LLEN mylist
(integer) 3
```

### Non-Existent Key

Getting the `LLEN` of a list `nonExistentList` which does not exist.

```bash
127.0.0.1:7379> LLEN nonExistentList
(integer) 0
```

### Invalid Usage: Key is Not a List

Trying to get the `LLEN` of a key `mystring` which is holding wrong data type `string`.

```bash
127.0.0.1:7379> SET mystring "Hello, World!"
OK
127.0.0.1:7379> LLEN mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Best Practices

- Check Key Type: Before using `LLEN`, ensure that the key is associated with a list to avoid errors.
- Handle Non-Existent Keys: Be prepared to handle the case where the key does not exist, as `LLEN` will return `0` in such scenarios.
- Use in Conjunction with Other List Commands: The `LLEN` command is often used alongside other list commands like [`RPUSH`](/commands/rpush), [`LPUSH`](/commands/lpush), [`LPOP`](/commands/lpop), and [`RPOP`](/commands/rpop) to manage and process lists effectively.