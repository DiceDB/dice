---
title: JSON.CLEAR
description: Documentation for the DiceDB command JSON.CLEAR
---

The `JSON.CLEAR` command allows you to manipulate JSON data stored in DiceDB. This command is used to clear the value at a specified path in a JSON document, effectively setting it to an empty state. This can be particularly useful when you want to reset a part of your JSON document without removing the key itself.

## Syntax

```bash
JSON.CLEAR key [path]
```

## Parameters

| Parameter | Description                                                                                                                                                                 | Type   | Required |
| --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | (String) The key under which the JSON document is stored.                                                                                                                   | String | Yes      |
| `path`    | (String) The path within the JSON document that you want to clear. The path should be specified in JSONPath format. If the path is omitted, the root path (`$`) is assumed. | String | No       |

## Return values

| Condition                                   | Return Value                                      |
| ------------------------------------------- | ------------------------------------------------- |
| Command is successful                       | `Integer` (The number of paths that were cleared) |
| Syntax or specified constraints are invalid | error                                             |

## Behaviour

When the `JSON.CLEAR` command is executed, it traverses the JSON document stored at the specified key and clears the value at the given path. The clearing operation depends on the type of the value at the path:

- For objects, it removes all key-value pairs.
- For arrays, it removes all elements.
- For strings, the value remaines unchanged.
- For numbers, it sets the value to `0`.
- For booleans, the value remaines unchanged.

If the specified path does not exist, the command does nothing and returns `0`.

## Errors

The `JSON.CLEAR` command can raise the following errors:

- `(error) ERR wrong number of arguments for 'json.clear' command`: This error is raised if the command is called with an incorrect number of arguments.
- `(error) ERR could not perform this operation on a key that doesn't exist`: This error is raised if the specified key does not exist in the DiceDB database.
- `(error) ERR invalid JSONPath`: This error is raised if the specified path is not a valid JSONPath expression.
- `(error) ERR Existing key has wrong Dice type`: This error is raised if the key exists but the value is not of the expected JSON type and encoding.

Note: If the specified path does not exist within the JSON document, the command will not raise an error but will simply not modify anything.

## Example Usage

### Clearing a JSON Object

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "address": {"street": "123 Main St", "city": "Anytown"}}'
OK
127.0.0.1:7379> JSON.CLEAR user:1001 $.address
(integer) 1
127.0.0.1:7379> JSON.GET user:1001
"{\"name\":\"John Doe\",\"age\":30,\"address\":{}}"
```

### Clearing a Number

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "address": {"street": "123 Main St", "city": "Anytown"}}'
OK
127.0.0.1:7379> JSON.CLEAR user:1001 $.age
(integer) 1
127.0.0.1:7379> JSON.GET user:1001
"{\"name\":\"John Doe\",\"age\":0,\"address\":{\"street\":\"123 Main St\",\"city\":\"Anytown\"}}"
```

### Clearing an Array

```bash
127.0.0.1:7379> JSON.SET user:1002 $ '{"name": "Jane Doe", "hobbies": ["reading", "swimming", "hiking"]}'
OK
127.0.0.1:7379> JSON.CLEAR user:1002 $.hobbies
(integer) 1
127.0.0.1:7379> JSON.GET user:1002
"{\"name\":\"Jane Doe\",\"hobbies\":[]}"
```

### Clearing the Root Path

```bash
127.0.0.1:7379> JSON.SET user:1003 $ '{"name": "Alice", "age": 25}'
OK
127.0.0.1:7379> JSON.CLEAR user:1003
(integer) 1
127.0.0.1:7379> JSON.GET user:1003
"{}"
```
