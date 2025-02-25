---
title: MGET
description: Documentation for the DiceDB command MGET
---

The `MGET` command in DiceDB is used to retrieve the values of multiple keys in a single call. This command is particularly useful for reducing the number of round trips between the client and the server when you need to fetch multiple values.

## Syntax

```bash
MGET key [key ...]
```

## Parameters

| Parameter | Description                                                                       | Type   | Required |
| --------- | --------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | One or more keys for which the values need to be retrieved. Each key is a string. | String | Yes      |

## Return Value

The `MGET` command returns an array of values corresponding to the specified keys. If a key does not exist, the corresponding value in the array will be `nil`.

| Condition                  | Return Value                                               |
| -------------------------- | ---------------------------------------------------------- |
| All specified keys exist   | An array of values corresponding to the specified keys     |
| Some keys do not exist     | An array where non-existent keys have `nil` as their value |
| Keys are of the wrong type | error                                                      |

## Behaviour

- The `MGET` command retrieves the values of multiple specified keys in the same order as they are listed in the command.
- If a key does not exist, its corresponding position in the returned array will contain `nil`.
- If any of the keys hold non-string values (e.g., lists or sets), the `MGET` command will return an error indicating that the operation was attempted on the wrong type of key.

## Errors

The `MGET` command can raise errors in the following scenarios:

1. `Wrong type of value or key`:

   - `Error Message`: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when one of the keys holds a value that is not a string.

2. `Invalid syntax`:
   - `Error Message`: `(error) ERR wrong number of arguments for 'mget' command`
   - Occurs when the command is issued without any keys.

## Example Usage

### Basic Example

Retrieving the values of multiple keys `key1`, `key2`, and `key3`:

```bash
127.0.0.1:7379> SET key1 "value1"
OK
```

```bash
127.0.0.1:7379> SET key2 "value2"
OK
```

```bash
127.0.0.1:7379> SET key3 "value3"
OK
```

```bash
127.0.0.1:7379> MGET key1 key2 key3
1) "value1"
2) "value2"
3) "value3"
```

### Example with Non-Existent Keys

```bash
127.0.0.1:7379> SET key1 "value1"
OK
```

```bash
127.0.0.1:7379> SET key2 "value2"
OK
```

```bash
127.0.0.1:7379> MGET key1 key2 key3
1) "value1"
2) "value2"
3) (nil)
```

### Key doesn't exist previously

```bash
127.0.0.1:7379> SET key1 "value1"
OK
```

```bash
127.0.0.1:7379> LPUSH key2 "value2"
(integer) 1
```

```bash
127.0.0.1:7379> MGET key1 key2
1) "value1"
2) (nil)
```

### Key exists previously with different datatype

```bash
127.0.0.1:7379> SET key1 "value1"
OK
```

```bash
127.0.0.1:7379> SET key2 "value2"
OK
```

```bash
127.0.0.1:7379> LPUSH key2 "value3"
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

```bash
127.0.0.1:7379> MGET key1 key2
1) "value1"
2) "value2"
```

## Notes

- The `MGET` command is atomic, meaning that it will either retrieve all the specified keys or none if an error occurs.
- The order of the returned values matches the order of the specified keys.
- Using `MGET` can be more efficient than multiple `GET` commands, especially when dealing with a large number of keys.

## Best Practices

- Ensure that all specified keys are of the string type to avoid `WRONGTYPE` errors.
- Use `MGET` to minimize the number of round trips to the DiceDB server when you need to fetch multiple values.
- Handle `nil` values in the returned array appropriately in your application logic to account for non-existent keys.

By following this documentation, you should be able to effectively use the `MGET` command in DiceDB to retrieve multiple values efficiently and handle any potential errors that may arise.
