---
title: LINSERT
description: Inserts element in the list stored at key either before or after the reference value pivot. When key does not exist, it is considered an empty list and no operation is performed. An error is returned when key exists but does not hold a list value.
---

The `LINSERT` command in DiceDB inserts an element in the list stored at key either before or after the reference value pivot. When key does not exist, it is considered an empty list and no operation is performed. An error is returned when key exists but does not hold a list value.

## Syntax

```bash
LINSERT key <BEFORE | AFTER> pivot element
```

## Parameters

| Parameter          | Description                                                         | Type   | Required |
| ------------------ | ------------------------------------------------------------------- | ------ | -------- |
| `key`              | The key associated with the list whose length you want to retrieve. | String | Yes      |
| `<before / after>` | Tells whether to insert the element before or after the pivot.      | String | Yes      |
| `pivot`            | The pivot element.                                                  | String | Yes      |
| `element`          | The element to be inserted.                                         | String | Yes      |

## Return values

| Condition                                   | Return Value                                                   |
| ------------------------------------------- | -------------------------------------------------------------- |
| Command is successful                       | Integer reply: Number of elements in the list after insertion. |
| If the key does not exist                   | 0                                                              |
| Syntax or specified constraints are invalid | error                                                          |

## Behaviour

- If the key exists and is associated with a list, the `LINSERT` command returns the list length after insertion.
- If the key does not exist, the `LINSERT` command returns 0.
- If the key exists but is not associated with a list, an error is returned.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key exists but is not associated with a list.

2. `Wrong number of arguments`:

   - Error Message: `(error) wrong number of arguments for LINSERT`
   - Occurs when invalid arguments are passed to the command.

## Example Usage

### Basic Usage

Performing the `LINSERT` on a list `mylist` with values `["one", "two"]`.

```bash
127.0.0.1:7379> RPUSH mylist "one"
(integer) 1
127.0.0.1:7379> RPUSH mylist "two"
(integer) 2
127.0.0.1:7379> LINSERT mylist after "two" "four"
(integer) 3
127.0.0.1:7379> LRANGE mylist 0 100
1) "one"
2) "two"
3) "four"
127.0.0.1:7379> LINSERT mylist before "four" "three"
(integer) 4
127.0.0.1:7379> LRANGE mylist 0 100
1) "one"
2) "two"
3) "three"
4) "four"
```

### Non-Existent Key

Getting the `LINSERT` on a list `nonExistentList` which does not exist.

```bash
127.0.0.1:7379> LINSERT nonExistentList before "two" "one"
(integer) 0
```

### Invalid usage

Trying to perform `LINSERT` on a key `myhash` which doens't hold list data type.

```bash
127.0.0.1:7379> HMSET myhash field1 "value1"
OK
127.0.0.1:7379> HGET myhash field1
"value1"
127.0.0.1:7379> LINSERT myhash before pivot element
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Best Practices

- Check Key Type: Before using `LINSERT`, ensure that the key is associated with a list to avoid errors.
- Handle Non-Existent Keys: Be prepared to handle the case where the key does not exist, as `LINSERT` will return `0` in such scenarios.
- Use in Conjunction with Other List Commands: The `LINSERT` command is often used alongside other list commands like [`RPUSH`](/commands/rpush), [`LPUSH`](/commands/lpush), [`LPOP`](/commands/lpop), and [`RPOP`](/commands/rpop) to manage and process lists effectively.