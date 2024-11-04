---
title: JSON.MSET
description: Documentation for the DiceDB command JSON.MSET
---

The `JSON.MSET` command in DiceDB is used to set multiple JSON values at once. This command is particularly useful when you need to update several keys with JSON data in a single atomic operation, ensuring that all updates are applied together.

## Parameters

The `JSON.MSET` command requires an even number of arguments. The arguments are provided in pairs, where each pair consists of a key and a JSON value.

- `key1`: The first key where the JSON value will be set.
- `json1`: The JSON value to be set at the first key.
- `key2`: The second key where the JSON value will be set.
- `json2`: The JSON value to be set at the second key.
- `...`: Additional key-value pairs can be provided as needed.

### Example

```bash
JSON.MSET key1 json1 key2 json2 ... keyN jsonN
```

## Return Value

The `JSON.MSET` command returns a simple string reply indicating the success of the operation.

- `OK`: If the operation is successful and all key-value pairs are set.

## Behaviour

When the `JSON.MSET` command is executed, DiceDB will:

1. Validate that the number of arguments is even.
2. Validate that each JSON value is a valid JSON string.
3. Set each key to its corresponding JSON value in an atomic operation.
4. Return `OK` if all operations are successful.

If any of the validations fail, the command will not set any of the keys, ensuring atomicity.

## Error Handling

The `JSON.MSET` command can raise errors in the following scenarios:

1. `Wrong number of arguments`: If the number of arguments is not even, DiceDB will return an error.

   - `Error Message`: `ERR wrong number of arguments for 'JSON.MSET' command`

2. `Invalid JSON`: If any of the provided JSON values are not valid JSON strings, DiceDB will return an error.

   - `Error Message`: `ERR invalid JSON string`

3. `Other DiceDB errors`: Any other standard DiceDB errors that might occur during the execution of the command.

## Example Usage

### Setting Multiple JSON Values

```bash
127.0.0.1:7379> JSON.MSET user:1 '{"name": "Alice", "age": 30}' user:2 '{"name": "Bob", "age": 25}'
OK
```

In this example, two keys (`user:1` and `user:2`) are set with their respective JSON values. The command returns `OK` indicating that both key-value pairs were successfully set.

### Error Example: Odd Number of Arguments

```bash
127.0.0.1:7379> JSON.MSET user:1 '{"name": "Alice", "age": 30}' user:2
(error) ERR wrong number of arguments for 'JSON.MSET' command
```

In this example, the command fails because the number of arguments is odd. DiceDB returns an error indicating the wrong number of arguments.

### Error Example: Invalid JSON

```bash
127.0.0.1:7379> JSON.MSET user:1 '{"name": "Alice", "age": 30}' user:2 '{name: "Bob", age: 25}'
(error) ERR invalid JSON string
```

In this example, the command fails because the JSON value for `user:2` is not a valid JSON string. DiceDB returns an error indicating the invalid JSON string.
