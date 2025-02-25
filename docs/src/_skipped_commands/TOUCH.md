---
title: TOUCH
description: The `TOUCH` command in DiceDB is used to update the last access time of one or more keys without modifying their values. This can be particularly useful for cache management, where you want to keep certain keys from expiring by marking them as recently used.
---

The `TOUCH` command in DiceDB is used to update the last access time of one or more keys without modifying their values. This can be particularly useful for cache management, where you want to keep certain keys from expiring by marking them as recently used.

## Syntax

```bash
TOUCH key [key ...]
```

## Parameters

| Parameter | Description                                                                                              | Type   | Required |
| --------- | -------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key(s) whose last access time you want to update. You can specify multiple keys separated by spaces. | String | Yes      |

## Return Value

| Condition                                               | Return Value |
| ------------------------------------------------------- | ------------ |
| The access time of the key(s) was successfully updated. | `integer`    |
| The type of key is unsupported.                         | `error`      |

## Behaviour

- When the `TOUCH` command is executed, it will Check whether or not the specified key(s) exists in the database.
- If a key exists, its last access time will be updated to the current time.
- If a key does not exist, it will be ignored.
- The command will return the count of keys that were successfully touched.

## Errors

The `TOUCH` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the key exists but is not of a type that supports the `TOUCH` operation (e.g., a key holding a non-string value).

## Example Usage

### Single Key

```bash
SET mykey "Hello"
TOUCH mykey
```

In this example, the `TOUCH` command updates the last access time of the key `mykey`.

### Multiple Keys

```bash
SET key1 "value1"
SET key2 "value2"
TOUCH key1 key2 key3
```

In this example, the `TOUCH` command attempts to update the last access time for `key1`, `key2`, and `key3`. Since `key3` does not exist, it will be ignored. The command will return `2`, indicating that two keys were successfully touched.

## Invalid usage

Trying to touch key `mylist` will result in a `WRONGTYPE` error because `mylist` is a list, not a string.

```bash
LPUSH mylist "element"
TOUCH mylist
```

## Best Practices

- Avoid using the `TOUCH` command on a large number of keys simultaneously, as it may slow down the server.

## Notes

- The `TOUCH` command only updates the last access time of a key, without modifying its value or other attributes.
