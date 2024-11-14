---
title: JSON.ARRPOP
description: Documentation for the DiceDB command JSON.ARRPOP
---

The `JSON.ARRPOP` command in DiceDB is used to pop an element from JSON array located at a specified path within a JSON document. This command is part of the DiceDBJSON module, which provides native JSON capabilities in DiceDB.

## Syntax

```bash
JSON.ARRPOP key [path [index]]
```

## Parameters

| Parameter | Description                                                                   | Type    | Required |
| --------- | ----------------------------------------------------------------------------- | ------- | -------- |
| `key`     | The key under which the JSON document is stored.                              | String  | Yes      |
| `path`    | The JSONPath expression that specifies the location within the JSON document. | String  | Yes      |
| `index`   | The index of the element that needs to be popped from the JSON Array at path. | Integer | No       |

## Return Value

- `string, number, object, array, boolean`: The element that is popped from the JSON Array.
- `Array`: The elements that are popped from the respective JSON Arrays.

## Behaviour

When the `JSON.ARRPOP` command is executed, the specified element is popped from the array located at the given index at the given path within the JSON document stored under the specified key. If the path does not exist or does not point to an array, an error will be raised.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

2. `Invalid Key`:

   - Error Message: `(error) ERR key does not exist`
   - Occurs when attempting to use the command on a key that does not exist.

3. `Invalid Path`:

   - Error Message: `(error) ERR path %s does not exist`
   - Occurs when attempting to use the command on a path that does not exist in the JSON document.

4. `Non Array Value at Path`:
   - Error Message: `(error) ERR path is not an array`
   - Occurs when attempting to use the command on a path that contains a non-array value.

## Example Usage

### Popping value from an array

```bash
127.0.0.1:7379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:7379> JSON.ARRPOP myjson .numbers 1
(integer) 2
127.0.0.1:7379> JSON.GET myjson
"{\"numbers\":[1,3]}"
```

### Error when key does not exist

```bash
127.0.0.1:7379> JSON.ARRPOP nonexistingkey .array 1
(error) ERR key does not exist
```

### Error when path does not exist

```bash
127.0.0.1:7379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:7379> JSON.ARRPOP myjson .nonexistingpath 4
(error) ERR path .nonexistingpath does not exist
```

### Error when path is not an array

```bash
127.0.0.1:7379> JSON.SET myjson . '{"numbers": [1, 2, 3]}'
OK
127.0.0.1:7379> JSON.ARRPOP myjson .numbers 4
(error) ERR path is not an array
```

## Notes

- Ensure that the DiceDBJSON module is loaded in your DiceDB instance to use the `JSON.ARRPOP` command.
- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.

By following this documentation, users can effectively utilize the `JSON.ARRPOP` command to manipulate JSON arrays within DiceDB.
