---
title: LRANGE
description: Returns the specified elements of the list stored at key.
---

The `LRANGE` command in DiceDB returns the specified elements of the list stored at key.

The offsets start and stop are zero-based indexes, with 0 being the first element of the list (the head of the list), 1 being the next element and so on.

These offsets can also be negative numbers indicating offsets starting at the end of the list.
For example, -1 is the last element of the list, -2 the penultimate, and so on.

Out of range indexes will not produce an error. If start is larger than the end of the list, an empty list is returned.
If stop is larger than the actual end of the list it will be treated like the last element of the list.

## Syntax

```bash
LRANGE key start stop
```

## Parameters

| Parameter | Description                                                         | Type   | Required |
| --------- | ------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key associated with the list whose length you want to retrieve. | String | Yes      |
| `start`   | The start offset.                                                   | String | Yes      |
| `stop`    | The stop offset.                                                    | String | Yes      |

## Return values

| Condition                                   | Return Value                                           |
| ------------------------------------------- | ------------------------------------------------------ |
| Command is successful                       | Array reply: a list of elements in the specified range |
| If the key does not exist                   | An empty array if the key doesn't exist.               |
| Syntax or specified constraints are invalid | error                                                  |

## Behaviour

- If the key exists and is associated with a list, the `LRANGE` command returns the specified elements of the list stored at key.
- If the key does not exist, the `LRANGE` command returns an empty array.
- If the key exists but is not associated with a list, an error is returned.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not associated with a list.

2. `Wrong number of arguments`:

   - Error Message: `(error) wrong number of arguments for LRANGE`
   - Occurs when invalid arguments are passed to the command.

3. `Invalid start or stop offsets`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs when the start/stop offset is not valid.

## Example Usage

### Basic Usage

Getting the `LRANGE` of a list `mylist` with values `["one", "two"]`.

```bash
127.0.0.1:7379> RPUSH mylist "one"
(integer) 1
127.0.0.1:7379> RPUSH mylist "two"
(integer) 2
127.0.0.1:7379> LRANGE mylist 0 100
1) "one"
2) "two"
127.0.0.1:7379> LRANGE mylist -1 10
1) "two"
```

### Non-Existent Key

Getting the `LRANGE` of a list `nonExistentList` which does not exist.

```bash
127.0.0.1:7379> LRANGE nonexistentlist 0 100
(empty array)
```

### Invalid start/stop offset

Trying to get the `LRANGE` on a key `mylist` with invalid stop offset.

```bash
127.0.0.1:7379> LRANGE mylist 0 10ff
(error) ERR value is not an integer or out of range
```

### Invalid usage

Trying to get the `LRANGE` on a key `myhash` which doens't hold list data type.

```bash
127.0.0.1:7379> HMSET myhash field1 "value1"
OK
127.0.0.1:7379> HGET myhash field1
"value1"
127.0.0.1:7379> LRANGE myhash 0 100
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Best Practices

- Check Key Type: Before using `LRANGE`, ensure that the key is associated with a list to avoid errors.
- Handle Non-Existent Keys: Be prepared to handle the case where the key does not exist, as `LRANGE` will return an empty array in such scenarios.
- Use in Conjunction with Other List Commands: The `LRANGE` command is often used alongside other list commands like [`RPUSH`](/commands/rpush), [`LPUSH`](/commands/lpush), [`LPOP`](/commands/lpop), and [`RPOP`](/commands/rpop) to manage and process lists effectively.