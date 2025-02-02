---
title: JSON.STRLEN
description: The 'JSON.STRLEN' command is used to get the length of a string at a given path in a JSON Document stored in DiceDB
---

The `JSON.STRLEN` command is used to determine the length of a JSON string at a specified path within a JSON document stored in DiceDB.

## Syntax

```bash
JSON.STRLEN <key> <path>
```

## Parameters

| Parameter | Description                                                                          | Type   | Required |
| --------- | ------------------------------------------------------------------------------------ | ------ | -------- |
| `key`     | The key under which the JSON document is stored.                                     | String | Yes      |
| `path`    | The JSONPath to the string within the JSON document. The path must contain a string. | String | No       |

## Return Value

| Condition                                  | Return Value                                                                     |
| ------------------------------------------ | -------------------------------------------------------------------------------- |
| JSON string is found at the specified path | Length of the JSON string.                                                       |
| JSONPath contains `*` wildcard             | Indicates the length of each key in the JSON. Returns `nil` for non-string keys. |

## Behaviour

When the `JSON.STRLEN` command is executed, DiceDB will:

1. If the specified `key` does not exist, the command returns `(nil)`.
2. If `path` is not specified, the command takes `$` as root path and if the data at root path is string, returns an integer that represents the length of the string and if the data at root is not a string, returns an error indicating a type mismatch.
3. `$` is considered as root path which returns the length of the string if the data at root path is of type string or returns `nil` if the data at root is not of type string.
4. If the specified path exists and points to a JSON string, the command returns the length of the string.
5. Multiple results are represented as a list in case of wildcards(\*), where each string length is returned in order of appearance and `nil` is returned if it's not a string.

## Errors

The following errors may be raised when executing the `JSON.STRLEN` command:

1. `Incorrect Number of Arguments`:
   - `(error) ERROR wrong number of arguments for 'JSON.STRLEN' command`
   - This error occurs if the number of arguments provided to the command is incorrect.
2. `Invalid JSONPath`:
   - `(error)  ERROR invalid JSONPath`
   - This error occurs if the specified path does not exist within the JSON document.
3. `Wrong Type Error`:
   - `(error) WRONGTYPE wrong type of path value - expected string but found {DataType at root path}`
   - This error occurs when a valid key is provided but no specific path value is given. By default, DiceDB assumes the root path ($) in such cases. If the data at the root path is not of type string, this error is returned, indicating a type mismatch between the expected string and the actual data type at the root.

## Example Usage

### Basic Usage

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
(integer) 8
```

### Nested JSON String

To get the length of the `city` string within the `address` object:

```bash
JSON.STRLEN user:1001 $.address.city
(integer) 8
```

### Non-Existent Path

If the path does not exist:

```bash
JSON.STRLEN user:1001 $.phone
(empty array)
```

### Path is Not a String

If the path points to a non-string value:

```bash
JSON.STRLEN user:1001 $.address
(nil)
```

## Notes

- JSONPath expressions are used to navigate within the JSON document. Ensure that the path provided is valid and points to a JSON string to avoid errors.

By following this documentation, users should be able to effectively utilize the `JSON.STRLEN` command to determine the length of JSON strings stored within their DiceDB database.
