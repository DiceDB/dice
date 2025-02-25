---
title: JSON.TYPE
description: Documentation for the DiceDB command JSON.TYPE
---

The `JSON.TYPE` command allows you to work with JSON data structures in DiceDB. This command is used to determine the type of value stored at a specified path within a JSON document.

## Syntax

```bash
JSON.TYPE <key> <path>
```

## Parameters

| Key  | Description                                                                                          | Example          | Required |
| ---- | ---------------------------------------------------------------------------------------------------- | ---------------- | -------- |
| key  | The key under which the JSON document is stored.                                                     | `user:1001`      | yes      |
| path | The JSONPath expression that specifies the location within the JSON. The default path is the root($) | `$.address.city` | yes      |

## Return Value

| Condition                    | Value     |
| ---------------------------- | --------- |
| Value is null                | `null`    |
| Value is a boolean           | `boolean` |
| Value is an integer or float | `number`  |
| Value is a string            | `string`  |
| Value is an array            | `array`   |
| Value is an object           | `object`  |

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

### Basic Usage

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John", "age": 30, "address": {"city": "New York", "zip": "10001"}}'
OK
127.0.0.1:7379> JSON.TYPE user:1001 $.name
"string"
127.0.0.1:7379> JSON.TYPE user:1001 $.age
"number"
127.0.0.1:7379> JSON.TYPE user:1001 $.address
"object"
127.0.0.1:7379> JSON.TYPE user:1001 $.address.city
"string"
```

### Non-Existent Path

```bash
127.0.0.1:7379> JSON.TYPE user:1001 $.nonexistent
(empty array)
```

### Key Does Not Exist

```bash
127.0.0.1:7379> JSON.TYPE user:9999 $.name
(nil)
```

### Invalid JSONPath

```bash
127.0.0.1:7379> JSON.TYPE user:1001 $..name
"string"
```

### Non-JSON Data

```bash
127.0.0.1:7379> SET mykey "This is a string"
OK
127.0.0.1:7379> JSON.TYPE mykey $
(error) ERROR Existing key has wrong Dice type
```
