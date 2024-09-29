---
title: JSON.NUMINCRBY
description: Documentation for the DiceDB command JSON.NUMINCRBY
---

The `JSON.NUMINCRBY` command is part of the DiceDBJSON module, which allows for manipulation of JSON data stored in DiceDB. This command increments a numeric value stored at a specified path within a JSON document by a given amount. It is particularly useful for atomic operations on numeric fields within JSON objects.

## Syntax

```plaintext
JSON.NUMINCRBY <key> <path> <increment>
```

## Parameters

- `key`: The key under which the JSON document is stored. This is a string.
- `path`: The JSONPath expression that specifies the location of the numeric value to be incremented. This is a string.
- `increment`: The numeric value by which to increment the target value. This can be an integer or a floating-point number.

## Return Value

The command returns the new value after the increment operation. The return type is a number, which can be either an integer or a floating-point number, depending on the type of the increment and the original value.

## Behaviour

When the `JSON.NUMINCRBY` command is executed, the following steps occur:

1. The command locates the JSON document stored at the specified key.
2. It navigates to the specified path within the JSON document.
3. It retrieves the current numeric value at that path.
4. It increments the current value by the specified increment.
5. It updates the JSON document with the new value.
6. It returns the new value.

If the path does not exist or does not contain a numeric value, an error will be raised.

## Error Handling

The `JSON.NUMINCRBY` command can raise the following errors:

- `ERR key does not exist`: This error is raised if the specified key does not exist in the DiceDB database.
- `ERR path does not exist`: This error is raised if the specified path does not exist within the JSON document.
- `ERR path is not a number`: This error is raised if the value at the specified path is not a numeric value.
- `ERR invalid JSONPath`: This error is raised if the provided JSONPath expression is not valid.

## Example Usage

### Example 1: Incrementing an Integer Value

Assume we have a JSON document stored under the key `user:1001`:

```json
{
  "name": "John Doe",
  "age": 30,
  "balance": 100.50
}
```

To increment the `age` by 1:

```plaintext
JSON.NUMINCRBY user:1001 $.age 1
```

`Return Value:`

```plaintext
31
```

### Example 2: Incrementing a Floating-Point Value

To increment the `balance` by 25.75:

```plaintext
JSON.NUMINCRBY user:1001 $.balance 25.75
```

`Return Value:`

```plaintext
126.25
```

### Example 3: Handling Non-Existent Path

If we try to increment a non-existent path:

```plaintext
JSON.NUMINCRBY user:1001 $.nonexistent 10
```

`Error:`

```plaintext
ERR path does not exist
```

### Example 4: Handling Non-Numeric Value

If we try to increment a non-numeric value:

```plaintext
JSON.NUMINCRBY user:1001 $.name 5
```

`Error:`

```plaintext
ERR path is not a number
```
