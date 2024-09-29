---
title: JSON.SET
description: The `JSON.SET` command in DiceDB is used to set the JSON value at a specified key and path. This command allows storing, updating, and querying JSON documents in DiceDB.
---

The `JSON.SET` command in DiceDB is used to set the JSON value at a specified key and path. This command allows storing, updating, and querying JSON documents in DiceDB.

## Syntax

```plaintext
JSON.SET <key> <path> <json> [NX | XX]
```

## Parameters

- `key`: The key under which the JSON document is stored. If the key does not exist, it will be created.
- `path`: The path within the JSON document where the value should be set. The path should be specified in JSONPath format. Use `$` to refer to the root of the document.
- `json`: The JSON value to set at the specified path. This should be a valid JSON string.
- `NX`: Optional flag. Only set the value if the key does not already exist.
- `XX`: Optional flag. Only set the value if the key already exists.

## Return Value

- `Simple String Reply`: Returns `OK` if the operation was successful.

### JSONPath Support

`JSON.SET` leverages the powerful JSONPath syntax to accurately pinpoint the target location for value insertion or modification. Key JSONPath concepts include:

- Root: Represented by `$`, indicates the root of the JSON document.
- Child Operators:
  - `.`: Accesses object properties.
  - `[]`: Accesses array elements.

## Behaviour

When the `JSON.SET` command is executed, the following behaviors are observed:

1. `Key Creation`: If the specified key does not exist and no flags are provided, a new key is created with the given JSON value.
2. `Path Creation`: If the specified path does not exist within the JSON document, it will be created.
3. `Conditional Set`: If the `NX` flag is provided, the value will only be set if the key does not already exist. If the `XX` flag is provided, the value will only be set if the key already exists.
4. `Overwrite`: If the key and path already exist, the existing value will be overwritten with the new JSON value.

## Error Handling

The `JSON.SET` command can raise the following errors:

1. `Syntax Error`: If the command is not used with the correct syntax, a syntax error will be raised.
   - `Error Message`: `(error) ERR wrong number of arguments for 'JSON.SET' command`
2. `Invalid JSON`: If the provided JSON value is not a valid JSON string, an error will be raised.
   - `Error Message`: `(error) ERR invalid JSON string`
3. `Path Error`: If the specified path is invalid or cannot be created, an error will be raised.
   - `Error Message`: `(error) ERR path not found`
4. `NX/XX Conflict`: If both `NX` and `XX` flags are provided, an error will be raised.
   - `Error Message`: `(error) ERR NX and XX flags are mutually exclusive`

## Example Usage

### Basic Usage

Set a JSON value at the root of a new key:

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30}'
OK
```

### Set Value at a Specific Path

Set a JSON value at a specific path within an existing JSON document:

```bash
127.0.0.1:7379> JSON.SET user:1001 $.address '{"city": "New York", "zip": "10001"}'
OK
```

### Conditional Set with NX Flag

Set a JSON value only if the key does not already exist:

```bash
127.0.0.1:7379> JSON.SET user:1002 $ '{"name": "Jane Doe", "age": 25}' NX
OK
```

### Conditional Set with XX Flag

Set a JSON value only if the key already exists:

```bash
127.0.0.1:7379> JSON.SET user:1001 $.age 31 XX
OK
```
