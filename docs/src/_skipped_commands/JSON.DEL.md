---
title: JSON.DEL
description: Documentation for the DiceDB command JSON.DEL
---

The `JSON.DEL` command is part of the DiceDBJSON module, which allows you to manipulate JSON data stored in DiceDB. This command is used to delete a specified path from a JSON document stored at a given key. If the path is not specified, the entire JSON document will be deleted.

## Syntax

```bash
JSON.DEL key [path]
```

## Parameters

| Parameter | Description                                                                                                                       | Type   | Required |
| --------- | --------------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key under which the JSON document is stored.                                                                                  | String | Yes      |
| `path`    | The JSONPath expression specifying the part of the JSON document to delete. If omitted, the entire JSON document will be deleted. | String | No       |

## Return Value

| Condition                                 | Return Value                          |
| ----------------------------------------- | ------------------------------------- |
| At least one path is successfully deleted | The number of paths deleted (Integer) |
| The specified key does not exist          | `0`                                   |

## Behaviour

When the `JSON.DEL` command is executed, it performs the following actions:

1. `Key Existence Check`: The command first checks if the specified key exists in the DiceDB database.
2. `Path Evaluation`: If a path is provided, the command evaluates the JSONPath expression to locate the part of the JSON document to delete.
3. `Deletion`: The specified path or the entire JSON document is deleted.
4. `Return`: - The command returns the number of paths that were successfully deleted. - If the specified path does not exist `0` is returned, indicating that no operation was performed.

## Error Handling

The `JSON.DEL` command can raise the following errors:

<!-- Error displayed in dicedb is different -->
<!-- - `(error) WRONGTYPE Operation against a key holding the wrong kind of value`: This error is raised if the specified key exists but does not hold a JSON document. -->

- `(error) ERROR Existing key has wrong Dice type`

<!-- - `(error) ERR Path does not exist`: This error is raised if the specified path does not exist within the JSON document. -->

## Example Usage

### Deleting an Entire JSON Document

```bash
127.0.0.1:7379> JSON.SET myjson $ '{"name": "John", "age": 30, "city": "New York"}'
OK
127.0.0.1:7379> JSON.DEL myjson
(integer) 1
127.0.0.1:7379> JSON.GET myjson
(nil)
```

### Deleting a Specific Path

```bash
127.0.0.1:7379> JSON.SET myjson $ '{"name": "John", "age": 30, "city": "New York"}'
OK
127.0.0.1:7379> JSON.DEL myjson $.age
(integer) 1
127.0.0.1:7379> JSON.GET myjson
"{\"name\":\"John\",\"city\":\"New York\"}"
```

### Deleting a Non-Existent Path

```bash
127.0.0.1:7379> JSON.SET myjson $ '{"name": "John", "age": 30, "city": "New York"}'
OK
127.0.0.1:7379> JSON.DEL myjson $.address
(integer) 0
```

### Error: Key Does Not Hold a JSON Document

```bash
127.0.0.1:7379> SET mystring "Hello, World!"
OK
127.0.0.1:7379> JSON.DEL mystring
(error) ERROR Existing key has wrong Dice type
```

## Notes

- The `JSON.DEL` command is part of the DiceDBJSON module, which must be installed and loaded into your DiceDB server.
- JSONPath expressions are used to specify the path within the JSON document. If the path is not provided, the entire document is deleted.
- The command is atomic and ensures that the deletion operation is performed consistently.

By understanding the `JSON.DEL` command, you can effectively manage and manipulate JSON data within your DiceDB database, ensuring efficient and accurate data operations.
