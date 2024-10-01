---
title: JSON.GET
description: The `JSON.GET` command allows you to store, update, and retrieve JSON values in DiceDB. The `JSON.GET` command retrieves JSON data stored at a specified key. This command is useful for fetching JSON objects, arrays, or values from the DiceDB database.
---

The `JSON.GET` command allows you to store, update, and retrieve JSON values in DiceDB. The `JSON.GET` command retrieves JSON data stored at a specified key. This command is useful for fetching JSON objects, arrays, or values from the DiceDB database.

## Syntax

```
JSON.GET <key> [path]
```

## Parameters

- `key`: (Required) The key under which the JSON data is stored in DiceDB.
- `path`: (Optional) A JSONPath expression to specify the part of the JSON document to retrieve. If not provided, the entire JSON document is returned.

## Return Value

The command returns the JSON data stored at the specified key and path. The data is returned as a JSON string. If the specified key or path does not exist, the command returns `nil`.

## Behaviour

When the `JSON.GET` command is executed:

1. DiceDB checks if the specified key exists.
2. If the key exists, DiceDB retrieves the JSON data stored at that key.
3. If a path is provided, DiceDB extracts the specified part of the JSON document using the JSONPath expression.
4. The retrieved JSON data is returned as a JSON string.

## Error Handling

The `JSON.GET` command can raise the following errors:

- `(error) ERR wrong number of arguments for 'JSON.GET' command`: This error occurs if the command is called with an incorrect number of arguments.
- `(error) ERR key does not exist`: This error occurs if the specified key does not exist in the DiceDB database.
- `(error) ERR invalid path`: This error occurs if the provided JSONPath expression is invalid or does not match any part of the JSON document.

## Example Usage

### Example 1: Retrieve Entire JSON Document

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}'
OK
127.0.0.1:7379> JSON.GET user:1001
"{\"name\":\"John Doe\",\"age\":30,\"email\":\"john.doe@example.com\"}"
```

### Example 2: Retrieve Specific Field from JSON Document

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}'
OK
127.0.0.1:7379> JSON.GET user:1001 $.name
"\"John Doe\""
```

### Example 3: Retrieve Nested Field from JSON Document

```bash
127.0.0.1:7379> JSON.SET user:1002 $ '{"name": "Jane Doe", "address": {"city": "New York", "zip": "10001"}}'
OK
127.0.0.1:7379> JSON.GET user:1002 $.address.city
"\"New York\""
```

### Example 4: Handling Non-Existent Key

```bash
127.0.0.1:7379> JSON.GET user:9999
(nil)
```

### Example 5: Handling Invalid Path

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}'
OK
127.0.0.1:7379> JSON.GET user:1001 $.nonexistent
(nil)
```

## Notes

- JSONPath expressions allow you to navigate and retrieve specific parts of a JSON document. Ensure that your JSONPath expressions are correctly formatted to avoid errors.

By understanding the `JSON.GET` command, you can efficiently retrieve JSON data stored in your DiceDB database, enabling you to build powerful and flexible applications that leverage the capabilities of DiceDBJSON.
