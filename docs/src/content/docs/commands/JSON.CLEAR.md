---
title: JSON.CLEAR
description: Documentation for the DiceDB command JSON.CLEAR
---

The `JSON.CLEAR` command is part of the DiceDBJSON module, which allows you to manipulate JSON data stored in DiceDB. This command is used to clear the value at a specified path in a JSON document, effectively setting it to an empty state. This can be particularly useful when you want to reset a part of your JSON document without removing the key itself.

## Parameters

- `key`: (String) The key under which the JSON document is stored.
- `path`: (String) The path within the JSON document that you want to clear. The path should be specified in JSONPath format. If the path is omitted, the root path (`$`) is assumed.

## Return Value

- `Integer`: The number of paths that were cleared.

## Behaviour

When the `JSON.CLEAR` command is executed, it traverses the JSON document stored at the specified key and clears the value at the given path. The clearing operation depends on the type of the value at the path:

- For objects, it removes all key-value pairs.
- For arrays, it removes all elements.
- For strings, it sets the value to an empty string.
- For numbers, it sets the value to `0`.
- For booleans, it sets the value to `false`.

If the specified path does not exist, the command does nothing and returns `0`.

## Error Handling

The `JSON.CLEAR` command can raise the following errors:

- `(error) ERR wrong number of arguments for 'json.clear' command`: This error is raised if the command is called with an incorrect number of arguments.
- `(error) ERR key does not exist`: This error is raised if the specified key does not exist in the DiceDB database.
- `(error) ERR path is not a valid JSONPath`: This error is raised if the specified path is not a valid JSONPath expression.
- `(error) ERR path does not exist`: This error is raised if the specified path does not exist within the JSON document.

## Example Usage

### Example 1: Clearing a JSON Object

Suppose you have a JSON document stored under the key `user:1001`:

```json
{
  "name": "John Doe",
  "age": 30,
  "address": {
    "street": "123 Main St",
    "city": "Anytown"
  }
}
```

To clear the `address` object, you would use the following command:

```sh
JSON.CLEAR user:1001 $.address
```

After executing this command, the JSON document would be:

```json
{
  "name": "John Doe",
  "age": 30,
  "address": {}
}
```

### Example 2: Clearing an Array

Suppose you have a JSON document stored under the key `user:1002`:

```json
{
  "name": "Jane Doe",
  "hobbies": ["reading", "swimming", "hiking"]
}
```

To clear the `hobbies` array, you would use the following command:

```sh
JSON.CLEAR user:1002 $.hobbies
```

After executing this command, the JSON document would be:

```json
{
  "name": "Jane Doe",
  "hobbies": []
}
```

### Example 3: Clearing the Root Path

Suppose you have a JSON document stored under the key `user:1003`:

```json
{
  "name": "Alice",
  "age": 25
}
```

To clear the entire JSON document, you would use the following command:

```sh
JSON.CLEAR user:1003
```

After executing this command, the JSON document would be:

```json
{}
```
