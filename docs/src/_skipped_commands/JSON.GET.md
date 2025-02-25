---
title: JSON.GET
description: The `JSON.GET` command allows you to store, update, and retrieve JSON values in DiceDB. The `JSON.GET` command retrieves JSON data stored at a specified key. This command is useful for fetching JSON objects, arrays, or values from the DiceDB database.
---

The `JSON.GET` command allows you to store, update, and retrieve JSON values in DiceDB. The `JSON.GET` command retrieves JSON data stored at a specified key. This command is useful for fetching JSON objects, arrays, or values from the DiceDB database.

## Syntax

```bash
JSON.GET <key> [path]
```

## Parameters

| Parameter | Description                                                                                                                                                      | Type   | Required |
| --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key against which the JSON data is stored in DiceDB                                                                                                          | String | Yes      |
| `path`    | A JSONPath expression to specify the part of the JSON document to retrieve. If not provided, the entire JSON document is returned. Default value is **$** (root) | String | No       |

## Return Values

| Condition                                                                    | Return Value                                         |
| ---------------------------------------------------------------------------- | ---------------------------------------------------- |
| The specified key does not exists                                            | `nil`                                                |
| The specified key exists and path argument is not specified                  | `String`: The entire JSON data for the key           |
| The specified key exists and the specified path exists in the JSON data      | `String`: The data for the key at the specified path |
| The specified key exists and specified path does not exists in the JSON data | `error`                                              |
| Syntax or specified constraints are invalid                                  | error                                                |

## Behaviour

When the `JSON.GET` command is executed:

1. DiceDB checks if the specified key exists.
2. If the key exists, DiceDB retrieves the JSON data stored at that key.
3. If a path is provided, DiceDB extracts the specified part of the JSON document using the JSONPath expression.
4. The retrieved JSON data is returned as a JSON string.

## Errors

1. `Incorrect number of arguments`
   - Error Message: `(error) ERR wrong number of arguments for 'json.get' command`
2. `Invalid JSONPath expression`
   - Error Message: `(error) ERR invalid JSONPath`
3. `Non-Existent JSONPath in the JSON data stored against a key`
   - Error Message: `(error) ERR Path '$.<path>' does not exist`

## Example Usage

### Retrieve Entire JSON Document

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}'
OK
127.0.0.1:7379> JSON.GET user:1001
"{\"name\":\"John Doe\",\"age\":30,\"email\":\"john.doe@example.com\"}"
```

### Retrieve Specific Field from JSON Document

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}'
OK
127.0.0.1:7379> JSON.GET user:1001 $.name
"\"John Doe\""
```

### Retrieve Nested Field from JSON Document

```bash
127.0.0.1:7379> JSON.SET user:1002 $ '{"name": "Jane Doe", "address": {"city": "New York", "zip": "10001"}}'
OK
127.0.0.1:7379> JSON.GET user:1002 $.address.city
"\"New York\""
```

### Handling Non-Existent Key

```bash
127.0.0.1:7379> JSON.GET user:9999
(nil)
```

### Handling Non-Existent Path

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}'
OK
127.0.0.1:7379> JSON.GET user:1001 $.nonexistent
(error) ERR Path '$.nonexistent' does not exist
```

## Notes

- JSONPath expressions allow you to navigate and retrieve specific parts of a JSON document. Ensure that your JSONPath expressions are correctly formatted to avoid errors.

By understanding the `JSON.GET` command, you can efficiently retrieve JSON data stored in your DiceDB database, enabling you to build powerful and flexible applications that leverage the capabilities of DiceDB.
