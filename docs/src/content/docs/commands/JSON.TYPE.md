---
title: JSON.TYPE
description: Documentation for the DiceDB command JSON.TYPE
---

The `JSON.TYPE` command is part of the DiceDBJSON module, which allows you to work with JSON data structures in DiceDB. This command is used to determine the type of value stored at a specified path within a JSON document.

## Parameters

### Key

- `Type:` String
- `Description:` The key under which the JSON document is stored.
- `Example:` `user:1001`

### Path

- `Type:` String
- `Description:` The JSONPath expression that specifies the location within the JSON document. The default path is the root (`$`).
- `Example:` `$.address.city`

## Return Value

### Type

- `String`
- `Description:` The type of the value at the specified path. Possible return values include:
  - `null`
  - `boolean`
  - `number`
  - `string`
  - `array`
  - `object`
- `Example:` `string`

## Behaviour

When the `JSON.TYPE` command is executed, DiceDB will:

1. Retrieve the JSON document stored at the specified key.
2. Navigate to the specified path within the JSON document.
3. Determine the type of the value located at that path.
4. Return the type as a string.

If the path does not exist within the JSON document, the command will return `null`.

## Error Handling

### Key Does Not Exist

- `Error:` `(nil)`
- `Description:` If the specified key does not exist in the DiceDB database, the command will return `nil`.

### Invalid JSONPath

- `Error:` `ERR wrong number of arguments for 'json.type' command`
- `Description:` If the provided JSONPath is invalid or malformed, DiceDB will return an error indicating that the number of arguments is incorrect.

### Non-JSON Data

- `Error:` `WRONGTYPE Operation against a key holding the wrong kind of value`
- `Description:` If the key exists but does not hold a JSON document, DiceDB will return an error indicating that the operation is against a key holding the wrong kind of value.

## Example Usage

### Example 1: Basic Usage

```shell
127.0.0.1:6379> JSON.SET user:1001 $ '{"name": "John", "age": 30, "address": {"city": "New York", "zip": "10001"}}'
OK
127.0.0.1:6379> JSON.TYPE user:1001 $.name
"string"
127.0.0.1:6379> JSON.TYPE user:1001 $.age
"number"
127.0.0.1:6379> JSON.TYPE user:1001 $.address
"object"
127.0.0.1:6379> JSON.TYPE user:1001 $.address.city
"string"
```

### Example 2: Non-Existent Path

```shell
127.0.0.1:6379> JSON.TYPE user:1001 $.nonexistent
(null)
```

### Example 3: Key Does Not Exist

```shell
127.0.0.1:6379> JSON.TYPE user:9999 $.name
(nil)
```

### Example 4: Invalid JSONPath

```shell
127.0.0.1:6379> JSON.TYPE user:1001 $..name
(error) ERR wrong number of arguments for 'json.type' command
```

### Example 5: Non-JSON Data

```shell
127.0.0.1:6379> SET mykey "This is a string"
OK
127.0.0.1:6379> JSON.TYPE mykey $
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```
