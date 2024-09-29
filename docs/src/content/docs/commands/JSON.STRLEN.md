---
title: JSON.STRLEN
description: Documentation for the DiceDB command JSON.STRLEN
---

The `JSON.STRLEN` command is part of the DiceDBJSON module, which allows you to work with JSON data in DiceDB. This command is used to determine the length of a JSON string at a specified path within a JSON document stored in DiceDB.

## Syntax

```plaintext
JSON.STRLEN <key> <path>
```

## Parameters

- `key`: The key under which the JSON document is stored in DiceDB.
- `path`: The JSONPath expression that specifies the location of the JSON string within the document. The path must point to a JSON string.

## Return Value

- `Integer`: The length of the JSON string at the specified path.
- `Null`: If the path does not exist or does not point to a JSON string.

## Behaviour

When the `JSON.STRLEN` command is executed, DiceDB will:

1. Retrieve the JSON document stored at the specified key.
2. Navigate to the specified path within the JSON document.
3. If the path points to a JSON string, return the length of the string.
4. If the path does not exist or does not point to a JSON string, return `null`.

## Error Handling

The following errors may be raised when executing the `JSON.STRLEN` command:

- `(error) ERR wrong number of arguments for 'JSON.STRLEN' command`: This error occurs if the number of arguments provided to the command is incorrect.
- `(error) ERR key does not exist`: This error occurs if the specified key does not exist in the DiceDB database.
- `(error) ERR path does not exist`: This error occurs if the specified path does not exist within the JSON document.
- `(error) ERR path is not a string`: This error occurs if the specified path does not point to a JSON string.

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

```plaintext
JSON.STRLEN user:1001 $.name
```

`Return Value`: `8`

### Example 2: Nested JSON String

To get the length of the `city` string within the `address` object:

```plaintext
JSON.STRLEN user:1001 $.address.city
```

`Return Value`: `8`

### Example 3: Non-Existent Path

If the path does not exist:

```plaintext
JSON.STRLEN user:1001 $.phone
```

`Return Value`: `null`

### Example 4: Path is Not a String

If the path points to a non-string value:

```plaintext
JSON.STRLEN user:1001 $.address
```

`Return Value`: `null`

## Notes

- The `JSON.STRLEN` command is specific to the DiceDBJSON module and will not work unless the module is installed and loaded in your DiceDB instance.
- JSONPath expressions are used to navigate within the JSON document. Ensure that the path provided is valid and points to a JSON string to avoid errors.

By following this documentation, users should be able to effectively utilize the `JSON.STRLEN` command to determine the length of JSON strings stored within their DiceDB database.

