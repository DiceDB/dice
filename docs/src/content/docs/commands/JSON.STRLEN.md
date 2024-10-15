---
title: JSON.STRLEN
description: The 'JSON.STRLEN' command is used to get the length of a string at a given path in a JSON Document stored in DiceDB
---

The `JSON.STRLEN` command is used to determine the length of a JSON string at a specified path within a JSON document stored in DiceDB.

## Syntax

```
JSON.STRLEN <key> <path>
```

## Parameters

| Parameter | Description                                                             | Type   | Required |
|-----------|-------------------------------------------------------------------------|--------|----------|
| `key`     | The key under which the JSON document is stored.                        | String | Yes      |
| `path`    | The JSONPath to the string within the JSON document. The path must contain a string.  | String | No       |

## Return Value

| Condition                                 | Return Value                                                                    |
|-------------------------------------------|---------------------------------------------------------------------------------|
| JSON string is found at the specified path| Length of the JSON string.                                                      |
| JSONPath contains `*` wildcard            | Indicates the length of each key in the JSON. Returns `nil` for non-string keys.|


## Behaviour

When the `JSON.STRLEN` command is executed, DiceDB will:

1. If the specified `key` does not exist, the command returns `(empty list or set)`.
2. If `path` is not specified, the command throws an error.
3. `$` is considered as root path which returns `nil`.
4. If the specified path exists and points to a JSON string, the command returns the length of the string.
5. If the path does not exist or does not point to a JSON string, an error is returned.
6. Multiple results are represented as a list in case of wildcards, where each string length is returned in order of appearance and `nil` is returned if it's not a string.

## Errors

The following errors may be raised when executing the `JSON.STRLEN` command:

- `(error) ERROR wrong number of arguments for 'JSON.STRLEN' command`: This error occurs if the number of arguments provided to the command is incorrect.
- `(error)  ERROR invalid JSONPath`: This error occurs if the specified path does not exist within the JSON document.
- `(error) WRONGTYPE wrong type of path value - expected string but found {DataType at root path}`: This error occurs when a valid key is provided but no specific path value is given. By default, DiceDB assumes the root path ($) in such cases. If the data at the root path is not of type string, this error is returned, indicating a type mismatch between the expected string and the actual data type at the root.

## Example Usage

### Example 1: Basic Usage

Assume we have a JSON document stored under the key `user:1001`:

```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "address": {
    "city": "New York",
    "zipcode": "10001"
  }
}
```

To get the length of the `name` string:

```bash
JSON.STRLEN user:1001 $.name
"8"
```

### Example 2: Nested JSON String

To get the length of the `city` string within the `address` object:

```bash
JSON.STRLEN user:1001 $.address.city
"8"
```

### Example 3: Non-Existent Path

If the path does not exist:

```bash
JSON.STRLEN user:1001 $.phone
(empty list or set)
```

`Return Value`: `(empty list or set)`

### Example 4: Path is Not a String

If the path points to a non-string value:

```bash
JSON.STRLEN user:1001 $.address
(nil)
```

## Notes

- JSONPath expressions are used to navigate within the JSON document. Ensure that the path provided is valid and points to a JSON string to avoid errors.

By following this documentation, users should be able to effectively utilize the `JSON.STRLEN` command to determine the length of JSON strings stored within their DiceDB database.

