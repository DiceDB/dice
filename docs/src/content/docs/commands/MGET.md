---
title: MGET
description: Documentation for the DiceDB command MGET
---

The `MGET` command in DiceDB is used to retrieve the values of multiple keys in a single call. This command is particularly useful for reducing the number of round trips between the client and the server when you need to fetch multiple values.

## Syntax

```
MGET key [key ...]
```

## Parameters

- `key [key ...]`: One or more keys for which the values need to be retrieved. Each key is a string.

## Return Value

The `MGET` command returns an array of values corresponding to the specified keys. If a key does not exist, the corresponding value in the array will be `nil`.

## Behaviour

When the `MGET` command is executed, DiceDB will:

1. Look up each specified key in the database.
2. Retrieve the value associated with each key.
3. Return an array of values in the same order as the keys were specified.

If a key does not exist, the corresponding position in the returned array will contain `nil`.

## Error Handling

The `MGET` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If any of the specified keys exist but are not of the string type, a `WRONGTYPE` error will be raised.
2. `Syntax Error`: If the command is not used with at least one key, a `syntax error` will be raised.

## Example Usage

### Basic Example

```DiceDB
SET key1 "value1"
SET key2 "value2"
SET key3 "value3"

MGET key1 key2 key3
```

`Output:`

```
1) "value1"
2) "value2"
3) "value3"
```

### Example with Non-Existent Keys

```DiceDB
SET key1 "value1"
SET key2 "value2"

MGET key1 key2 key3
```

`Output:`

```
1) "value1"
2) "value2"
3) (nil)
```

### Example with Mixed Types

```DiceDB
SET key1 "value1"
LPUSH key2 "value2"

MGET key1 key2
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
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

