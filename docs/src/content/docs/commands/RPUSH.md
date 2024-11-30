---
title: RPUSH
description: Documentation for the DiceDB command RPUSH
---

The `RPUSH` command is used in DiceDB to insert one or multiple values at the tail (right end) of a list. If the list does not exist, it is created as an empty list before performing the push operations. This command is useful for maintaining a list of elements where new elements are added to the end of the list.

## Syntax

```bash
RPUSH key value [value ...]
```

## Parameters

| Parameter          | Description                                                                               | Type   | Required |
| ------------------ | ----------------------------------------------------------------------------------------- | ------ | -------- |
| `key`              | The name of the list where values are inserted. If it does not exist, it will be created. | String | Yes      |
| `value [value...]` | One or more space separated values to be inserted at the tail of the list.                | String | Yes      |

## Return Value

| Condition                                   | Return Value                                   |
| ------------------------------------------- | ---------------------------------------------- |
| Command is successful                       | `Integer` - length of the list after execution |
| Syntax or specified constraints are invalid | error                                          |

## Behaviour

- When the `RPUSH` command is executed, the specified values are inserted at the tail of the list identified by the given key.
- If the key does not exist, a new list is created and the values are inserted.
- If the key exists but is not a list, an error is returned.
- The `RPUSH` operation is atomic. If multiple clients issue `RPUSH` commands concurrently, DiceDB ensures that the list remains consistent and the values are appended in the order the commands were received.

## Errors

1. `Wrong Number of Arguments`

   - Error Message: `(error) ERR wrong number of arguments for 'rpush' command`
   - Occurs if the key parameters is not provided or at least one value is not provided.

2. `Wrong Type of Key or Value`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists and is not a list. DiceDB expects the key to either be non-existent or to hold a list data type.

## Example Usage

### Basic Usage

Inserting a value `hello` into `mylist`

```bash
127.0.0.1:7379> RPUSH mylist "hello"
(integer) 1
```

### Inserting Multiple Values

Inserting multiple values `world`, `foo`, `bar` into `mylist`

```bash
127.0.0.1:7379> RPUSH mylist "world" "foo" "bar"
(integer) 3
```

### Invalid Usage: Key is Not a List

Insert the value `value` at the tail of the key `mystring`, which stores a string, not a list.

```shell
127.0.0.1:7379> SET mystring "not a list"
OK
127.0.0.1:7379> RPUSH mystring "value"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Invalid Usage: Wrong Number of Arguments

Calling `RPUSH` without passing any values:

```bash
RPUSH mylist
(error) ERR wrong number of arguments for 'rpush' command
```

## Best Practices

- `Check Key Type`: Before using `RPUSH`, ensure that the key is associated with a list to avoid errors.
- `Use in Conjunction with Other List Commands`: The `RPUSH` command is often used alongside other list commands like [`LLEN`](/commands/llen), [`LPUSH`](/commands/lpush), [`LPOP`](/commands/lpop), and [`RPOP`](/commands/rpop) to manage and process lists effectively.